package service

import (
	"context"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository"
	"github.com/ecodeclub/ekit/queue"
	"github.com/ecodeclub/ekit/slice"
	"math"
	"time"
)

type RankingService interface {
	TopN(ctx context.Context) error
}

type BatchRankingService struct {
	artSvc    ArticleService
	intrSvc   InteractiveService
	repo      repository.RankingRepository
	batchSize int
	n         int
	biz       string
	// scoreFunc 不能返回负数
	scoreFunc func(t time.Time, likeCnt int64) float64
}

func NewBatchRankingService(artSvc ArticleService,
	intrSvc InteractiveService,
	repo repository.RankingRepository) RankingService {
	return &BatchRankingService{
		artSvc:    artSvc,
		intrSvc:   intrSvc,
		repo:      repo,
		n:         100,
		batchSize: 100,
		biz:       "article",
		scoreFunc: func(t time.Time, likeCnt int64) float64 {
			sec := time.Since(t).Seconds()
			return float64(likeCnt-1) / math.Pow(float64(sec+2), 1.5)
		},
	}
}

func (svc *BatchRankingService) TopN(ctx context.Context) error {
	arts, err := svc.topN(ctx)
	if err != nil {
		return err
	}
	// 在这里，存起来
	return svc.repo.ReplaceTopN(ctx, arts)

}

func (svc *BatchRankingService) topN(ctx context.Context) ([]domain.Article, error) {
	// 我只取七天内的数据
	now := time.Now()
	// 先拿一批数据
	offset := 0
	type ArticleWithScore struct {
		article domain.Article
		score   float64
	}
	// 这里可以用非并发安全
	topN := queue.NewConcurrentPriorityQueue[ArticleWithScore](svc.n,
		func(src ArticleWithScore, dst ArticleWithScore) int {
			if src.score > dst.score {
				return 1
			} else if src.score == dst.score {
				return 0
			} else {
				return -1
			}
		})

	for {
		// 这里拿了一批
		arts, err := svc.artSvc.ListPub(ctx, now, offset, svc.batchSize)
		if err != nil {
			return nil, err
		}
		ids := slice.Map[domain.Article, int64](arts, func(idx int, src domain.Article) int64 {
			return src.Id
		})
		// 要去找到对应的点赞数据
		intrs, err := svc.intrSvc.GetByIds(ctx, svc.biz, ids)
		if err != nil {
			return nil, err
		}

		// 合并计算 score
		// 排序
		for _, art := range arts {
			intr := intrs[art.Id]
			//if !ok {
			//	// 你都没有，肯定不可能是热榜
			//	continue
			//}
			score := svc.scoreFunc(art.Utime, intr.LikeCnt)
			// 我要考虑，我这个 score 在不在前一百名
			// 拿到热度最低的
			err = topN.Enqueue(ArticleWithScore{
				article: art,
				score:   score,
			})
			// 这种写法，要求 topN 已经满了
			if err == queue.ErrOutOfCapacity {
				val, _ := topN.Dequeue()
				if val.score < score {
					err = topN.Enqueue(ArticleWithScore{
						article: art,
						score:   score,
					})

					if err != nil {
						topN.Enqueue(val)
					}
				} else {
					topN.Enqueue(val)
				}
			}
		}

		// 一批已经处理完了，问题来了，我要不要进入下一批？我怎么知道还有没有？
		if len(arts) < svc.batchSize ||
			now.Sub(arts[len(arts)-1].Utime).Hours() > 7*24 {
			// 我这一批都没取够，我当然可以肯定没有下一批了
			// 又或者已经取到了七天之前的数据了，说明可以中断了
			break
		}
		// 这边要更新 offset
		offset += len(arts)
	}
	// 最后得出结果
	res := make([]domain.Article, svc.n)
	// 热度从低到高
	for i := svc.n - 1; i >= 0; i-- {
		val, err := topN.Dequeue()
		if err != nil {
			// 说明取完了，不够 n
			break
		}
		res[i] = val.article
	}
	return res, nil

}
