package services

import "github.com/dghubble/go-twitter/twitter"

type QuickReplyBuilder struct {
	options []twitter.DirectMessageQuickReplyOption
}

func NewQRBuilder() *QuickReplyBuilder {
	return &QuickReplyBuilder{[]twitter.DirectMessageQuickReplyOption{}}
}

func (qr *QuickReplyBuilder) CustomQuickRepy(qrs ...twitter.DirectMessageQuickReplyOption) *QuickReplyBuilder {
	qr.options = append(qr.options, qrs...)
	return qr
}

func (qr *QuickReplyBuilder) StartButton() *QuickReplyBuilder {
	qr.options = append(qr.options, twitter.DirectMessageQuickReplyOption{
		Label:       "/start",
		Description: "Start convo 🚀🚀🚀",
		Metadata:    "external_id_1",
	})
	return qr
}
func (qr *QuickReplyBuilder) StopButton() *QuickReplyBuilder {
	qr.options = append(qr.options, twitter.DirectMessageQuickReplyOption{
		Label:       "/stop",
		Description: "Stop convo 🛑❌🛑❌🛑❌",
		Metadata:    "external_id_2",
	})
	return qr
}

func (qr *QuickReplyBuilder) GetQuickReply() *twitter.DirectMessageQuickReply {
	return &twitter.DirectMessageQuickReply{
		Type:    "options",
		Options: qr.options,
	}
}
