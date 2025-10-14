package websocket

import (
	"time"
)

// 连接状态
type ConnectionState int

const (
	StateDisconnected ConnectionState = iota
	StateConnecting
	StateConnected
	StateReconnecting
	StateStopped
)

// 退避策略
type BackoffStrategy int

const (
	FixedBackoff BackoffStrategy = iota
	ExponentialBackoff
	LinearBackoff
)

// 重连配置
type ReconnectConfig struct {
	Enable          bool
	InitialInterval time.Duration
	MaxInterval     time.Duration
	MaxRetries      int // 0 表示无限重连
	BackoffStrategy BackoffStrategy
	Jitter          bool // 是否添加随机抖动
}

// 回调函数类型
type Callable func(conn *WSConnection, buffer []byte)

// 默认重连配置
var DefaultReconnectConfig = ReconnectConfig{
	Enable:          true,
	InitialInterval: 1 * time.Second,
	MaxInterval:     60 * time.Second,
	MaxRetries:      0,
	BackoffStrategy: ExponentialBackoff,
	Jitter:          true,
}
