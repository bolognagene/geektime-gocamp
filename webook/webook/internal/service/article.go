package service

import (
	"context"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/events/article"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/logger"
)

type ArticleService interface {
	Save(ctx context.Context, article domain.Article) (int64, error)
	Publish(ctx context.Context, article domain.Article) (int64, error)
	Withdraw(ctx context.Context, article domain.Article) error
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	Detail(ctx context.Context, id int64, uid int64) (domain.Article, error)
	PubDetail(ctx context.Context, id int64, uid int64) (domain.Article, error)
}

type articleService struct {
	repo     repository.ArticleRepository
	producer article.Producer
	l        logger.Logger
}

func NewArticleService(repo repository.ArticleRepository, producer article.Producer, l logger.Logger) ArticleService {
	return &articleService{
		repo:     repo,
		producer: producer,
		l:        l,
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

func (s *articleService) Withdraw(ctx context.Context, article domain.Article) error {
	article.Status = domain.ArticleStatusPrivate
	return s.repo.SyncStatus(ctx, article)
}

func (s *articleService) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return s.repo.List(ctx, uid, offset, limit)
}

func (s *articleService) Detail(ctx context.Context, id int64, uid int64) (domain.Article, error) {
	return s.repo.GetById(ctx, id, uid)
}

func (s *articleService) PubDetail(ctx context.Context, id int64, uid int64) (domain.Article, error) {
	// 不用kafka增加read_cnt时
	//return s.repo.GetPublishedById(ctx, id, uid)

	// 这里如果用kafka增加read_cnt时
	art, err := s.repo.GetPublishedById(ctx, id, uid)
	if err == nil {
		go func() {
			err1 := s.producer.ProduceReadEvent(ctx, article.ReadEvent{
				Uid: uid,
				Aid: id,
			})
			if err1 != nil {
				s.l.Error("发送读者阅读事件失败", logger.Error(err1),
					logger.Int64("Uid", uid), logger.Int64("Aid", id))
			}
		}()
	}

	return art, err
}
