package data

import (
	"errors"
)

var (
	ErrDuplicateEmail      = errors.New("email address already exists")
	ErrNotFound            = errors.New("resource not found")
	ErrDoctorUnavailable   = errors.New("doctor unavailable at the specified time")
	ErrResourceUnavailable = errors.New("resource is unavailable during the requested time slot")
)

type PaginationResult struct {
	Total    int64
	Page     int
	PageSize int
}
