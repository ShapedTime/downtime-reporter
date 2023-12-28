package core

type Reader interface {
	Read() (Result, error)
}
