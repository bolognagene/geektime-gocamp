package repository

import (
	"context"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/dao"
	"github.com/gin-gonic/gin"
)

type ArticleRepository interface {
	Create(ctx context.Context, article domain.Article) (int64, error)
	Update(ctx context.Context, article domain.Article) error
	Sync(ctx context.Context, article domain.Article) (int64, error)
	SyncStatus(ctx *gin.Context, article domain.Article) error
}

type CachedArticleRepository struct {
	dao dao.ArticleDAO
}

func NewCachedArticleRepository(dao dao.ArticleDAO) ArticleRepository {
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
	return repo.dao.SyncStatus(ctx, article)
}

func (repo *CachedArticleRepository) toEntity(article domain.Article) dao.Article {
	return dao.Article{
		Id:       article.Id,
		Content:  article.Content,
		Title:    article.Title,
		AuthorId: article.Author.Id,
		Status:   article.Status.ToUint8(),
	}
}
