package domain

import "errors"

// Sentinel errors for business logic.
var (
	ErrNotFound     = errors.New("resource not found")
	ErrBadRequest   = errors.New("bad request")
	ErrUnauthorized = errors.New("unauthorized")
	ErrRateLimited  = errors.New("rate limited")
	ErrInternal     = errors.New("internal server error")
)
