package xtuis

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lfcypo/go-xtuis/internal/limiter"
)

const (
	// DefaultServerURL 默认推送服务器地址
	DefaultServerURL = "https://wx.xtuis.cn"
)

// Client 推送客户端
type Client struct {
	serverURL string
	token     string

	endpoint   string
	httpClient *http.Client
}

// ClientOption 推送客户端配置
type ClientOption struct {
	// ServerURL 推送服务器地址
	ServerURL string

	// Timeout 超时时间
	Timeout time.Duration
}

// NewClient 创建推送客户端
func NewClient(token string, options ...ClientOption) *Client {
	endpoint := DefaultServerURL
	timeout := 5 * time.Second
	if len(options) > 0 {
		endpoint = strings.TrimRight(options[0].ServerURL, "/")
		timeout = options[0].Timeout
	}

	url := fmt.Sprintf("%s/%s.send", endpoint, token)

	return &Client{
		serverURL: endpoint,
		token:     token,
		endpoint:  url,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Send 发送消息
func (c *Client) Send(payload *Payload) error {
	if payload.err != nil {
		return payload.err
	}

	if !limiter.DayLimiter.Limit() {
		return errors.New("you have reached the limit of sending messages today")
	}
	if !limiter.MinuteLimiter.Limit() {
		return errors.New("you have reached the limit of sending messages this minute")
	}

	return c.send(payload)
}

// send 发送消息
func (c *Client) send(payload *Payload) error {
	formData := payload.toFormValues()
	formBytes := []byte(formData.Encode())

	request, err := http.NewRequestWithContext(
		context.Background(),
		"POST",
		c.endpoint,
		bytes.NewBuffer(formBytes),
	)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("server returned status code " + resp.Status)
	}

	return nil
}
