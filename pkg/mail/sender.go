package mail

import (
	"fmt"
	"net/smtp"

	"github.com/No2004LTC/gopher-social-ecom/config"
)

// EmailSender interface giúp Clean Architecture (dễ dàng thay thế bằng SendGrid/AWS sau này)
type EmailSender interface {
	SendEmail(to string, subject string, content string) error
}

type gmailSender struct {
	name              string
	fromEmailAddress  string
	fromEmailPassword string
	smtpServerAddress string
	smtpServerPort    string
}

// NewGmailSender khởi tạo một bộ gửi thư bằng Gmail
func NewGmailSender(cfg *config.Config) EmailSender {
	return &gmailSender{
		name:              "Gopher Social System",
		fromEmailAddress:  cfg.SenderEmail,
		fromEmailPassword: cfg.SMTPPassword,
		smtpServerAddress: cfg.SMTPHost,
		smtpServerPort:    cfg.SMTPPort,
	}
}

func (sender *gmailSender) SendEmail(to string, subject string, content string) error {
	// Định dạng lại thư để tránh lỗi font tiếng Việt và hiển thị đẹp hơn
	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-version: 1.0;\r\n"+
		"Content-Type: text/html; charset=\"UTF-8\";\r\n\r\n"+
		"%s", to, subject, content))

	auth := smtp.PlainAuth("", sender.fromEmailAddress, sender.fromEmailPassword, sender.smtpServerAddress)
	smtpAddr := fmt.Sprintf("%s:%s", sender.smtpServerAddress, sender.smtpServerPort)

	return smtp.SendMail(smtpAddr, auth, sender.fromEmailAddress, []string{to}, msg)
}
