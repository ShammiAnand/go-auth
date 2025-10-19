package types

// ApiResponse represents the standard API response structure
type ApiResponse struct {
	Status  string      `json:"status"`           // "success" or "failure"
	Message string      `json:"message"`          // Human-readable message
	Data    interface{} `json:"data,omitempty"`   // Response payload (omitted if nil)
	Error   *ErrorDetail `json:"error,omitempty"` // Error details (omitted if nil)
}

// ErrorDetail provides structured error information
type ErrorDetail struct {
	ErrorCode string      `json:"error_code"`         // Machine-readable error code
	ErrorMsg  string      `json:"error_msg"`          // Human-readable error message
	Details   interface{} `json:"details,omitempty"`  // Additional error context
}

// SuccessResponse creates a success API response
func SuccessResponse(message string, data interface{}) ApiResponse {
	return ApiResponse{
		Status:  ResponseStatus.Success,
		Message: message,
		Data:    data,
	}
}

// ErrorResponse creates an error API response
func ErrorResponse(message string, errorCode string, errorMsg string) ApiResponse {
	return ApiResponse{
		Status:  ResponseStatus.Failure,
		Message: message,
		Error: &ErrorDetail{
			ErrorCode: errorCode,
			ErrorMsg:  errorMsg,
		},
	}
}
