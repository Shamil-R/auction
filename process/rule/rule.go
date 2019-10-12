package rule

import (
	"encoding/json"
	"gitlab/nefco/auction/core"
	"gitlab/nefco/auction/errors"
	"gitlab/nefco/auction/process"
	"net/http"
	"sort"
	"time"

	"go.uber.org/zap"
	validator "gopkg.in/go-playground/validator.v9"
)

var (
	validationFailed = errors.NewError("Validation failed", http.StatusUnprocessableEntity)
)

func ValidateRule(validate *validator.Validate) {
	validate.RegisterValidation("rule",
		func(fl validator.FieldLevel) bool {
			rule := fl.Field().String()
			return rule == Normal || rule == Extra || rule == Pick
		},
	)
}

func Rules(configs []*core.RuleConfig, defaultConf *Config) ([]process.Rule, error) {
	logger := zap.L().Named("rules")

	rules := make([]process.Rule, len(configs))

	validate := validator.New()

	for i, conf := range configs {
		interval, err := process.ParseInterval(conf.Start, conf.Duration.Duration())
		if err != nil {
			return nil, err
		}

		var rule process.Rule

		switch conf.Type {
		case Normal:
			rule = newNormal(interval, *defaultConf.Normal)
		case Extra:
			rule = newExtra(interval, *defaultConf.Extra)
		case Pick:
			rule = newPick(interval, *defaultConf.Pick)
		}

		if err := json.Unmarshal(conf.Props, rule); err != nil {
			return nil, err
		}

		if err := validate.Struct(rule.Config()); err != nil {
			logger.Error("validation failed", zap.Error(err))
			return nil, validationFailed
		}

		rules[i] = rule
	}

	return fillWait(rules)
}

func fillWait(in []process.Rule) ([]process.Rule, error) {
	sort.Slice(in, func(i, j int) bool {
		a := in[i].Interval().Start()
		b := in[j].Interval().Start()
		return a.Before(b)
	})

	out := make([]process.Rule, 0, len(in))

	d := 24 * time.Hour

	fill, err := process.ParseInterval("00:00:00", d)
	if err != nil {
		return nil, err
	}

	for _, r := range in {
		if r.Interval().Start().Sub(fill.Start()) > 0 {
			ri, err := process.NewInterval(fill.Start(), r.Interval().Start().Sub(fill.Start()))
			if err != nil {
				return nil, err
			}

			out = append(out, newWait(ri), r)

			d := fill.Duration() - r.Interval().Duration() - ri.Duration()
			if d < 0 {
				d = 0
			}

			fill, err = process.NewInterval(r.Interval().End(), d)
			if err != nil {
				return nil, err
			}
		} else {
			out = append(out, r)

			d := fill.Duration() - r.Interval().Duration()
			if d < 0 {
				d = 0
			}

			fill, err = process.NewInterval(r.Interval().End(), d)
			if err != nil {
				return nil, err
			}
		}
	}

	if fill.Duration() > 0 {
		out = append(out, newWait(fill))
	}

	// for _, r := range out {
	// 	zap.L().Debug("out rule",
	// 		zap.String("rule", r.Rule()),
	// 		zap.Time("start", r.Interval().Start()),
	// 		zap.Time("end", r.Interval().End()),
	// 		zap.Duration("duration", r.Interval().Duration()),
	// 	)
	// }

	return out, nil
}

func now() time.Time {
	return time.Now().UTC()
}
