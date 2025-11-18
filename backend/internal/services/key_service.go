package services

import (
	"errors"
)

var (
	ErrKeyServiceUnavailable = errors.New("key service unavailable")
	ErrredisUnavailable      = errors.New("redis unavailable")
)
