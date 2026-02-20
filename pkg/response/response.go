package response

import "github.com/labstack/echo/v4"

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    int    `json:"code"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func Success(c echo.Context, code int, message string, data interface{}) error {
	return c.JSON(code, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(c echo.Context, code int, message string, err string) error {
	return c.JSON(code, Response{
		Success: false,
		Message: message,
		Error:   err,
		Code:    code,
	})
}
