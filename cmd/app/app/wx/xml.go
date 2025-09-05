package wx

import (
	"encoding/xml"
	"time"
)

// TextMessage 定义接收到的微信文本消息结构
type TextMessage struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`   // 开发者微信号（公众号的原始ID）
	FromUserName string   `xml:"FromUserName"` // 发送方帐号（用户的OpenID）
	CreateTime   int64    `xml:"CreateTime"`   // 消息创建时间（时间戳）
	MsgType      string   `xml:"MsgType"`      // 消息类型（text, image, voice, video, shortvideo, location, link, event等）
	Content      string   `xml:"Content"`      // 文本消息内容
	MsgId        int64    `xml:"MsgId"`        // 消息ID（64位整型）
}

// EventMessage 示例：关注/取消关注事件消息结构
type EventMessage struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`
	FromUserName string   `xml:"FromUserName"`
	CreateTime   int64    `xml:"CreateTime"`
	MsgType      string   `xml:"MsgType"`  // 此处为 "event"
	Event        string   `xml:"Event"`    // 事件类型（如 subscribe-关注, unsubscribe-取消关注, CLICK-点击菜单）
	EventKey     string   `xml:"EventKey"` // 事件KEY值
}

// TextResponse 定义回复给微信服务器的文本消息结构
type TextResponse struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   CDATA    `xml:"ToUserName"`   // 接收方帐号（用户的OpenID）
	FromUserName CDATA    `xml:"FromUserName"` // 发送方帐号（公众号的原始ID）
	CreateTime   int64    `xml:"CreateTime"`   // 消息创建时间（时间戳）
	MsgType      CDATA    `xml:"MsgType"`      // 消息类型（此处为 "text"）
	Content      CDATA    `xml:"Content"`      // 回复的文本内容
}

// CDATA 处理XML CDATA标签（如果需要生成回复，包含CDATA时有用）
type CDATA struct {
	Value string `xml:",cdata"`
}

// CreateTextResponse 辅助函数，用于创建文本回复消息结构体
func CreateTextResponse(toUser, fromUser, content string) TextResponse {
	return TextResponse{
		ToUserName:   CDATA{Value: toUser},
		FromUserName: CDATA{Value: fromUser},
		CreateTime:   time.Now().Unix(),
		MsgType:      CDATA{Value: "text"},
		Content:      CDATA{Value: content},
	}
}
