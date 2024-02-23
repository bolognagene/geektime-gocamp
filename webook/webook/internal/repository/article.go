package repository

import (
	"context"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/dao/article"
	"github.com/gin-gonic/gin"
)

type ArticleRepository interface {
	Create(ctx context.Context, article domain.Article) (int64, error)
	Update(ctx context.Context, article domain.Article) error
	Sync(ctx context.Context, article domain.Article) (int64, error)
	SyncStatus(ctx *gin.Context, article domain.Article) error
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

func (repo *CachedArticleRepository) SyncStatus(ctx *gin.Context, article domain.Article) error {
	return repo.dao.SyncStatus(ctx, repo.toEntity(article))
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
