package object

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"gitlab/nefco/auction/db"
	"net/url"

	validator "gopkg.in/go-playground/validator.v9"
)

var (
	objectTypes = map[string]func() ObjectAction{
		ObjectTypeTrip: func() ObjectAction {
			return &TripAction{}
		},
	}
)

type ObjectFilter interface {
	Fill(values url.Values) error
}

type ObjectData interface {
	CheckFilter(f ObjectFilter) bool
}

type ObjectAction interface {
	ObjectData() ObjectData
	ObjectFilter() ObjectFilter
	CheckConfirm(info JSONData) (bool, error)
	CheckComplete(info JSONData) (bool, error)
	FilterConfirm(info JSONData, f ObjectFilter) bool
}

type ObjectType struct {
	ObjectAction
	objectType string
}

func NewObjectType(objectType string) ObjectType {
	return ObjectType{objectType: objectType}
}

func (o *ObjectType) Type() string {
	return o.objectType
}

func (o *ObjectType) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	if err := json.Unmarshal(b, &o.objectType); err != nil {
		return err
	}
	factory, ok := objectTypes[o.objectType]
	if ok {
		o.ObjectAction = factory()
	}
	return nil
}

func (o ObjectType) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.objectType)
}

func (o ObjectType) Value() (driver.Value, error) {
	return o.objectType, nil
}

func (o *ObjectType) Scan(src interface{}) error {
	s, err := db.ScanString(src)
	if err != nil {
		return err
	}
	o.objectType = s
	return nil
}

type Object struct {
	Type ObjectType `json:"type" validate:"required"`
	Data ObjectData `json:"data" validate:"required"`
}

func (o *Object) UnmarshalJSON(b []byte) error {
	var s struct {
		Type ObjectType      `json:"type"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	o.Type = s.Type
	o.Data = s.Type.ObjectData()
	if err := json.Unmarshal(s.Data, &o.Data); err != nil {
		return err
	}
	return nil
}

func (o Object) Value() (driver.Value, error) {
	return db.JSONValue(&o)
}

func (o *Object) Scan(src interface{}) error {
	s, err := db.ScanString(src)
	if err != nil {
		return err
	}
	return o.UnmarshalJSON([]byte(s))
}

func ValidateObjectType(validate *validator.Validate) {
	validate.RegisterValidation("object_type",
		func(fl validator.FieldLevel) bool {
			objectType := fl.Field().Interface().(*ObjectType)
			_, ok := objectTypes[objectType.objectType]
			return ok
		},
	)
}

type JSONData []byte

func (c JSONData) MarshalJSON() ([]byte, error) {
	if c == nil {
		return []byte("null"), nil
	}
	return c, nil
}

func (c *JSONData) UnmarshalJSON(data []byte) error {
	if c == nil {
		return errors.New("json.RawMessage: UnmarshalJSON on nil pointer")
	}
	*c = append((*c)[0:0], data...)
	return nil
}

func (c JSONData) Value() (driver.Value, error) {
	return db.JSONValue(&c)
}

func (c *JSONData) Scan(src interface{}) error {
	return db.JSONScan(src, c)
}
