package repositories

type ErrNotFound struct {
}

func (e *ErrNotFound) Error() string {
	return "not found"
}

func IsNotFound(err error) bool {
	_, ok := err.(*ErrNotFound)
	return ok
}
