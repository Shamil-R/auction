package core

type Window struct {
	Start string `json:"start"`
	End   string `json:"end"`
	Open  bool   `json:"open"`
}

type LoadDate struct {
	Date    string   `json:"date"`
	Windows []Window `json:"windows"`
}
