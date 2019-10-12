package json_fileds

import (
	"gopkg.in/go-playground/validator.v9"
	"time"
)

// lots.complete json field
type Complete struct {
	ActNumber         string     `json:"act_number" validate:"required"`
	ActDate           time.Time  `json:"act_date" validate:"required"`
	SendDate          time.Time  `json:"send_date" validate:"required"`
	ReceiveDate       *time.Time `json:"receive_date,omitempty"`
	ReceiveRepeatDate *time.Time `json:"receive_repeat_date,omitempty"`
	RegistryDate      *time.Time `json:"registry_date,omitempty"`
	PaymentDate       *time.Time `json:"payment_date,omitempty"`
	RiseInPrice       *int       `json:"rise_in_price,omitempty"`
	Comment           *string    `json:"comment,omitempty"`
	AllowChange       *int       `json:"allow_change"`
}

func (c *Complete) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}