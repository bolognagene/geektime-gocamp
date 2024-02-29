package web

import (
	"fmt"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service"
	myjwt "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/web/jwt"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/ginx"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/logger"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

var _ handler = (*ArticleHandler)(nil)

type ArticleHandler struct {
	svc      service.ArticleService
	interSvc service.InteractiveService
	l        logger.Logger
	biz      string
}

func NewArticleHandler(svc service.ArticleService, interSvc service.InteractiveService, l logger.Logger) *ArticleHandler {
	return &ArticleHandler{
		svc:      svc,
		interSvc: interSvc,
		l:        l,
		biz:      "article",
	}
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")
	g.POST("/edit", ginx.WrapBodyAndToken[ArticleReq, myjwt.UserClaims](h.Edit, "EditArticle", h.l))
	g.POST("/publish", ginx.WrapBodyAndToken[ArticleReq, myjwt.UserClaims](h.Publish, "PublishArticle", h.l))
	g.POST("/withdraw", ginx.WrapBodyAndToken[WithdrawReq, myjwt.UserClaims](h.Withdraw, "WithdrawArticle", h.l))
	g.POST("/list", ginx.WrapBodyAndToken[ListReq, myjwt.UserClaims](h.List, "ListArticle", h.l))
	g.GET("/detail/:id", ginx.WrapToken[myjwt.UserClaims](h.Detail, "DetailArticle", h.l))

	gpub := server.Group("/pub")
	gpub.GET("/:id", ginx.WrapToken[myjwt.UserClaims](h.PubDetail, "DetailPubArticle", h.l))
	// 点赞是这个接口，取消点赞也是这个接口
	// RESTful 风格
	//gpub.GET("/like/:id", ginx.WrapToken[myjwt.UserClaims](h.PubDetail, "DetailPubArticle", h.l))
	gpub.POST("/like", ginx.WrapBodyAndToken[LikeReq, myjwt.UserClaims](h.Like, "LikeArticle", h.l))
}

func (h *ArticleHandler) Edit(ctx *gin.Context, req ArticleReq, uc myjwt.UserClaims) (ginx.Result, error) {
	uid := uc.Uid

	id, err := h.svc.Save(ctx, req.toDomain(uid))

	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}

	return ginx.Result{
		Code: 2,
		Msg:  "创建成功",
		Data: id,
	}, nil

}

func (h *ArticleHandler) Publish(ctx *gin.Context, req ArticleReq, uc myjwt.UserClaims) (ginx.Result, error) {
	uid := uc.Uid

	id, err := h.svc.Publish(ctx, req.toDomain(uid))

	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}

	return ginx.Result{
		Code: 2,
		Msg:  "发表成功",
		Data: id,
	}, nil

}

func (h *ArticleHandler) Withdraw(ctx *gin.Context, req WithdrawReq, uc myjwt.UserClaims) (ginx.Result, error) {
	uid := uc.Uid

	err := h.svc.Withdraw(ctx, domain.Article{
		Id: req.Id,
		Author: domain.Author{
			Id: uid,
		},
	})

	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}

	return ginx.Result{
		Code: 2,
		Msg:  "withdraw成功",
		Data: req.Id,
	}, nil

}

func (h *ArticleHandler) List(ctx *gin.Context, req ListReq, uc myjwt.UserClaims) (ginx.Result, error) {
	uid := uc.Uid

	arts, err := h.svc.List(ctx, uid, req.Offset, req.Limit)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	// 在列表页，不显示全文，只显示一个"摘要"
	// 比如说，简单的摘要就是前几句话
	// 强大的摘要是 AI 帮你生成的
	return ginx.Result{
		Code: 2,
		Data: slice.Map[domain.Article, ArticleVO](arts,
			func(idx int, src domain.Article) ArticleVO {
				return ArticleVO{
					Id:       src.Id,
					Title:    src.Title,
					Abstract: src.Abstract(),
					Status:   src.Status.ToUint8(),
					// 这个列表请求，不需要返回内容
					//Content: src.Content,
					// 这个是创作者看自己的文章列表，也不需要这个字段
					//Author: src.Author
					Ctime: src.Ctime.Format(time.DateTime),
					Utime: src.Utime.Format(time.DateTime),
				}
			}),
	}, nil

}

func (h *ArticleHandler) Detail(ctx *gin.Context, uc myjwt.UserClaims) (ginx.Result, error) {
	uid := uc.Uid
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		return ginx.Result{
			Code: 4,
			Msg:  "参数错误",
		}, fmt.Errorf("前端输入id错误，%v", err)
	}

	article, err := h.svc.Detail(ctx, id, uid)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}

	return ginx.Result{
		Code: 2,
		Data: ArticleVO{
			Id:       article.Id,
			Title:    article.Title,
			Content:  article.Content,
			Abstract: article.Abstract(),
			Status:   article.Status.ToUint8(),
			Utime:    article.Utime.Format(time.DateTime),
			Ctime:    article.Ctime.Format(time.DateTime),
		},
	}, nil

}

func (h *ArticleHandler) PubDetail(ctx *gin.Context, uc myjwt.UserClaims) (ginx.Result, error) {
	uid := uc.Uid
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		return ginx.Result{
			Code: 4,
			Msg:  "参数错误",
		}, fmt.Errorf("前端输入id错误，%v", err)
	}

	article, err := h.svc.PubDetail(ctx, id, uid)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}

	// 增加阅读计数。
	go func() {
		err1 := h.interSvc.IncrReadCnt(ctx, h.biz, article.Id)
		if err1 != nil {
			h.l.Error("增加阅读计数失败",
				logger.Int64("aid", article.Id),
				logger.Error(err))
		}

	}()

	return ginx.Result{
		Code: 2,
		Data: ArticleVO{
			Id:       article.Id,
			Title:    article.Title,
			Content:  article.Content,
			Abstract: article.Abstract(),
			Status:   article.Status.ToUint8(),
			Author:   article.Author.Name,
			Utime:    article.Utime.Format(time.DateTime),
			Ctime:    article.Ctime.Format(time.DateTime),
		},
	}, nil
}

func (h *ArticleHandler) Like(ctx *gin.Context, req LikeReq, uc myjwt.UserClaims) (ginx.Result, error) {
	uid := uc.Uid
	var err error

	if req.Like {
		err = h.interSvc.Like(ctx, h.biz, req.Id, uid)
	} else {
		err = h.interSvc.Unlike(ctx, h.biz, req.Id, uid)
	}
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}

	return ginx.Result{
		Code: 2,
		Msg:  "OK",
	}, nil

}
