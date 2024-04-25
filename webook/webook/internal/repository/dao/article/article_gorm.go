package article

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type GORMArticleDAO struct {
	db *gorm.DB
}

func (dao *GORMArticleDAO) GetById(ctx context.Context, id int64, uid int64) (Article, error) {
	var article Article
	err := dao.db.WithContext(ctx).Model(&Article{}).
		Where("id = ? AND author_id = ?", id, uid).
		Find(&article).Error
	if err != nil {
		return Article{}, err
	}

	return article, nil
}

func NewGORMArticleDAO(db *gorm.DB) ArticleDAO {
	return &GORMArticleDAO{
		db: db,
	}
}

func (dao *GORMArticleDAO) Insert(ctx context.Context, article Article) (int64, error) {
	now := time.Now().UnixMilli()
	article.Ctime = now
	article.Utime = now
	err := dao.db.WithContext(ctx).Create(&article).Error
	return article.Id, err
}

func (dao *GORMArticleDAO) UpdateById(ctx context.Context, article Article) error {
	now := time.Now().UnixMilli()
	article.Utime = now
	// 依赖 gorm 忽略零值的特性，会用主键进行更新
	//dao.db.WithContext(ctx).Save(&article)
	// 可读性很差
	// 这里需要加上文章作者和传入的uid一致，以防止修改其他人的文章
	res := dao.db.WithContext(ctx).Model(&Article{}).
		Where("id = ? AND author_id = ?", article.Id, article.AuthorId).
		Updates(map[string]any{
			"title":   article.Title,
			"content": article.Content,
			"utime":   article.Utime,
			"status":  article.Status,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		//dangerousDBOp.Count(1)
		// 补充一点日志
		return fmt.Errorf("更新失败，可能是创作者非法 id %d, author_id %d",
			article.Id, article.AuthorId)

	}

	return res.Error
}

func (dao *GORMArticleDAO) Sync(ctx context.Context, article Article) (int64, error) {
	// 先操作制作库（此时应该是表），后操作线上库（此时应该是表）
	var id = article.Id

	// tx => Transaction, trx, txn
	// 在事务内部，这里采用了闭包形态
	// GORM 帮助我们管理了事务的生命周期
	// Begin，Rollback 和 Commit 都不需要我们操心
	err := dao.db.Transaction(func(tx *gorm.DB) error {
		var err error
		txDao := NewGORMArticleDAO(tx)
		if id > 0 {
			err = txDao.UpdateById(ctx, article)
		} else {
			id, err = txDao.Insert(ctx, article)
		}
		if err != nil {
			return err
		}

		// 操作线上库了
		return txDao.Upsert(ctx, PublishArticle(article))
	})

	return id, err
}

// Upsert INSERT OR UPDATE
func (dao *GORMArticleDAO) Upsert(ctx context.Context, article PublishArticle) error {
	now := time.Now().UnixMilli()
	article.Ctime = now
	article.Utime = now
	// 这个是插入
	// OnConflict 的意思是数据冲突了
	err := dao.db.Clauses(clause.OnConflict{
		// SQL 2003 标准
		// INSERT AAAA ON CONFLICT(BBB) DO NOTHING
		// INSERT AAAA ON CONFLICT(BBB) DO UPDATES CCC WHERE DDD

		// 哪些列冲突
		//Columns: []clause.Column{clause.Column{Name: "id"}},
		// 意思是数据冲突，啥也不干
		// DoNothing:
		// 数据冲突了，并且符合 WHERE 条件的就会执行 DO UPDATES
		// Where:

		// MySQL 只需要关心这里
		DoUpdates: clause.Assignments(map[string]any{
			"title":   article.Title,
			"content": article.Content,
			"utime":   article.Utime,
			"status":  article.Status,
		}),
	}).Create(&article).Error
	// MySQL 最终的语句 INSERT xxx ON DUPLICATE KEY UPDATE xxx

	// 一条 SQL 语句，都不需要开启事务
	// auto commit: 意思是自动提交
	return err
}

func (dao *GORMArticleDAO) SyncStatus(ctx context.Context, article Article) error {
	now := time.Now().UnixMilli()
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).
			Where("id = ? AND author_id = ?", article.Id, article.AuthorId).
			Updates(map[string]any{
				"status": article.Status,
				"utime":  now,
			})

		if res.Error != nil {
			return res.Error
		}

		if res.RowsAffected == 0 {
			// 要么 ID 是错的，要么作者不对
			// 后者情况下，你就要小心，可能有人在搞你的系统
			// 没必要再用 ID 搜索数据库来区分这两种情况
			// 用 prometheus 打点，只要频繁出现，你就告警，然后手工介入排查
			return fmt.Errorf("可能有人在搞你，误操作非自己的文章, uid: %d, aid: %d",
				article.AuthorId, article.Id)

		}

		return tx.Model(&PublishArticle{}).
			Where("id = ?", article.Id).
			Updates(map[string]any{
				"status": article.Status,
				"utime":  now,
			}).Error
	})
	return err
}

func (dao *GORMArticleDAO) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error) {
	var arts []Article
	err := dao.db.WithContext(ctx).Model(&Article{}).
		Where("author_id=?", uid).Offset(offset).Limit(limit).
		Order("utime DESC").
		//Order(clause.OrderBy{Columns: []clause.OrderByColumn{
		//	{Column: clause.Column{Name: "utime"}, Desc: true},
		//	{Column: clause.Column{Name: "ctime"}, Desc: false},
		//}}).
		Find(&arts).Error

	return arts, err
}

func (dao *GORMArticleDAO) GetPublishedById(ctx context.Context, id int64) (Article, error) {
	var article Article
	err := dao.db.WithContext(ctx).Model(&PublishArticle{}).
		Where("id = ?", id).
		Find(&article).Error
	if err != nil {
		return Article{}, err
	}

	return article, nil
}

func (dao *GORMArticleDAO) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]Article, error) {
	var arts []Article
	err := dao.db.WithContext(ctx).Model(&Article{}).
		Where("utime < ?", start.UnixMilli()).Offset(offset).Limit(limit).
		Order("utime DESC").
		//Order(clause.OrderBy{Columns: []clause.OrderByColumn{
		//	{Column: clause.Column{Name: "utime"}, Desc: true},
		//	{Column: clause.Column{Name: "ctime"}, Desc: false},
		//}}).
		Find(&arts).Error

	return arts, err
}

// 事务传播机制是指如果当前有事务，就在事务内部执行 Insert
// 如果没有事务：
// 1. 开启事务，执行 Insert
// 2. 直接执行
// 3. 报错
