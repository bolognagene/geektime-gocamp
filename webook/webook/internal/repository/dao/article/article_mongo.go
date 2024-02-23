package article

import (
	"context"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type MongoArticle struct {
	client *mongo.Client
	// 代表 webook 的
	database *mongo.Database
	// 代表的是制作库
	col *mongo.Collection
	// 代表的是线上库
	liveCol *mongo.Collection
	node    *snowflake.Node
}

func NewMongoArticle(client *mongo.Client, node *snowflake.Node) ArticleDAO {
	ma := &MongoArticle{
		client:   client,
		database: client.Database("webook"),
		node:     node,
	}
	ma.col = ma.database.Collection("articles")
	ma.liveCol = ma.database.Collection("published_articles")
	InitCollections(ma.database)

	return ma
}

func (m *MongoArticle) Insert(ctx context.Context, article Article) (int64, error) {
	now := time.Now().UnixMilli()
	article.Ctime = now
	article.Utime = now

	id := m.node.Generate().Int64()
	article.Id = id

	_, err := m.col.InsertOne(ctx, article)
	// 你没有自增主键
	// GLOBAL UNIFY ID (GUID，全局唯一ID）
	return id, err
}

func (m *MongoArticle) UpdateById(ctx context.Context, article Article) error {
	now := time.Now().UnixMilli()
	article.Utime = now

	filter := bson.M{"id": article.Id, "author_id": article.AuthorId}
	update := bson.M{"$set": bson.M{
		"content": article.Content,
		"title":   article.Title,
		"utime":   article.Utime,
		"status":  article.Status,
	}}
	res, err := m.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if res.ModifiedCount == 0 {
		return fmt.Errorf("更新失败，可能是创作者非法 id %d, author_id %d",
			article.Id, article.AuthorId)
	}

	return nil
}

func (m *MongoArticle) Sync(ctx context.Context, article Article) (int64, error) {
	// 没法子引入事务的概念
	// 首先第一步，保存制作库
	var (
		id  = article.Id
		err error
	)

	if id > 0 {
		err = m.UpdateById(ctx, article)
	} else {
		id, err = m.Insert(ctx, article)
	}

	if err != nil {
		return 0, err
	}
	article.Id = id

	err = m.Upsert(ctx, PublishArticle{
		Article: article,
	})

	return id, err

}

func (m *MongoArticle) Upsert(ctx context.Context, article PublishArticle) error {
	// 操作线上库
	now := time.Now().UnixMilli()
	article.Utime = now
	filter := bson.M{"id": article.Id}
	upsert := bson.M{
		// 更新，如果不存在，就是插入，
		"$set":         article,
		"$setOnInsert": bson.M{"ctime": now},
	}

	_, err := m.liveCol.UpdateOne(ctx, filter, upsert,
		options.Update().SetUpsert(true))

	return err
}

func (m *MongoArticle) SyncStatus(ctx *gin.Context, article Article) error {
	now := time.Now().UnixMilli()
	article.Utime = now

	session, err := m.client.StartSession()
	if err != nil {
		return err
	}

	defer session.EndSession(ctx)

	// 开始事务
	err = session.StartTransaction()
	if err != nil {
		return err
	}

	filter := bson.M{"id": article.Id, "author_id": article.AuthorId}
	update := bson.M{"$set": bson.M{"status": article.Status, "utime": now}}

	res, err := m.col.UpdateOne(ctx, filter, update)
	if err != nil {
		// 回滚事务
		err2 := session.AbortTransaction(ctx)
		if err2 != nil {
			return fmt.Errorf("操作制作库失败： %v, 且事务回滚失败: %v, "+
				"article Id is: %d, status: %d", err, err2, article.Id, article.Status)
		}
		return fmt.Errorf("操作制作库失败： %v, "+
			"article Id is: %d, status: %d", err, article.Id, article.Status)
	}

	if res.ModifiedCount == 0 {
		return fmt.Errorf("操作失败，可能是创作者非法 id %d, author_id %d",
			article.Id, article.AuthorId)
	}

	res, err = m.liveCol.UpdateOne(ctx, filter, update)
	if err != nil {
		// 回滚事务
		err2 := session.AbortTransaction(ctx)
		if err2 != nil {
			return fmt.Errorf("操作线上库失败： %v, 且事务回滚失败: %v, "+
				"article Id is: %d, status: %d", err, err2, article.Id, article.Status)
		}
		return fmt.Errorf("操作线上库失败： %v, "+
			"article Id is: %d, status: %d", err, article.Id, article.Status)
	}

	err = session.CommitTransaction(ctx)
	if err != nil {
		return fmt.Errorf("提交失败导致操作失败： %v, "+
			"article Id is: %d, status: %d", err, article.Id, article.Status)
	}

	return nil
}

func InitCollections(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	index := []mongo.IndexModel{
		{
			Keys:    bson.D{bson.E{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{bson.E{Key: "author_id", Value: 1},
				bson.E{Key: "ctime", Value: 1},
			},
			Options: options.Index(),
		},
	}
	_, err := db.Collection("articles").Indexes().
		CreateMany(ctx, index)
	if err != nil {
		return err
	}
	_, err = db.Collection("published_articles").Indexes().
		CreateMany(ctx, index)
	return err
}
