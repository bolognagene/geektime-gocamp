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
	svc service.ArticleService
	l   logger.Logger
}

func NewArticleHandler(svc service.ArticleService, l logger.Logger) *ArticleHandler {
	return &ArticleHandler{
		svc: svc,
		l:   l,
	}
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")
	g.POST("/edit", ginx.WrapBodyAndToken[ArticleReq, myjwt.UserClaims](h.EditArticle, "EditArticle", h.l))
	g.POST("/publish", ginx.WrapBodyAndToken[ArticleReq, myjwt.UserClaims](h.PublishArticle, "PublishArticle", h.l))
	g.POST("/withdraw", ginx.WrapBodyAndToken[WithdrawReq, myjwt.UserClaims](h.WithdrawArticle, "WithdrawArticle", h.l))
	g.POST("/list", ginx.WrapBodyAndToken[ListReq, myjwt.UserClaims](h.ListArticle, "ListArticle", h.l))
	g.GET("/detail/:id", ginx.WrapToken[myjwt.UserClaims](h.DetailArticle, "DetailArticle", h.l))

	gpub := server.Group("/pub")
	gpub.GET("/:id", ginx.WrapToken[myjwt.UserClaims](h.DetailPubArticle, "DetailPubArticle", h.l))
}

func (h *ArticleHandler) EditArticle(ctx *gin.Context, req ArticleReq, uc myjwt.UserClaims) (ginx.Result, error) {
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

func (h *ArticleHandler) PublishArticle(ctx *gin.Context, req ArticleReq, uc myjwt.UserClaims) (ginx.Result, error) {
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

func (h *ArticleHandler) WithdrawArticle(ctx *gin.Context, req WithdrawReq, uc myjwt.UserClaims) (ginx.Result, error) {
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

func (h *ArticleHandler) ListArticle(ctx *gin.Context, req ListReq, uc myjwt.UserClaims) (ginx.Result, error) {
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

func (h *ArticleHandler) DetailArticle(ctx *gin.Context, uc myjwt.UserClaims) (ginx.Result, error) {
	uid := uc.Uid
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		return ginx.Result{
			Code: 4,
			Msg:  "参数错误",
		}, fmt.Errorf("前端输入id错误，%v", err)
	}

	h.svc.Detail(ctx, id, uid)
}

func (h *ArticleHandler) DetailPubArticle(ctx *gin.Context, uc myjwt.UserClaims) (ginx.Result, error) {

}
