package email

import "context"

type Attachment struct {
	Filename string
	Content  []byte
}

type SendEmailRequest struct {
	To          []string
	Subject     string
	HTML        string
	Text        string
	Attachments []Attachment
}

type Service interface {
	SendEmail(ctx context.Context, req *SendEmailRequest) error
}
