package service

import (
	"gitlab/nefco/auction/errors"
	"strings"

	"go.uber.org/zap"
	resty "gopkg.in/resty.v1"
)

type ConfigFeedbackService struct {
	EmailServer  string `mapstructure:"email_server"`
	EmailAddress string `mapstructure:"email_address"`
	Subject      string `mapstructure:"subject"`
}

func DefaultConfigFeedbackService() *ConfigFeedbackService {
	return &ConfigFeedbackService{
		Subject: "Auction Feedback",
	}
}

type feedbackService struct {
	cfg    *ConfigFeedbackService
	logger *zap.Logger
}

func NewFeedbackService(cfg *ConfigFeedbackService) *feedbackService {
	return &feedbackService{
		cfg:    cfg,
		logger: zap.L().Named("feedback_service"),
	}
}

func (svc *feedbackService) Send(message string) error {
	logger := svc.logger.Named("send")

	if len(strings.TrimSpace(svc.cfg.EmailServer)) == 0 ||
		len(strings.TrimSpace(svc.cfg.EmailAddress)) == 0 {
		logger.Warn("no set config",
			zap.String("email_server", svc.cfg.EmailServer),
			zap.String("email_address", svc.cfg.EmailAddress),
		)
		return nil
	}

	data := &struct {
		Email   string `json:"email"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}{
		Email:   svc.cfg.EmailAddress,
		Subject: svc.cfg.Subject,
		Message: message,
	}

	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetBody(data).
		Post(svc.cfg.EmailServer)

	if err != nil {
		logger.Error("post failed", zap.Error(err))
		return err
	}

	if resp.StatusCode() != 200 {
		logger.Error("post status code failed",
			zap.Int("code", resp.StatusCode()),
			zap.Any("response", resp),
		)
		return errors.New("post status code failed")
	}

	return nil
}
