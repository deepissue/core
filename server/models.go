package server

// Response 给客户端的返回数据结构
type Response struct {
	Code       int    `json:"code" xml:"code"`
	Message    string `json:"message" xml:"message"`
	Content    any    `json:"content" xml:"content"`
	Pagination any    `json:"pagination,omitempty" xml:"pagination,omitempty"`
	TraceID    string `json:"trace_id,omitempty" xml:"trace_id,omitempty"`
	Timestamp  int64  `json:"timestamp" xml:"timestamp"`
	Sign       string `json:"sign,omitempty" xml:"sign,omitempty"`
}
