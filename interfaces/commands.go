package interfaces

// Интерфейсы команд, которые должны быть реализованы

type AutoBookingCommander interface {
	LotGetter
	Executor
	UserIdGetter
}

type EditActCommander interface {
	ActGetter
	Executor
	LotIdGetter
}

type AllowChangeActCommander interface {
	Executor
	LotIdGetter
	ActAllowChangeParamGetter
}