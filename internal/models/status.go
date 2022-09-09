package models

type Status int

const (
	Ready Status = 1 + iota
	End
	InConvo
)
