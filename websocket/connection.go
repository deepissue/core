package websocket

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

// WSConnection 封装 WebSocket 连接
type WSConnection struct {
	net.Conn
	ctx    context.Context
	side   ws.State
	reader *wsutil.Reader

	// 写入保护
	writeMutex sync.Mutex

	// 关闭保护
	closeOnce sync.Once
	closed    chan struct{}

	remoteAddr string
}

// NewWSConnection 创建 WebSocket 连接包装器
func NewWSConnection(ctx context.Context, conn net.Conn, side ws.State) *WSConnection {
	return &WSConnection{
		Conn:       conn,
		ctx:        ctx,
		side:       side,
		reader:     wsutil.NewReader(conn, side),
		closed:     make(chan struct{}),
		remoteAddr: conn.RemoteAddr().String(),
	}
}

// ReadMessage 读取消息
func (wc *WSConnection) ReadMessage() (ws.OpCode, []byte, error) {
	hdr, err := wc.readFrameWithTimeout()
	if err != nil {
		return 0, nil, err
	}

	buffer := make([]byte, hdr.Length)
	if _, err := io.ReadFull(wc.Conn, buffer); err != nil {
		return 0, nil, err
	}

	if hdr.Masked {
		ws.Cipher(buffer, hdr.Mask, 0)
	}

	return hdr.OpCode, buffer, nil
}

// WriteMessage 写入消息
func (wc *WSConnection) WriteMessage(opCode ws.OpCode, data []byte) error {
	wc.writeMutex.Lock()
	defer wc.writeMutex.Unlock()

	select {
	case <-wc.closed:
		return net.ErrClosed
	default:
		return wsutil.WriteMessage(wc.Conn, wc.side, opCode, data)
	}
}

// Close 关闭连接
func (wc *WSConnection) Close() error {
	var err error
	wc.closeOnce.Do(func() {
		close(wc.closed)
		err = wc.Conn.Close()
	})
	return err
}

// IsClosed 检查连接是否已关闭
func (wc *WSConnection) IsClosed() bool {
	select {
	case <-wc.closed:
		return true
	default:
		return false
	}
}

// RemoteAddr 获取远程地址
func (wc *WSConnection) RemoteAddr() string {
	return wc.remoteAddr
}

func (wc *WSConnection) readFrameWithTimeout() (ws.Header, error) {
	wc.Conn.SetReadDeadline(time.Now().Add(time.Second / 2))
	defer wc.Conn.SetReadDeadline(time.Time{})
	return wc.reader.NextFrame()
}
