package dao

import (
	dao2 "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/interacitve/repository/dao"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/dao/article"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/cronJobScheduler/repository/dao"
	"gorm.io/gorm"
)

func InitTable(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &article.Article{}, &article.PublishArticle{},
		&dao2.Interactive{}, &dao2.UserLikeBiz{}, &dao2.Collection{}, &dao2.UserCollectionBiz{}, &dao.Job{})
}
