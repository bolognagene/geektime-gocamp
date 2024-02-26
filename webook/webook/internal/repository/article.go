package repository

import (
	"context"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/dao/article"
	"github.com/ecodeclub/ekit/slice"
	"time"
)

type ArticleRepository interface {
	Create(ctx context.Context, article domain.Article) (int64, error)
	Update(ctx context.Context, article domain.Article) error
	Sync(ctx context.Context, article domain.Article) (int64, error)
	SyncStatus(ctx context.Context, article domain.Article) error
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
}

type CachedArticleRepository struct {
	dao article.ArticleDAO
}

func NewCachedArticleRepository(dao article.ArticleDAO) ArticleRepository {
	return &CachedArticleRepository{
		dao: dao,
	}
}

func (repo *CachedArticleRepository) Create(ctx context.Context, article domain.Article) (int64, error) {
	return repo.dao.Insert(ctx, repo.toEntity(article))
}

func (repo *CachedArticleRepository) Update(ctx context.Context, article domain.Article) error {
	return repo.dao.UpdateById(ctx, repo.toEntity(article))
}

func (repo *CachedArticleRepository) Sync(ctx context.Context, article domain.Article) (int64, error) {
	return repo.dao.Sync(ctx, repo.toEntity(article))
}

func (repo *CachedArticleRepository) SyncStatus(ctx context.Context, article domain.Article) error {
	return repo.dao.SyncStatus(ctx, repo.toEntity(article))
}

func (repo *CachedArticleRepository) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	arts, err := repo.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}
	data := slice.Map[article.Article, domain.Article](arts, func(idx int, src article.Article) domain.Article {
		return repo.toDomain(src)
	})
	return data, nil
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
