package models

type ValidationErrorResponse struct {
	Message string `json:"message"`
	Errors map[string]string `json:"errors,omitempty"`
}