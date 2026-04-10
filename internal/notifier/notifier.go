package notifier

import (
	"fmt"
	"gitpigeon/internal/config"

	"github.com/wneessen/go-mail"
)

type Notifier interface {
	Notify(msg *Notification, to string) error
}

type EmailNotifier struct {
	smtpConfig *config.SMTPConfig
	client     *mail.Client
}

type Notification struct {
	RepoName    string
	TagName     string
	ReleaseName string
	ReleaseURL  string
	PublishedAt string
}

func NewEmailNotifier(smtpConfig *config.SMTPConfig) (Notifier, error) {
	opts := []mail.Option{
		mail.WithPort(smtpConfig.Port),
		mail.WithUsername(smtpConfig.User),
		mail.WithPassword(smtpConfig.Password),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
	}

	// Port 465 uses implicit TLS (SMTPS), other ports use STARTTLS
	if smtpConfig.Port == 465 {
		opts = append(opts, mail.WithSSLPort(true))
	} else {
		opts = append(opts, mail.WithTLSPortPolicy(mail.TLSMandatory))
	}

	client, err := mail.NewClient(smtpConfig.Host, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create mail client: %w", err)
	}
	return &EmailNotifier{
		smtpConfig: smtpConfig,
		client:     client,
	}, nil
}

func (n *EmailNotifier) Notify(msg *Notification, to string) error {
	message := mail.NewMsg()
	if err := message.From(n.smtpConfig.From); err != nil {
		return fmt.Errorf("failed to set from address: %w", err)
	}
	if err := message.To(to); err != nil {
		return fmt.Errorf("failed to set to address: %w", err)
	}
	message.Subject("GitPigeon: New release in " + msg.RepoName)
	body := fmt.Sprintf("New release %s available for %s!\nDetails: %s", msg.TagName, msg.RepoName, msg.ReleaseURL)
	message.SetBodyString(mail.TypeTextPlain, body)
	if err := n.client.DialAndSend(message); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}
