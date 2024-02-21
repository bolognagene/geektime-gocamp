package web

import (
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service"
	myjwt "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/web/jwt"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/ginx"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/logger"
	"github.com/gin-gonic/gin"
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
}

type ArticleReq struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (req ArticleReq) toDomain(uid int64) domain.Article {
	return domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uid,
		},
	}
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
		Msg:  "创建成功",
		Data: id,
	}, nil

}
