package repositories

import (
	"errors"
)

type ErrNotFound struct {
}

func (e *ErrNotFound) Error() string {
	return "not found"
}

func IsNotFound(err error) bool {
	return errors.Is(err, &ErrNotFound{})
}

type ErrNameExists struct {
}

func (e *ErrNameExists) Error() string {
	return "name already exists"
}

func IsNameExists(err error) bool {
	return errors.Is(err, &ErrNameExists{})
}
