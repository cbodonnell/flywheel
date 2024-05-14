package repositories

import "errors"

type ErrNotFound struct {
}

func (e *ErrNotFound) Error() string {
	return "not found"
}

func IsNotFound(err error) bool {
	return errors.Is(err, &ErrNotFound{})
}
