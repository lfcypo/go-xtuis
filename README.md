# go-xtuis

[虾推啥](https://xtuis.cn/) 的非官方 Go SDK

## 安装

```bash
go get -u github.com/lfcypo/go-xtuis
```

## 使用

```go 
package main

import "github.com/lfcypo/go-xtuis"

func main() {
    client := xtuis.NewClient("<YOUR_TOKEN>")
    payload := xtuis.NewPayload("消息标题", "消息内容")
    err := client.Send(payload)
    if err != nil {
        panic(err)
    }
}
``` 

## 功能

### 限流

根据 [虾推啥](https://xtuis.cn/) 官方说明，内置限流阈值如下：

| 限制类别    | 限制条数 |
|---------|------|
| 每日推送上限  | 300  |
| 每分钟推送上限 | 10   |

> [!WARNING]  
> 限流信息存储于内存中，这意味着每次重启应用限流信息都将会丢失。
> 我们建议您在长时间运行的服务上才依赖内置的限流策略。