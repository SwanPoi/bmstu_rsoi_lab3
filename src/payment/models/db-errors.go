package models

import "errors"

var (
	ErrorNotFound error = errors.New("not found")
	ErrorAlreadyExists error = errors.New("already exists")
)
