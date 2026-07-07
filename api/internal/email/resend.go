package email

import (
	"context"

	"github.com/resend/resend-go/v2"
)

type ResendService struct {
	client   *resend.Client
	fromAddr string
}

func NewResendService(apiKey string, fromAddr string) *ResendService {
	return &ResendService{
		client:   resend.NewClient(apiKey),
		fromAddr: fromAddr,
	}
}

func (s *ResendService) SendEmail(ctx context.Context, req *SendEmailRequest) error {
	params := &resend.SendEmailRequest{
		From:    s.fromAddr,
		To:      req.To,
		Subject: req.Subject,
		Html:    req.HTML,
		Text:    req.Text,
	}

	if len(req.Attachments) > 0 {
		var resendAttachments []*resend.Attachment
		for _, att := range req.Attachments {
			resendAttachments = append(resendAttachments, &resend.Attachment{
				Filename: att.Filename,
				Content:  att.Content,
			})
		}
		params.Attachments = resendAttachments
	}

	_, err := s.client.Emails.SendWithContext(ctx, params)
	return err
}
