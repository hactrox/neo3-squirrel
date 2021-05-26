package models

type ABI struct {
	Events  []Event       `json:"events"`
	Methods []EventMethod `json:"methods"`
}

type Event struct {
	Name       string           `json:"name"`
	Parameters []EventParameter `json:"parameters"`
}

type EventParameter struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type EventMethod struct {
	Name       string           `json:"name"`
	Safe       bool             `json:"safe"`
	Offset     uint             `json:"offset"`
	Parameters []EventParameter `json:"parameters"`
	ReturnType string           `json:"returntype"`
}
