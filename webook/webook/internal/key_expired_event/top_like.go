package key_expired_event

import (
	"context"
	"fmt"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/interacitve/repository"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/web"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/logger"
)

type TopLikeKey struct {
	repo repository.InteractiveRepository
	l    logger.Logger
	biz  string
}

func NewTopLikeKey(repo repository.InteractiveRepository,
	l logger.Logger, biz string) *TopLikeKey {
	return &TopLikeKey{
		repo: repo,
		l:    l,
		biz:  biz,
	}
}

func (t *TopLikeKey) Process(key string) error {
	// 是过期的key就处理
	if key == fmt.Sprintf("top_like_%s", t.biz) {
		_, err := t.repo.GetTopLike(context.Background(), t.biz,
			web.TopLikeN.Load(), web.TopLikeLimit.Load())
		if err != nil {
			t.l.Debug("top like key到期处理失败", logger.Error(err),
				logger.String("biz", t.biz))
		}
	}

	return nil
}
