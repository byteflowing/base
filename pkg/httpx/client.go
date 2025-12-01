package httpx

import (
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	_defaultConfig *Config
	_defaultClient *http.Client

	defaultClientOnce sync.Once
)

func init() {
	_defaultConfig = &Config{
		Timeout:               30 * time.Second,
		MaxIdleConns:          200,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DialTimeout:           5 * time.Second,
		KeepAlive:             30 * time.Second,
	}
}

type Config struct {
	Timeout               time.Duration // 请求整体超时
	MaxIdleConns          int           // 全局最大空闲连接数
	MaxIdleConnsPerHost   int           // 每个主机最大空闲连接数
	IdleConnTimeout       time.Duration // 空闲连接存活时间
	TLSHandshakeTimeout   time.Duration // TLS 握手超时
	ExpectContinueTimeout time.Duration // Expect: 100-continue 超时
	DialTimeout           time.Duration // 建立 TCP 连接超时
	KeepAlive             time.Duration // TCP keep-alive 时间
}

// NewClient 根据配置生成一个 http.Client
func NewClient(cfg *Config) *http.Client {
	transport := &http.Transport{
		MaxIdleConns:          cfg.MaxIdleConns,
		MaxIdleConnsPerHost:   cfg.MaxIdleConnsPerHost,
		IdleConnTimeout:       cfg.IdleConnTimeout,
		TLSHandshakeTimeout:   cfg.TLSHandshakeTimeout,
		ExpectContinueTimeout: cfg.ExpectContinueTimeout,
		DialContext: (&net.Dialer{
			Timeout:   cfg.DialTimeout,
			KeepAlive: cfg.KeepAlive,
		}).DialContext,
	}

	return &http.Client{
		Timeout:   cfg.Timeout,
		Transport: transport,
	}
}

func GetDefaultConfig() *Config {
	return _defaultConfig
}

// Default 返回一个带默认配置的 http.Client
// 单例模式，若要创建一个新的实例，使用NewClient,GetDefaultConfig可以获取默认配置
func Default() *http.Client {
	defaultClientOnce.Do(func() {
		_defaultClient = NewClient(_defaultConfig)
	})
	return _defaultClient
}
