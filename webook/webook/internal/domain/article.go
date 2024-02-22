package domain

type Article struct {
	Id      int64
	Title   string
	Content string
	Author  Author
	Status  ArticleStatus
}

type Author struct {
	Id   int64
	Name string
}

type ArticleStatus uint8

const (
	ArticleStatusUnknown ArticleStatus = iota
	ArticleStatusUnpublished
	ArticleStatusPublished
	ArticleStatusPrivate
)

func (as ArticleStatus) ToUint8() uint8 {
	return uint8(as)
}

func (as ArticleStatus) NonPublished() bool {
	return as != ArticleStatusPublished
}

func (as ArticleStatus) String() string {
	switch as {
	case ArticleStatusPrivate:
		return "Private"
	case ArticleStatusPublished:
		return "Published"
	case ArticleStatusUnpublished:
		return "Unpublished"
	default:
		return "Unknown"
	}
}

func (as ArticleStatus) IsValid() bool {
	return as != ArticleStatusUnknown
}
