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

	g.POST("/edit", ginx.WrapBodyAndToken[EditReq, myjwt.UserClaims](h.EditArticle, "EditArticle", h.l))
}

type EditReq struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (h *ArticleHandler) EditArticle(ctx *gin.Context, req EditReq, uc myjwt.UserClaims) (ginx.Result, error) {
	uid := uc.Uid

	id, err := h.svc.Save(ctx, domain.Article{
		Title:   req.Title,
		Content: req.Content,
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
		Msg:  "创建成功",
		Data: id,
	}, nil

}
