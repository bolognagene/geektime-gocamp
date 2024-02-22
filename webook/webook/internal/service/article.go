package service

import (
	"context"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository"
	"github.com/gin-gonic/gin"
)

type ArticleService interface {
	Save(ctx context.Context, article domain.Article) (int64, error)
	Publish(ctx context.Context, article domain.Article) (int64, error)
	Withdraw(ctx *gin.Context, article domain.Article) error
}

type articleService struct {
	repo repository.ArticleRepository
}

func NewArticleService(repo repository.ArticleRepository) ArticleService {
	return &articleService{
		repo: repo,
	}
}

func (s *articleService) Save(ctx context.Context, article domain.Article) (int64, error) {
	article.Status = domain.ArticleStatusUnpublished
	// 如何article的Id大于0， 证明该文章已经有了，所以是Update，否则是Create
	if article.Id > 0 {
		err := s.repo.Update(ctx, article)
		return article.Id, err
	}
	return s.repo.Create(ctx, article)
}

func (s *articleService) Publish(ctx context.Context, article domain.Article) (int64, error) {
	article.Status = domain.ArticleStatusPublished
	return s.repo.Sync(ctx, article)
}

func (s *articleService) Withdraw(ctx *gin.Context, article domain.Article) error {
	article.Status = domain.ArticleStatusPrivate
	return s.repo.SyncStatus(ctx, article)
}
