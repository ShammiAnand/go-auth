package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/shammianand/go-auth/internal/common/types"
)

// RespondJSON sends a JSON response using the standard ApiResponse structure
func RespondJSON(c *gin.Context, statusCode int, response types.ApiResponse) {
	c.JSON(statusCode, response)
}

// RespondSuccess sends a successful JSON response
func RespondSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	RespondJSON(c, statusCode, types.SuccessResponse(message, data))
}

// RespondError sends an error JSON response
func RespondError(c *gin.Context, statusCode int, message string, errorCode string, errorMsg string) {
	RespondJSON(c, statusCode, types.ErrorResponse(message, errorCode, errorMsg))
}

// BindJSON binds request JSON to a struct and handles errors
func BindJSON(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		RespondError(c, types.HTTP.BadRequest, "Invalid request body", "INVALID_JSON", err.Error())
		return err
	}
	return nil
}
