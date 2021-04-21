package web

import "errors"

var (
	// ErrPing is not really an error, but is a special status
	// code that's returned during version extraction to abort
	// further processing because this was just a health check
	// from the remote system.
	ErrPing = errors.New("pong")
)
