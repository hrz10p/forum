package models

import "time"

type Session struct {
	ID         string
	UID        string
	ExpireTime time.Time
}
