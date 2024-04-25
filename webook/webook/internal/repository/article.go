package repository

import (
	"context"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/cache"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/dao"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/dao/article"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/logger"
	"github.com/ecodeclub/ekit/slice"
	"time"
)

type ArticleRepository interface {
	Create(ctx context.Context, article domain.Article) (int64, error)
	Update(ctx context.Context, article domain.Article) error
	Sync(ctx context.Context, article domain.Article) (int64, error)
	SyncStatus(ctx context.Context, article domain.Article) error
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64, uid int64) (domain.Article, error)
	GetPublishedById(ctx context.Context, id int64, uid int64) (domain.Article, error)
	ListPub(ctx context.Context, start time.Time, offset, limit int) ([]domain.Article, error)
}

type CachedArticleRepository struct {
	dao     article.ArticleDAO
	userDAO dao.UserDAO
	cache   cache.ArticleCache
	l       logger.Logger
}

func NewCachedArticleRepository(dao article.ArticleDAO, userDAO dao.UserDAO,
	cache cache.ArticleCache, l logger.Logger) ArticleRepository {
	return &CachedArticleRepository{
		dao:     dao,
		userDAO: userDAO,
		cache:   cache,
		l:       l,
	}
}

func (repo *CachedArticleRepository) Create(ctx context.Context, article domain.Article) (int64, error) {
	return repo.dao.Insert(ctx, repo.toEntity(article))
}

func (repo *CachedArticleRepository) Update(ctx context.Context, article domain.Article) error {
	return repo.dao.UpdateById(ctx, repo.toEntity(article))
}

func (repo *CachedArticleRepository) Sync(ctx context.Context, article domain.Article) (int64, error) {
	id, err := repo.dao.Sync(ctx, repo.toEntity(article))
	if err == nil {
		go func() {
			// 删除缓存
			err1 := repo.cache.DelFirstPage(ctx, article.Author.Id)
			if err1 != nil {
				repo.l.Debug("删除首页缓存失败", logger.Int64("uid", article.Author.Id),
					logger.Error(err1))
			}
			// 同时将该文章放入缓存
			err1 = repo.cache.SetPub(ctx, article, time.Minute*30)
			if err1 != nil {
				repo.l.Debug("放置读者文章缓存失败", logger.Int64("id", article.Id),
					logger.Error(err1))
			}
		}()

	}

	return id, err
}

func (repo *CachedArticleRepository) SyncStatus(ctx context.Context, article domain.Article) error {
	return repo.dao.SyncStatus(ctx, repo.toEntity(article))
}

func (repo *CachedArticleRepository) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	// 你在这个地方，集成你的复杂的缓存方案
	// 经验表明当作者打开第一页列表时，大概率会打开第一篇文章，
	// 因此在这里将第一篇文章做一个preCache, 同时到期时间设置很短，这样即使
	// 预测失败，也不会消耗太多空间
	if offset == 0 && limit <= 100 {
		data, err := repo.cache.GetFirstPage(ctx, uid)
		if err == nil {
			go func() {
				// 缓存第一篇文章
				err1 := repo.preCacheFirstArticle(ctx, data)
				if err1 != nil {
					repo.l.Debug("提前预加载缓存失败", logger.Int64("uid", uid),
						logger.Error(err1))
				}
			}()
			//return data[:limit], err
			return data, err
		}
	}
	arts, err := repo.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}
	data := slice.Map[article.Article, domain.Article](arts, func(idx int, src article.Article) domain.Article {
		return repo.toDomain(src)
	})

	// 回写缓存的时候，可以同步，也可以异步
	go func() {
		err = repo.cache.SetFirstPage(ctx, uid, data)
		if err == nil {
			err1 := repo.preCacheFirstArticle(ctx, data)
			if err1 != nil {
				repo.l.Debug("提前预加载缓存失败", logger.Int64("uid", uid), logger.Error(err1))
			}
		} else {
			repo.l.Debug("设置首页缓存失败", logger.Int64("uid", uid), logger.Error(err))
		}
	}()

	return data, nil
}

func (repo *CachedArticleRepository) GetById(ctx context.Context, id int64, uid int64) (domain.Article, error) {
	article, err := repo.cache.Get(ctx, id)
	if err == nil {
		return article, nil
	}

	data, err := repo.dao.GetById(ctx, id, uid)
	if err != nil {
		return domain.Article{}, err
	}
	return repo.toDomain(data), nil
}

func (repo *CachedArticleRepository) GetPublishedById(ctx context.Context, id int64, uid int64) (domain.Article, error) {
	article, err := repo.cache.GetPub(ctx, id)
	if err == nil {
		return article, nil
	}

	data, err := repo.dao.GetPublishedById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}

	// 这里还要得到用户的昵称
	user, err := repo.userDAO.FindById(ctx, uid)
	if err != nil {
		return domain.Article{}, err
	}
	article = repo.toDomainWithUser(data, user)

	// 回写到缓存
	go func() {
		err1 := repo.cache.SetPub(ctx, article, time.Minute*30)
		if err1 != nil {
			repo.l.Debug("GetPublishedById写入缓存失败.",
				logger.Error(err1), logger.Int64("article ID", article.Id))
		}
	}()

	return article, nil
}

func (repo *CachedArticleRepository) preCacheFirstArticle(ctx context.Context, data []domain.Article) error {
	if len(data) > 0 && len(data[0].Content) < 1024*1024 {
		err := repo.cache.Set(ctx, data[0], time.Minute)
		return err
	}
	return nil
}

func (repo *CachedArticleRepository) ListPub(ctx context.Context, start time.Time, offset, limit int) ([]domain.Article, error) {
	res, err := repo.dao.ListPub(ctx, start, offset, limit)
	if err != nil {
		return nil, err
	}

	return slice.Map[article.Article, domain.Article](res, func(idx int, src article.Article) domain.Article {
		return repo.toDomain(src)
	}), nil
}

func (repo *CachedArticleRepository) toEntity(art domain.Article) article.Article {
	return article.Article{
		Id:       art.Id,
		Content:  art.Content,
		Title:    art.Title,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	}
}

func (repo *CachedArticleRepository) toDomain(art article.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: domain.Author{
			Id: art.AuthorId,
		},
		Status: domain.ArticleStatus(art.Status),
		Utime:  time.UnixMilli(art.Utime),
		Ctime:  time.UnixMilli(art.Ctime),
	}
}

func (repo *CachedArticleRepository) toDomainWithUser(art article.Article, user dao.User) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: domain.Author{
			Id:   art.AuthorId,
			Name: user.Nickname,
		},
		Status: domain.ArticleStatus(art.Status),
		Utime:  time.UnixMilli(art.Utime),
		Ctime:  time.UnixMilli(art.Ctime),
	}
}
