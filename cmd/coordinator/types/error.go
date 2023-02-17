package types

// ErrorResponse is the base error message
type ErrorResponse struct {
	Type    string `json:"types"`
	Code    string `json:"code,omitempty"`
	Param   string `json:"param,omitempty"`
	Message string `json:"message"`
}

// ValidationError is the form validation error struct
type ValidationError struct {
	Errors []ErrorResponse
}

func (e *ValidationError) Error() string {
	return "Validation Error"
}
