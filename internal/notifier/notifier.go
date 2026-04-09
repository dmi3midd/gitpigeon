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
	client, err := mail.NewClient(smtpConfig.Host, mail.WithPort(smtpConfig.Port), mail.WithUsername(smtpConfig.User), mail.WithPassword(smtpConfig.Password))
	if err != nil {
		return nil, err
	}
	return &EmailNotifier{
		smtpConfig: smtpConfig,
		client:     client,
	}, nil
}

func (n *EmailNotifier) Notify(msg *Notification, to string) error {
	message := mail.NewMsg()
	if err := message.From(n.smtpConfig.From); err != nil {
		return fmt.Errorf("failed to set from address: %v", err)
	}
	if err := message.To(to); err != nil {
		return fmt.Errorf("failed to set to address: %v", err)
	}
	message.Subject("GitPigeon: New release in " + msg.RepoName)
	body := fmt.Sprintf("New release %s available for %s!\nDetails: %s", msg.TagName, msg.RepoName, msg.ReleaseURL)
	message.SetBodyString(mail.TypeTextPlain, body)
	if err := n.client.DialAndSend(message); err != nil {
		return err
	}
	return nil
}
