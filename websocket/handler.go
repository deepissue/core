package websocket

import (
	"context"
	"encoding/binary"
	"errors"
	"os"
	"time"

	"github.com/gobwas/ws"
	"github.com/hashicorp/go-hclog"
)

var sides = map[ws.State]string{
	ws.StateClientSide: "client",
	ws.StateServerSide: "server",
}

// Handler 通用消息处理器
type Handler struct {
	ctx    context.Context
	side   ws.State
	logger hclog.Logger

	// 消息回调
	onPing   Callable
	onPong   Callable
	onText   Callable
	onBinary Callable
	onClose  Callable

	// ping/pong 管理
	pingInterval time.Duration
	pongTimeout  time.Duration

	// 停止信号
	stopChan chan struct{}
}

// NewHandler 创建消息处理器
func NewHandler(ctx context.Context, side ws.State, logger hclog.Logger) *Handler {
	return &Handler{
		ctx:          ctx,
		side:         side,
		logger:       logger,
		stopChan:     make(chan struct{}),
		pingInterval: time.Second * 30,
		pongTimeout:  time.Second * 60,
	}
}

// 设置回调函数
func (h *Handler) OnPing(callback Callable)   { h.onPing = callback }
func (h *Handler) OnPong(callback Callable)   { h.onPong = callback }
func (h *Handler) OnText(callback Callable)   { h.onText = callback }
func (h *Handler) OnBinary(callback Callable) { h.onBinary = callback }
func (h *Handler) OnClose(callback Callable)  { h.onClose = callback }

// SetPingInterval 设置 ping 间隔
func (h *Handler) SetPingInterval(interval time.Duration) {
	h.pingInterval = interval
	h.pongTimeout = interval * 2
}

// HandleConnection 处理连接生命周期
func (h *Handler) HandleConnection(conn *WSConnection) error {
	defer func() {
		h.logger.Trace("stop handling connection", "side", sides[h.side], "remote", conn.RemoteAddr())
		if r := recover(); r != nil {
			h.logger.Error("panic recovered in HandleConnection", sides[h.side], "remote", conn.RemoteAddr(), "error", r)
		}
	}()
	h.logger.Trace("start handling connection", "side", sides[h.side], "remote", conn.RemoteAddr())

	// 启动 ping 循环
	if h.side == ws.StateClientSide {
		go h.pingLoop(conn)
	}

	// 消息处理循环
	return h.messageLoop(conn)
}

// Stop 停止处理
func (h *Handler) Stop() {
	select {
	case <-h.stopChan:
		return // 已经停止
	default:
		close(h.stopChan)
	}
}

// messageLoop 消息处理循环
func (h *Handler) messageLoop(conn *WSConnection) error {
	for {
		select {
		case <-h.ctx.Done():
			return h.ctx.Err()
		case <-h.stopChan:
			return nil
		default:
			opCode, data, err := conn.ReadMessage()
			if err != nil {
				if errors.Is(err, os.ErrDeadlineExceeded) {
					continue
				}
				h.logger.Error("read message failed", "side", sides[h.side],
					"remote", conn.RemoteAddr(), "error", err)
				return err
			}

			if err := h.handleMessage(conn, opCode, data); err != nil {
				return err
			}
		}
	}
}

// handleMessage 处理单个消息
func (h *Handler) handleMessage(conn *WSConnection, opCode ws.OpCode, data []byte) error {
	h.logger.Trace("received message", "side", sides[h.side],
		"remote", conn.RemoteAddr(), "opCode", opCode, "length", len(data))

	switch opCode {
	case ws.OpClose:
		code, reason := parseCloseFrame(data)
		h.logger.Info("connection closed", "side", sides[h.side],
			"remote", conn.RemoteAddr(), "code", code, "reason", reason)
		if h.onClose != nil {
			h.onClose(conn, data)
		}
		return errors.New("connection closed")

	case ws.OpPing:
		h.logger.Trace("received ping", "side", sides[h.side], "remote", conn.RemoteAddr())
		if h.onPing != nil {
			h.onPing(conn, data)
		}
		// 自动回复 pong
		return conn.WriteMessage(ws.OpPong, data)

	case ws.OpPong:
		h.logger.Trace("received pong", "side", sides[h.side], "remote", conn.RemoteAddr())
		if h.onPong != nil {
			h.onPong(conn, data)
		}

	case ws.OpText:
		if h.logger.IsTrace() {
			h.logger.Trace("received text message", "side", sides[h.side], "remote", conn.RemoteAddr(), "text", string(data))
		} else {
			h.logger.Trace("received text message", "side", sides[h.side], "remote", conn.RemoteAddr(), "text", len(data))
		}
		if h.onText != nil {
			h.onText(conn, data)
		}

	case ws.OpBinary:
		if h.logger.IsTrace() {
			h.logger.Trace("received binary message", "side", sides[h.side], "remote", conn.RemoteAddr(), "text", string(data))
		} else {
			h.logger.Trace("received binary message", "side", sides[h.side], "remote", conn.RemoteAddr(), "text", len(data))
		}
		if h.onBinary != nil {
			h.onBinary(conn, data)
		}
	}

	return nil
}

// pingLoop ping 循环（仅客户端）
func (h *Handler) pingLoop(conn *WSConnection) {
	ticker := time.NewTicker(h.pingInterval)
	defer ticker.Stop()

	var ticketCounter uint64 = 1

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-h.stopChan:
			return
		case <-ticker.C:
			// 发送 ping
			buf := make([]byte, 8)
			binary.PutUvarint(buf, ticketCounter)
			if err := conn.WriteMessage(ws.OpPing, buf); err != nil {
				h.logger.Error("send ping failed", "error", err)
				return
			}
			ticketCounter++
		}
	}
}

// parseCloseFrame 解析关闭帧
func parseCloseFrame(payload []byte) (ws.StatusCode, string) {
	if len(payload) < 2 {
		return ws.StatusNormalClosure, ""
	}
	code := binary.BigEndian.Uint16(payload[:2])
	reason := string(payload[2:])
	return ws.StatusCode(code), reason
}
