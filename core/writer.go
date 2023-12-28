package core

type Writer interface {
	Write(r Result) error
}
