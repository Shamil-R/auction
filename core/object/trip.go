package object

import (
	"encoding/json"
	"gitlab/nefco/auction/core/object/json_fileds"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	validator "gopkg.in/go-playground/validator.v9"
)

const ObjectTypeTrip = "trip"

type TripFilter struct {
	CityFrom      string     `json:"city_from"`
	CityTo        string     `json:"city_to"`
	FromStartDate *time.Time `json:"from_start_date"`
	FromEndDate   *time.Time `json:"from_end_date"`
	ToStartDate   *time.Time `json:"to_start_date"`
	ToEndDate     *time.Time `json:"to_end_date"`
	StartTonnage  uint       `json:"start_tonnage"`
	EndTonnage    uint       `json:"end_tonnage"`
	Invoice       string     `json:"invoice"`
	Driver        string     `json:"driver"`
}

func (f *TripFilter) Fill(values url.Values) error {
	if cityFrom := values.Get("city_from"); len(cityFrom) != 0 {
		f.CityFrom = cityFrom
	}
	if cityTo := values.Get("city_to"); len(cityTo) != 0 {
		f.CityTo = cityTo
	}
	if paramStartDate := values.Get("from_start_date"); len(paramStartDate) != 0 {
		startDate, err := time.Parse(time.RFC3339, paramStartDate)
		if err != nil {
			return err
		}
		f.FromStartDate = &startDate
	}
	if paramEndDate := values.Get("from_end_date"); len(paramEndDate) != 0 {
		endDate, err := time.Parse(time.RFC3339, paramEndDate)
		if err != nil {
			return err
		}
		f.FromEndDate = &endDate
	}
	if paramStartDate := values.Get("to_start_date"); len(paramStartDate) != 0 {
		startDate, err := time.Parse(time.RFC3339, paramStartDate)
		if err != nil {
			return err
		}
		f.ToStartDate = &startDate
	}
	if paramEndDate := values.Get("to_end_date"); len(paramEndDate) != 0 {
		endDate, err := time.Parse(time.RFC3339, paramEndDate)
		if err != nil {
			return err
		}
		f.ToEndDate = &endDate
	}
	if paramStartTonnage := values.Get("start_tonnage"); len(paramStartTonnage) != 0 {
		startTonnage, err := strconv.ParseUint(paramStartTonnage, 10, 32)
		if err != nil {
			return err
		}
		f.StartTonnage = uint(startTonnage)
	}
	if paramEndTonnage := values.Get("end_tonnage"); len(paramEndTonnage) != 0 {
		endTonnage, err := strconv.ParseUint(paramEndTonnage, 10, 32)
		if err != nil {
			return err
		}
		f.EndTonnage = uint(endTonnage)
	}
	if invoice := values.Get("invoice"); len(invoice) != 0 {
		f.Invoice = invoice
	}
	if driver := values.Get("driver"); len(driver) != 0 {
		f.Driver = driver
	}
	return nil
}

type Order struct {
	ConsigneeName    string   `json:"consignee_name" validate:"required"`
	ConsigneeAddress string   `json:"consignee_address" validate:"required"`
	Invoices         []string `json:"invoices,omitempty"`
}

type Point struct {
	Unloading bool      `json:"unloading" validate:"required"`
	Company   string    `json:"company,omitempty"`
	Address   string    `json:"address,omitempty" validate:"required"`
	Date      time.Time `json:"date" validate:"required"`
	Note      string    `json:"note,omitempty"`
}

type Trip struct {
	Orders      []*Order `json:"orders" validate:"required,min=1"`
	Points      []*Point `json:"points" validate:"required,min=2"`
	Tonnage     uint     `json:"tonnage" validate:"required"`
	Type        string   `json:"type,omitempty"`
	Description string   `json:"description,omitempty"`
	DocsPackURL string   `json:"docs_pack_url" validate:"required,url"`
}

func (t Trip) CheckFilter(of ObjectFilter) bool {
	tf, ok := of.(*TripFilter)
	if !ok {
		return false
	}

	sort.Slice(t.Points, func(a int, b int) bool {
		return t.Points[a].Date.Before(t.Points[b].Date)
	})

	fromPoint := t.Points[0]
	toPoint := t.Points[len(t.Points)-1]

	if len(strings.TrimSpace(tf.CityFrom)) > 0 && !strings.Contains(
		strings.ToLower(fromPoint.Address), strings.ToLower(tf.CityFrom)) {
		return false
	}

	if len(strings.TrimSpace(tf.CityTo)) > 0 && !strings.Contains(
		strings.ToLower(toPoint.Address), strings.ToLower(tf.CityTo)) {
		return false
	}

	if tf.FromStartDate != nil && fromPoint.Date.Before(*tf.FromStartDate) {
		return false
	}

	if tf.FromEndDate != nil && fromPoint.Date.After(*tf.FromEndDate) {
		return false
	}

	if tf.ToStartDate != nil && toPoint.Date.Before(*tf.ToStartDate) {
		return false
	}

	if tf.ToEndDate != nil && toPoint.Date.After(*tf.ToEndDate) {
		return false
	}

	if tf.StartTonnage > 0 && t.Tonnage < tf.StartTonnage {
		return false
	}

	if tf.EndTonnage > 0 && t.Tonnage > tf.EndTonnage {
		return false
	}

	invoice := strings.TrimSpace(tf.Invoice)
	if len(invoice) > 0 && t.Orders != nil {
		f := false
		for _, order := range t.Orders {
			if order.Invoices == nil {
				continue
			}
			for _, invoice := range order.Invoices {
				if strings.Contains(
					strings.ToLower(invoice),
					strings.ToLower(tf.Invoice),
				) {
					f = true
					break
				}
			}
			if f {
				break
			}
		}
		if !f {
			return false
		}
	}

	return true
}

type TripAction struct{}

func (TripAction) ObjectData() ObjectData {
	return &Trip{}
}

func (TripAction) ObjectFilter() ObjectFilter {
	return &TripFilter{}
}

type TripConfirm struct {
	Surname        string     `json:"surname" validate:"required"`
	Name           string     `json:"name" validate:"required"`
	Patronymic     *string    `json:"patronymic,omitempty"`
	TruckNumber    string     `json:"truck_number" validate:"required"`
	TrailerNumber  *string    `json:"trailer_number,omitempty"`
	PhoneNumber    string     `json:"phone_number" validate:"required"`
	PassportSerie  *string    `json:"passport_serie,omitempty"`
	PassportNumber *string    `json:"passport_number,omitempty"`
	PassportIssued *string    `json:"passport_issued,omitempty"`
	PassportDate   *time.Time `json:"passport_date,omitempty"`
	DrLicSerie     *string    `json:"dr_lic_serie,omitempty"`
	DrLicNumber    *string    `json:"dr_lic_number,omitempty"`
	DateArrival    *time.Time `json:"date_arrival" validate:"required"`
}

func (TripAction) CheckConfirm(info JSONData) (bool, error) {
	confirm := TripConfirm{}

	if err := json.Unmarshal(info, &confirm); err != nil {
		return false, err
	}

	validate := validator.New()
	if err := validate.Struct(&confirm); err != nil {
		return false, nil
	}

	return true, nil
}

func (TripAction) FilterConfirm(info JSONData, f ObjectFilter) bool {
	if info == nil || len(info) == 0 {
		info = []byte("null")
	}

	tf, ok := f.(*TripFilter)
	if !ok {
		return false
	}

	confirm := TripConfirm{}

	if err := json.Unmarshal(info, &confirm); err != nil {
		return false
	}

	driver := strings.ToUpper(strings.Replace(tf.Driver, " ", "", -1))
	if len(driver) > 0 {
		name := strings.TrimSpace(confirm.Surname) +
			strings.TrimSpace(confirm.Name)
		if confirm.Patronymic != nil {
			name += strings.TrimSpace(*confirm.Patronymic)
		}
		if !strings.Contains(strings.ToUpper(name), driver) {
			return false
		}
	}

	return true
}

func (TripAction) CheckComplete(info JSONData) (bool, error) {
	complete := &json_fileds.Complete{}

	if err := json.Unmarshal(info, complete); err != nil {
		return false, err
	}

	validate := validator.New()
	if err := validate.Struct(complete); err != nil {
		return false, nil
	}

	return true, nil
}
