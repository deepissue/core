package websocket

import (
	"context"
	"net/http"

	"github.com/deepissue/core/utils"
	"github.com/gobwas/ws"
	"github.com/hashicorp/go-hclog"
)

// Server WebSocket 服务端连接
type Server struct {
	handler    *Handler
	connection *WSConnection
	ctx        context.Context
	logger     hclog.Logger
}

// UpgradeHTTP 从 HTTP 升级到 WebSocket
func UpgradeHTTP(ctx context.Context, request *http.Request,
	writer http.ResponseWriter, logger hclog.Logger) (*Server, error) {

	conn, _, _, err := ws.UpgradeHTTP(request, writer)
	if err != nil {
		return nil, err
	}

	// 创建连接包装器
	wsConn := NewWSConnection(ctx, conn, ws.StateServerSide)
	wsConn.remoteAddr = utils.GetRemoteAddr(request)

	// 创建消息处理器
	handler := NewHandler(ctx, ws.StateServerSide, logger)

	return &Server{
		handler:    handler,
		connection: wsConn,
		ctx:        ctx,
		logger:     logger,
	}, nil
}

// 消息回调代理
func (s *Server) OnPing(callback Callable)   { s.handler.OnPing(callback) }
func (s *Server) OnPong(callback Callable)   { s.handler.OnPong(callback) }
func (s *Server) OnText(callback Callable)   { s.handler.OnText(callback) }
func (s *Server) OnBinary(callback Callable) { s.handler.OnBinary(callback) }
func (s *Server) OnClose(callback Callable)  { s.handler.OnClose(callback) }

// 发送消息方法
func (s *Server) SendText(data []byte) error   { return s.connection.WriteMessage(ws.OpText, data) }
func (s *Server) SendBinary(data []byte) error { return s.connection.WriteMessage(ws.OpBinary, data) }

// HandleConnection 处理连接
func (s *Server) HandleConnection() error {
	return s.handler.HandleConnection(s.connection)
}

// RemoteAddr 获取远程地址
func (s *Server) RemoteAddr() string {
	return s.connection.RemoteAddr()
}

// Close 关闭连接
func (s *Server) Close() error {
	s.handler.Stop()
	return s.connection.Close()
}
