package notification

import (
	"fmt"

	"github.com/slack-go/slack"
	"go.uber.org/zap"
)

const (
	DeclaredIconURL = "https://img.icons8.com/pastel-glyph/64/26e07f/point-to-beginning.png"
	SuccessIconURL  = "https://img.icons8.com/ios-glyphs/60/26e07f/rocket.png"
	FailedIconURL   = "https://img.icons8.com/external-tal-revivo-color-tal-revivo/96/fa314a/external-fatal-error-notification-in-computer-operating-system-basic-color-tal-revivo.png"
)

type SlackConfig struct {
	*TmpUpdate
	*slack.WebhookMessage
}

func SlackInit(tmp *TmpUpdate, msg string) *SlackConfig {
	return &SlackConfig{tmp, &slack.WebhookMessage{
		Channel: tmp.SlackChannel,
		Text:    msg,
	}}
}

func (s *SlackConfig) SlackDeclareNotify() error {
	s.Username = s.Command.HelpName + ": declared"
	s.IconURL = DeclaredIconURL
	return s.slackPostMsg("declared")
}

func (s *SlackConfig) SlackSuccessNotify() error {
	s.Username = s.Command.HelpName + ": success"
	s.IconURL = SuccessIconURL
	return s.slackPostMsg("success")
}

func (s *SlackConfig) SlackFailNotify() error {
	s.Username = s.Command.HelpName + ": failed"
	s.IconURL = FailedIconURL
	return s.slackPostMsg("failed")
}

func (s *SlackConfig) slackPostMsg(status string) error {
	if !s.SlackNotifications {
		return nil
	} else {
		if len(s.SlackWebHook) == 0 || len(s.SlackChannel) == 0 {
			zap.S().Fatalf("parameters --slack-webhook, --slack-channel not set for 'rmk config init' command," +
				"required if Slack notifications are enabled")
		}
	}

	zap.S().Infof("sending message %s to Slack channel: %s", status, s.Channel)
	if err := slack.PostWebhook(s.SlackWebHook, s.WebhookMessage); err != nil {
		return fmt.Errorf("failed to notify to Slack: %w", err)
	}

	return nil
}
