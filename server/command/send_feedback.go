package command

import (
	"gitlab/nefco/auction/core"

	"github.com/labstack/echo"
)

const CommandSendFeedback = "send.feedback"

type SendFeedback struct {
	*base
	FeedbackMessage string `json:"message"`
}

func newSendFeedback(user *core.User) *SendFeedback {
	return &SendFeedback{
		base: newBase(CommandSendFeedback, AccessUser, user),
	}
}

func (cmd *SendFeedback) Message() string {
	return cmd.FeedbackMessage
}

func (cmd *SendFeedback) eject(ctx echo.Context) error {
	return ctx.Bind(cmd)
}
