package errors

import (
	"errors"
)

func Is(err1, err2 error) bool {
	return errors.Is(err1, err2)
}
