package service

import (
	"encoding/json"
	"gitlab/nefco/auction/core"
	"gitlab/nefco/auction/core/object"
	"gitlab/nefco/auction/errors"
	"gitlab/nefco/auction/interfaces"
	lib "gitlab/nefco/auction/lib"
	"time"

	"go.uber.org/zap"
	resty "gopkg.in/resty.v1"
)

type ConfigBackService struct {
	Managers map[string]string `mapstructure:"managers"`
}

func DefaultConfigBackService() *ConfigBackService {
	return &ConfigBackService{
		Managers: map[string]string{},
	}
}

type backService struct {
	cfg      *ConfigBackService
	executor *core.User
	lot      *core.Lot
	logger   *zap.Logger
}

func NewBackService(
	cfg *ConfigBackService, executor *core.User, lot *core.Lot) *backService {
	return &backService{
		cfg:      cfg,
		lot:      lot,
		executor: executor,
		logger: zap.L().Named("back_service").With(
			zap.Uint("executor_id", executor.ID),
			zap.Uint("lot_id", lot.ID),
		),
	}
}

func (svc backService) PostConfirmation(info object.JSONData) error {
	data := struct {
		ObjectID uint            `json:"objectID"`
		Info     object.JSONData `json:"info"`
		UserID   uint            `json:"userID"`
	}{
		svc.lot.ObjectID,
		info,
		svc.executor.ID,
	}

	return svc.post("PostConfirmation", data)
}

func (svc backService) DeleteConfirmation() error {
	data := struct {
		ObjectID uint `json:"objectID"`
		UserID   uint `json:"userID"`
	}{
		svc.lot.ObjectID,
		svc.executor.ID,
	}

	return svc.post("DeleteConfirmation", data)
}

func (svc backService) SetDateBook(t time.Time, bet *core.Bet) error {
	data := struct {
		ObjectID   uint          `json:"objectID"`
		DateBooked core.JSONTime `json:"date_booked"`
		UserID     uint          `json:"userID"`
		*core.Bet
	}{
		svc.lot.ObjectID,
		core.JSONTime(t),
		svc.executor.ID,
		bet,
	}

	return svc.post("SetDateBook", data)
}

func (svc backService) ResetDateBook() error {
	data := struct {
		ObjectID uint `json:"objectID"`
		UserID   uint `json:"userID"`
	}{
		svc.lot.ObjectID,
		svc.executor.ID,
	}

	return svc.post("ResetDateBook", data)
}

func (svc backService) ActDataSync(info object.JSONData) error {
	data := struct {
		ObjectID uint            `json:"objectID"`
		Info     object.JSONData `json:"info"`
		UserID   uint            `json:"userID"`
	}{
		svc.lot.ObjectID,
		info,
		svc.executor.ID,
	}

	return svc.post("ActDataSync", data)
}

func (svc backService) GetBookingWindow() ([]core.LoadDate, error) {
	data := struct {
		ObjectID uint `json:"objectID"`
		UserID   uint `json:"userID"`
	}{
		svc.lot.ObjectID,
		svc.executor.ID,
	}

	body, err := svc.postWithResult("GetBookingWindow", data)
	if err != nil {
		return nil, err
	}

	var result []core.LoadDate

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}


func (svc backService) ChangeActData(docData interfaces.DocData) error {
	doc := struct {
		ObjectID  uint         `json:"object_id"`
		DocNumber string       `json:"act_number"`
		Date      lib.DateTime `json:"act_date"`
	}{
		ObjectID: docData.ObjectId(),
		DocNumber: docData.DocNumber(),
		Date: docData.DocDate(),
	}

	type response struct {
		Status  bool `json:"status"`
		Message string `json:"message"`
	}

	r := &response{}

	result, err := svc.postWithResult("updateActData", doc)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(result, r); err != nil {
		return errors.BadRequest("incorrect response type")
	}
	if !r.Status {
		return errors.BadRequest(r.Message)
	}
	return nil
}

func (svc backService) postWithResult(method string, data interface{}) ([]byte, error) {
	logger := svc.logger.Named("post").With(zap.String("method", method))

	logger.Debug("post method",
		zap.String("method", method),
		zap.Any("data", data),
	)

	user := svc.lot.User

	backURL, ok := svc.cfg.Managers[user.Username]
	if !ok {
		logger.Error("no set back url")
		return nil, errors.New("no set back url")
	}

	r := resty.R().
		SetHeader("Content-Type", "application/json").
		SetBody(data)

	resp, err := r.Post(backURL + method + "?secret_key=" + user.BackKey)

	if err != nil {
		logger.Error("post failed", zap.Error(err))
		return nil, err
	}

	if resp.StatusCode() != 200 {
		logger.Error("post status code failed",
			zap.Int("code", resp.StatusCode()),
			zap.Any("response", resp),
			zap.String("method", method),
			zap.Any("data", data),
		)
		return nil, errors.New("post status code failed")
	}

	logger.Debug("post success")

	return resp.Body(), nil
}

func (svc backService) post(method string, data interface{}) error {
	if _, err := svc.postWithResult(method, data); err != nil {
		return err
	}
	return nil
}
