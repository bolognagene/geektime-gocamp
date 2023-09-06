package web

// Code 0:  成功
// Code 1:
// Code 2:
// Code 3:
// Code 4:  业务错误
// Code 5:  系统错误

type Result struct {
	// 这个叫做业务错误码
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}
