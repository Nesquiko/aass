package app

import (
	"fmt"

	"github.com/Nesquiko/wac/pkg/api"
)

type ValidationError struct {
	api.ErrorDetail
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("error %q, status %d, detail %q", e.Title, e.Status, e.Detail)
}
