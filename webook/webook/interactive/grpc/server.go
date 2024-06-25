package grpc

import (
	"context"
	intrv1 "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/api/proto/gen/intr/v1"
	domain2 "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/interactive/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/interactive/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// InteractiveServiceServer 我这里只是把 service 包装成一个 grpc 而已
// 和 grpc 有关的操作，就限定在这里
type InteractiveServiceServer struct {
	intrv1.UnimplementedInteractiveServiceServer
	svc service.InteractiveService
}

func (i *InteractiveServiceServer) IncrReadCnt(ctx context.Context, request *intrv1.IncrReadCntRequest) (*intrv1.IncrReadCntResponse, error) {
	err := i.svc.IncrReadCnt(ctx, request.GetBiz(), request.GetBizId())
	return &intrv1.IncrReadCntResponse{}, err
}

func (i *InteractiveServiceServer) Like(ctx context.Context, request *intrv1.LikeRequest) (*intrv1.LikeResponse, error) {
	err := i.svc.Like(ctx, request.GetBiz(), request.GetBizId(), request.GetUid(), request.GetLimit())
	return &intrv1.LikeResponse{}, err
}

func (i *InteractiveServiceServer) Unlike(ctx context.Context, request *intrv1.UnlikeRequest) (*intrv1.UnlikeResponse, error) {
	// 也可以考虑用 grpc 的插件
	if request.Uid <= 0 {
		return nil, status.Error(codes.InvalidArgument, "uid 错误")
	}
	err := i.svc.Unlike(ctx, request.GetBiz(), request.GetBizId(), request.GetUid(), request.GetLimit())
	return &intrv1.UnlikeResponse{}, err
}

func (i *InteractiveServiceServer) AddCollect(ctx context.Context, request *intrv1.AddCollectRequest) (*intrv1.AddCollectResponse, error) {
	err := i.svc.AddCollect(ctx, request.GetBiz(), request.GetBizId(), request.GetCid(), request.GetUid())
	return &intrv1.AddCollectResponse{}, err
}

func (i *InteractiveServiceServer) DeleteCollect(ctx context.Context, request *intrv1.DeleteCollectRequest) (*intrv1.DeleteCollectResponse, error) {
	err := i.svc.DeleteCollect(ctx, request.GetBiz(), request.GetBizId(), request.GetCid(), request.GetUid())
	return &intrv1.DeleteCollectResponse{}, err
}

func (i *InteractiveServiceServer) Get(ctx context.Context, request *intrv1.GetRequest) (*intrv1.GetResponse, error) {
	res, err := i.svc.Get(ctx, request.GetBiz(), request.GetBizId(), request.GetUid())
	if err != nil {
		return &intrv1.GetResponse{}, err
	}
	return &intrv1.GetResponse{
		Intr: i.toDTO(res),
	}, nil
}

func (i *InteractiveServiceServer) TopLike(ctx context.Context, request *intrv1.TopLikeRequest) (*intrv1.TopLikeResponse, error) {
	res, err := i.svc.TopLike(ctx, request.GetBiz(), request.GetN(), request.GetLimit())
	if err != nil {
		return &intrv1.TopLikeResponse{}, nil
	}
	toplikes := make([]*intrv1.TopWithScore, len(res))
	for i, top := range res {
		toplikes[i] = &(intrv1.TopWithScore{
			Score:  float32(top.Score),
			Member: top.Member,
		})
	}

	return &intrv1.TopLikeResponse{
		TopWithScores: toplikes,
	}, err
}

func (i *InteractiveServiceServer) GetByIds(ctx context.Context, request *intrv1.GetByIdsRequest) (*intrv1.GetByIdsResponse, error) {
	res, err := i.svc.GetByIds(ctx, request.GetBiz(), request.GetBizIds())
	intrs := make(map[int64]*intrv1.Interactive, len(res))
	if err != nil {
		return &intrv1.GetByIdsResponse{}, err
	}
	for k, v := range res {
		intrs[k] = i.toDTO(v)
	}

	return &intrv1.GetByIdsResponse{
		Intrs: intrs,
	}, nil
}

func (i *InteractiveServiceServer) mustEmbedUnimplementedInteractiveServiceServer() {
	//TODO implement me
	panic("implement me")
}

// DTO data transfer object
func (i *InteractiveServiceServer) toDTO(intr domain2.Interactive) *intrv1.Interactive {
	return &intrv1.Interactive{
		Biz:        intr.Biz,
		BizId:      intr.BizId,
		CollectCnt: intr.CollectCnt,
		Collected:  intr.Collected,
		LikeCnt:    intr.LikeCnt,
		Liked:      intr.Liked,
		ReadCnt:    intr.ReadCnt,
	}
}
