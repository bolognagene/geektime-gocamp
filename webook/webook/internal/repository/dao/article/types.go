package article

import (
	"context"
)

type ArticleDAO interface {
	Insert(ctx context.Context, article Article) (int64, error)
	UpdateById(ctx context.Context, article Article) error
	Sync(ctx context.Context, article Article) (int64, error)
	Upsert(ctx context.Context, article PublishArticle) error
	SyncStatus(ctx context.Context, article Article) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error)
}

// Article 这是制作库的
// 准备在 articles 表中准备十万/一百万条数据，author_id 各不相同（或者部分相同）
// 准备 author_id = 123 的，插入两百条数据
// 执行 SELECT * FROM articles WHERE author_id = 123 ORDER BY ctime DESC
// 比较两种索引的性能
type Article struct {
	Id int64 `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	// 长度 1024
	Title   string `gorm:"type=varchar(1024)" bson:"title,omitempty"`
	Content string `gorm:"type=BLOB" bson:"content,omitempty"`
	// 如何设计索引
	// 在帖子这里，什么样查询场景？
	// 对于创作者来说，是不是看草稿箱，看到所有自己的文章？
	// SELECT * FROM articles WHERE author_id = 123 ORDER BY `ctime` DESC;
	// 产品经理告诉你，要按照创建时间的倒序排序
	// 单独查询某一篇 SELECT * FROM articles WHERE id = 1
	// 在查询接口，我们深入讨论这个问题
	// - 在 author_id 和 ctime 上创建联合索引
	// - 在 author_id 上创建索引

	// 学学 Explain 命令

	// 在 author_id 上创建索引
	AuthorId int64 `gorm:"index" bson:"author_id,omitempty"`
	Status   uint8 `bson:"status,omitempty"`
	//AuthorId int64 `gorm:"index=aid_ctime"`
	//Ctime    int64 `gorm:"index=aid_ctime"`
	Ctime int64 `bson:"ctime,omitempty"`
	Utime int64 `bson:"utime,omitempty"`
}

// PublishArticle 这个代表的是线上表
/*type PublishArticle struct {
	Article
}*/

type PublishArticle Article

type PublishedArticleV1 struct {
	Id       int64
	Title    string
	AuthorId int64
	Status   uint8
	Ctime    int64
	Utime    int64
}
