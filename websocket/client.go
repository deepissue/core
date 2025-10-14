package websocket

import (
	"context"
	"crypto/tls"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/google/uuid"
	"github.com/hashicorp/go-hclog"
)

// Client WebSocket 客户端
type Client struct {
	// 配置
	addr            string
	header          ws.HandshakeHeader
	tlsConfig       *tls.Config
	reconnectConfig ReconnectConfig

	// 组件
	handler    *Handler
	connection *WSConnection

	// 状态管理
	ctx        context.Context
	state      ConnectionState
	stateMutex sync.RWMutex
	sessionId  string

	// 重连控制
	reconnectCount int
	stopReconnect  chan struct{}

	// 回调
	onOpen          func()
	onReconnect     func(attempt int)
	onReconnectFail func(attempt int, err error)
	onDisconnect    func(reason string)

	logger hclog.Logger
}

// NewClient 创建客户端
func NewClient(ctx context.Context, addr string, header ws.HandshakeHeader, logger hclog.Logger) *Client {
	// 创建消息处理器
	return &Client{
		addr:            addr,
		header:          header,
		tlsConfig:       &tls.Config{InsecureSkipVerify: true},
		reconnectConfig: DefaultReconnectConfig,
		ctx:             ctx,
		state:           StateDisconnected,
		stopReconnect:   make(chan struct{}),
		logger:          logger,
		handler:         NewHandler(ctx, ws.StateClientSide, logger),
	}
}

// SetReconnectConfig 设置重连配置
func (c *Client) SetReconnectConfig(config ReconnectConfig) {
	c.reconnectConfig = config
}

// 设置回调函数
func (c *Client) OnOpen(callback func())                    { c.onOpen = callback }
func (c *Client) OnReconnect(callback func(int))            { c.onReconnect = callback }
func (c *Client) OnReconnectFail(callback func(int, error)) { c.onReconnectFail = callback }
func (c *Client) OnDisconnect(callback func(string))        { c.onDisconnect = callback }

// 消息回调代理
func (c *Client) OnPing(callback Callable) {
	if c.handler != nil {
		c.handler.OnPing(callback)
	}
}

func (c *Client) OnPong(callback Callable) {
	if c.handler != nil {
		c.handler.OnPong(callback)
	}
}

func (c *Client) OnText(callback Callable) {
	if c.handler != nil {
		c.handler.OnText(callback)
	}
}

func (c *Client) OnBinary(callback Callable) {
	if c.handler != nil {
		c.handler.OnBinary(callback)
	}
}

func (c *Client) OnClose(callback Callable) {
	if c.handler != nil {
		c.handler.OnClose(callback)
	}
}

// Connect 连接服务器
func (c *Client) Connect() error {
	return c.connect()
}

// WaitConnected 等待连接建立
func (c *Client) WaitConnected(timeout time.Duration) error {
	if c.GetState() == StateConnected {
		return nil
	}

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	timeoutTimer := time.NewTimer(timeout)
	defer timeoutTimer.Stop()

	for {
		select {
		case <-ticker.C:
			if c.GetState() == StateConnected {
				return nil
			}
		case <-timeoutTimer.C:
			return fmt.Errorf("timeout waiting for agent connection")
		case <-c.ctx.Done():
			return nil
		}
	}
}

// ConnectWithReconnect 连接服务器并支持自动重连
func (c *Client) ConnectWithReconnect() error {
	for {
		select {
		case <-c.ctx.Done():
			return c.ctx.Err()
		case <-c.stopReconnect:
			return nil
		default:
			if err := c.connect(); err != nil {
				if !c.reconnectConfig.Enable {
					return err
				}

				c.reconnectCount++
				if c.reconnectConfig.MaxRetries > 0 &&
					c.reconnectCount > c.reconnectConfig.MaxRetries {
					return fmt.Errorf("max reconnect attempts reached: %d", c.reconnectCount)
				}

				if c.onReconnectFail != nil {
					c.onReconnectFail(c.reconnectCount, err)
				}

				interval := c.calculateBackoffInterval()
				c.logger.Warn("connection failed, retrying",
					"attempt", c.reconnectCount, "interval", interval, "error", err)

				c.setState(StateReconnecting)

				select {
				case <-time.After(interval):
					continue
				case <-c.ctx.Done():
					return c.ctx.Err()
				case <-c.stopReconnect:
					return nil
				}
			}

			// 连接成功，重置重连计数
			c.reconnectCount = 0
			if c.onReconnect != nil {
				c.onReconnect(c.reconnectCount)
			}

			// 处理连接
			c.handleConnection()
		}
	}
}

// connect 建立连接
func (c *Client) connect() error {
	c.setState(StateConnecting)
	c.logger.Info("connecting", "addr", c.addr)
	dialer := ws.Dialer{
		TLSConfig: c.tlsConfig,
		Header:    c.header,
	}

	conn, _, _, err := dialer.Dial(c.ctx, c.addr)
	if err != nil {
		c.setState(StateDisconnected)
		return err
	}

	// 创建连接包装器
	c.connection = NewWSConnection(c.ctx, conn, ws.StateClientSide)

	c.setState(StateConnected)
	c.setSession()

	if c.onOpen != nil {
		c.onOpen()
	}
	c.logger.Info("connected", "addr", c.addr)
	return nil
}

// handleConnection 处理连接
func (c *Client) handleConnection() {
	if err := c.handler.HandleConnection(c.connection); err != nil {
		c.logger.Error("handle connection failed", "error", err)
		if c.onDisconnect != nil {
			c.onDisconnect(err.Error())
		}
	}

	c.setState(StateDisconnected)

	// 清理资源
	if c.connection != nil {
		c.connection.Close()
		c.connection = nil
	}
}

// 发送消息方法
func (c *Client) SendText(data []byte) error {
	if c.connection == nil || c.connection.IsClosed() {
		return fmt.Errorf("connection not available")
	}
	return c.connection.WriteMessage(ws.OpText, data)
}

func (c *Client) SendBinary(data []byte) error {
	if c.connection == nil || c.connection.IsClosed() {
		return fmt.Errorf("connection not available")
	}
	return c.connection.WriteMessage(ws.OpBinary, data)
}

func (c *Client) IsConnected() bool {
	c.stateMutex.RLock()
	defer c.stateMutex.RUnlock()
	return c.GetState() == StateConnected
}

// GetState 获取连接状态
func (c *Client) GetState() ConnectionState {
	c.stateMutex.RLock()
	defer c.stateMutex.RUnlock()
	return c.state
}

func (c *Client) setState(state ConnectionState) {
	c.stateMutex.Lock()
	defer c.stateMutex.Unlock()
	c.state = state
}

func (c *Client) setSession() {
	c.stateMutex.Lock()
	defer c.stateMutex.Unlock()
	c.sessionId = strings.ReplaceAll(uuid.New().String(), "-", "")
}

func (c *Client) SessionId() string {
	c.stateMutex.Lock()
	defer c.stateMutex.Unlock()
	return c.sessionId
}

// Close 关闭客户端
func (c *Client) Close() error {
	close(c.stopReconnect)

	if c.handler != nil {
		c.handler.Stop()
	}

	if c.connection != nil {
		return c.connection.Close()
	}

	return nil
}

// calculateBackoffInterval 计算退避间隔
func (c *Client) calculateBackoffInterval() time.Duration {
	var interval time.Duration

	switch c.reconnectConfig.BackoffStrategy {
	case FixedBackoff:
		interval = c.reconnectConfig.InitialInterval

	case ExponentialBackoff:
		interval = time.Duration(float64(c.reconnectConfig.InitialInterval) *
			math.Pow(2, float64(c.reconnectCount-1)))

	case LinearBackoff:
		interval = c.reconnectConfig.InitialInterval * time.Duration(c.reconnectCount)

	default:
		interval = c.reconnectConfig.InitialInterval
	}

	// 添加抖动
	if c.reconnectConfig.Jitter {
		jitterRange := float64(interval) * 0.25
		jitter := time.Duration(rand.Float64()*jitterRange*2 - jitterRange)
		interval += jitter

		// 确保不小于最小间隔
		if interval < time.Second {
			interval = time.Second
		}
	}

	// 限制最大间隔
	if interval > c.reconnectConfig.MaxInterval {
		interval = c.reconnectConfig.MaxInterval
	}

	return interval
}
