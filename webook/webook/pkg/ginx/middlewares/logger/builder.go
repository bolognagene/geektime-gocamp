package logger

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"hash"
	"io"
	"strings"
	"time"
)

type Builder struct {
	allowReqBody  bool
	allowRespBody bool
	allowEncrypt  bool
	countLimit    int
	h             hash.Hash
	loggerFunc    func(ctx context.Context, al *AccessLog)
}

func NewBuilder(loggerFunc func(ctx context.Context, al *AccessLog)) *Builder {
	return &Builder{
		loggerFunc: loggerFunc,
		h:          sha1.New(),
	}
}

func (b *Builder) AllowReqBody() *Builder {
	b.allowReqBody = true
	return b
}

func (b *Builder) AllowRespBody() *Builder {
	b.allowRespBody = true
	return b
}

func (b *Builder) AllowEncrypt() *Builder {
	b.allowEncrypt = true
	return b
}

func (b *Builder) CountLimit(limit int) *Builder {
	b.countLimit = limit
	return b
}

func (b *Builder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		method := ctx.Request.Method
		url := ctx.Request.URL.String()
		//最多记录limit的数据
		if len(url) > b.countLimit {
			url = url[:b.countLimit]
		}
		al := &AccessLog{
			Method: method,
			Url:    url,
		}
		if b.allowReqBody && ctx.Request.Body != nil {
			// Body 读完就没有了
			//reqBody := ctx.Request.Body
			body, _ := ctx.GetRawData()
			reader := io.NopCloser(bytes.NewReader(body))
			ctx.Request.Body = reader
			if len(body) > b.countLimit {
				body = body[:b.countLimit]
			}
			// 这其实是一个很消耗 CPU 和内存的操作
			// 因为会引起复制

			// 这里进行脱敏
			if b.allowEncrypt {
				var reqJson map[string]interface{}
				_ = json.Unmarshal(body, &reqJson)
				for key, value := range reqJson {
					if strings.EqualFold(key, "phone") ||
						strings.EqualFold(key, "email") ||
						strings.Contains(strings.ToLower(key), "password") {
						b.h.Write([]byte(value.(string)))
						reqJson[key] = hex.EncodeToString(b.h.Sum(nil))
					}
				}
				body, _ = json.Marshal(reqJson)
			}

			al.ReqBody = string(body)
		}

		if b.allowRespBody {
			ctx.Writer = responseWriter{
				al:             al,
				ResponseWriter: ctx.Writer,
			}
		}

		defer func() {
			al.Duration = time.Since(start).String()
			if al.RespBody != "" && len(al.RespBody) > b.countLimit {
				al.RespBody = al.RespBody[:b.countLimit]
			}
			b.loggerFunc(ctx, al)
		}()

		ctx.Next()

	}
}

type responseWriter struct {
	al *AccessLog
	gin.ResponseWriter
}

func (rw responseWriter) WriteHeader(statusCode int) {
	rw.al.Status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw responseWriter) Write(data []byte) (int, error) {
	rw.al.RespBody = string(data)
	return rw.ResponseWriter.Write(data)
}

func (rw responseWriter) WriteString(s string) (int, error) {
	rw.al.RespBody = s
	return rw.ResponseWriter.WriteString(s)
}

type AccessLog struct {
	// HTTP 请求的方法
	Method string
	// Url 整个请求 URL
	Url      string
	Duration string
	ReqBody  string
	RespBody string
	Status   int
}