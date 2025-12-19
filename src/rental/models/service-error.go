package models

import "errors"

var (
	Forbidden 		error = errors.New("forbidden")
	InvalidStatus 	error = errors.New("Invalid status")
)
