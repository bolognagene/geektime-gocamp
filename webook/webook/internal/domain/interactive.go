package domain

type Interactive struct {
	LikeCnt    int64
	ReadCnt    int64
	CollectCnt int64
	// 这个是当下这个资源，你有没有点赞或者收集
	// 你也可以考虑把这两个字段分离出去，作为一个单独的结构体
	Liked     bool `json:"liked"`
	Collected bool `json:"collected"`
}

type Self struct {
	Liked     bool `json:"liked"`
	Collected bool `json:"collected"`
}

// max(发送者总速率/单一分区写入速率, 发送者总速率/单一消费者速率) + buffer
