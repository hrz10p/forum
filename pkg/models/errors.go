package models

import (
	"database/sql"
	"errors"
)

var (
	ErrSqlNoRows             = sql.ErrNoRows
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrSessionExpired        = errors.New("session expired")
	UniqueConstraintEmail    = errors.New("duplicate email")
	UniqueConstraintUsername = errors.New("duplicate username")
	ErrSessionAlreadyExists  = errors.New("session already exists")
	NotFoundAnything         = errors.New("no data")
	ValueMismatch            = errors.New("value is incorrect")
	NoCatsSelected           = errors.New("no cats selected")
	SignIsMismatch           = errors.New("sign is mismatched")
)
