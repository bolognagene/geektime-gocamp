package dao

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type ArticleDAO interface {
	Insert(ctx context.Context, article Article) (int64, error)
	UpdateById(ctx context.Context, article Article) error
	Sync(ctx context.Context, article Article) (int64, error)
	Upsert(ctx context.Context, article PublishArticle) error
}

type GORMArticleDAO struct {
	db *gorm.DB
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
	res := dao.db.WithContext(ctx).Model(&article).
		Where("id = ? AND author_id = ?", article.Id, article.AuthorId).
		Updates(map[string]any{
			"title":   article.Title,
			"content": article.Content,
			"utime":   article.Utime,
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
		return txDao.Upsert(ctx, PublishArticle{
			Article: article,
		})
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
		}),
	}).Create(&article).Error
	// MySQL 最终的语句 INSERT xxx ON DUPLICATE KEY UPDATE xxx

	// 一条 SQL 语句，都不需要开启事务
	// auto commit: 意思是自动提交
	return err
}

// 事务传播机制是指如果当前有事务，就在事务内部执行 Insert
// 如果没有事务：
// 1. 开启事务，执行 Insert
// 2. 直接执行
// 3. 报错

// Article 这是制作库的
// 准备在 articles 表中准备十万/一百万条数据，author_id 各不相同（或者部分相同）
// 准备 author_id = 123 的，插入两百条数据
// 执行 SELECT * FROM articles WHERE author_id = 123 ORDER BY ctime DESC
// 比较两种索引的性能
type Article struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 长度 1024
	Title   string `gorm:"type=varchar(1024)"`
	Content string `gorm:"type=BLOB"`
	// 如何设计索引
	// 在帖子这里，什么样查询场景？
	// 对于创作者来说，是不是看草稿箱，看到所有自己的文章？
	// SELECT * FROM articles WHERE author_id = 123 ORDER BY `ctime` DESC;
	// 产品经理告诉你，要按照创建时间的倒序排序
	// 单独查询某一篇 SELECT * FROM articles WHERE id = 1
	// 在查询接口，我们深入讨论这个问题
	// - 在 author_id 和 ctime 上创建联合索引
	// - 在 author_id 上创建索引

	// 学学 Explain 命令

	// 在 author_id 上创建索引
	AuthorId int64 `gorm:"index"`
	//AuthorId int64 `gorm:"index=aid_ctime"`
	//Ctime    int64 `gorm:"index=aid_ctime"`
	Ctime int64
	Utime int64
}

// PublishArticle 这个代表的是线上表
type PublishArticle struct {
	Article
}
