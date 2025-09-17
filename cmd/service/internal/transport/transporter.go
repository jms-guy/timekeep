package transport

type Transporter struct {
	Shutdown chan struct{}
}

func NewTransporter() *Transporter {
	return &Transporter{Shutdown: make(chan struct{})}
}
