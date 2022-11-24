package model

type Args string

const (
	Monitor      = "monitor"
	AlertContact = "alertcontact"
	Delete       = "delete"
	Create       = "create"
	Update       = "update"
)

type Result struct {
	ErrorResultField error
	NameResultField  interface{}
}

const (
	ErrorResultField       = "error"
	MonitorNameResultField = "monitor"
)
