package app

import (
	"errors"

	"github.com/Nesquiko/wac/pkg/data"
)

var (
	ErrDuplicateEmail      = errors.New("email address already exists")
	ErrNotFound            = errors.New("resource not found")
	ErrDoctorUnavailable   = errors.New("doctor unavailable at the specified time")
	ErrResourceUnavailable = errors.New("resource is unavailable during the requested time slot")
)

func New(db *data.MongoDb) MonolithApp { return MonolithApp{db} }

type MonolithApp struct {
	db *data.MongoDb
}
