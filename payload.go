package xtuis

import (
	"errors"
	"net/url"
)

// Payload 消息体
type Payload struct {
	// Text 标题
	Text string

	// Description 消息内容 最大 64 KB
	Description string

	err error
}

// NewPayload 创建消息体
func NewPayload(text, description string) *Payload {
	var err error
	if text == "" || description == "" {
		err = errors.New("text and description cannot be empty")
	}
	if len(text) > 64 || len(description) > 64*1024 {
		err = errors.New("text and description cannot be longer than 64 KB")
	}
	return &Payload{
		Text:        text,
		Description: description,

		err: err,
	}
}

// toFormValues 转换为表单数据
func (p *Payload) toFormValues() url.Values {
	formData := url.Values{}
	formData.Set("text", p.Text)
	formData.Set("desp", p.Description)
	return formData
}
