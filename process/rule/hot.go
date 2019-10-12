package rule

import (
	"gitlab/nefco/auction/process"
)

const Hot = "hot"

type hot struct {
	*common
	*ExtraConfig
}

func newHot(config *ExtraConfig) (*hot, error) {
	n := now()

	interval, err := process.NewInterval(n, config.HotDuration)
	if err != nil {
		return nil, err
	}

	return &hot{newCommon(Hot, interval, config), config}, nil
}

func (rule *hot) CancelBet(act process.Action, prx process.RuleProxy) error {
	return cancelBetDisabled
}
