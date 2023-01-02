package view

type BasicViewModel interface {
	SetError(string)
	SetSuccess(string)
}

type Messages struct {
	Success string
	Error   string
}

func (m *Messages) SetError(message string) {
	m.Error = message
}

func (m *Messages) SetSuccess(message string) {
	m.Success = message
}
