package types

import "net/http"

// HTTP status codes
var HTTP = struct {
	Ok                  int
	Created             int
	NoContent           int
	BadRequest          int
	Unauthorized        int
	Forbidden           int
	NotFound            int
	Conflict            int
	InternalServerError int
}{
	Ok:                  http.StatusOK,
	Created:             http.StatusCreated,
	NoContent:           http.StatusNoContent,
	BadRequest:          http.StatusBadRequest,
	Unauthorized:        http.StatusUnauthorized,
	Forbidden:           http.StatusForbidden,
	NotFound:            http.StatusNotFound,
	Conflict:            http.StatusConflict,
	InternalServerError: http.StatusInternalServerError,
}

// Response status values
var ResponseStatus = struct {
	Success string
	Failure string
}{
	Success: "success",
	Failure: "failure",
}
