package interfaces

import (
	"gitlab/nefco/auction/core"
	lib "gitlab/nefco/auction/lib"
)

type DocData interface {
	ObjectIdGetter
	DocDateGetter
	DocNumberGetter
}

type ActGetter interface {
	GetObjectAct() *core.Act
}

type DocNumberGetter interface {
	DocNumber() string
}

type DocDateGetter interface {
	DocDate() lib.DateTime
}

type ActAllowChangeParamGetter interface {
	GetAllowChange() int
}

