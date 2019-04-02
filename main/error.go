package main

import (
	"github.com/ant0ine/go-json-rest/rest"
)

type ApiError int

func (err ApiError) Error() string {
	switch err {
	case 400:
		return "Bad request"
	case 404:
		return "Resource not found"
	case 409:
		return "Resource conflicts"
	default:
		return "Unknown error"
	}
}

const (
	Unauthorized      ApiError = 401
	ResourceNotFound  ApiError = 404
	ResourceConflicts ApiError = 409
	BadRequest        ApiError = 400
)

func ResponseError(w rest.ResponseWriter, err error) {
	aerr, ok := err.(ApiError)
	if !ok {
		aerr = 500
	}

	w.WriteHeader(int(aerr))
	w.WriteJson(map[string]interface{}{
		"error": aerr.Error(),
	})
}
