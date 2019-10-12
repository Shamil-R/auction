package interfaces
// todo: убрать дублирование в других файлах
import "gitlab/nefco/auction/core"

type LotIdGetter interface {
	LotID() uint
}

type LotGetter interface {
	GetLot() *core.Lot
}

type Executor interface {
	Executor() *core.User
}

type ObjectIdGetter interface {
	ObjectId() uint
}
