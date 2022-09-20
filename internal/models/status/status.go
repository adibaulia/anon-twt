package status

type Status int

const (
	Ready Status = 1 + iota
	End
	InConvo
	Timeout
)
