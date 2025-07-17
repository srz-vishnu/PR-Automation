package e

import (
	"net/http"
	"strconv"
)

type WrapError struct {
	ErrorCode int
	Msg       string
	RootCause error
}

// For client
type HttpError struct {
	StatusCode int
	Code       int
	Message    string
}

func (e *WrapError) Error() string {
	return e.RootCause.Error()
}

// NewError : create a new error instance, get rootcause error and return as WrapError.
func NewError(errCode int, msg string, rootCause error) *WrapError {
	err := &WrapError{
		ErrorCode: errCode,
		Msg:       msg,
		RootCause: rootCause,
	}
	return err
}

// NewAPIError : create http error from NewError to pass api.Fail.
// err is expecting WrapError type.
func NewAPIError(err error, msg string) *HttpError {
	if err == nil {
		return nil
	}
	// checking err is type of WrapError
	appErr, ok := err.(*WrapError)
	if ok {
		appErr.Msg = msg
	} else {
		return nil
	}

	httpErr := &HttpError{
		StatusCode: GetHttpStatusCode(appErr.ErrorCode),
		Code:       appErr.ErrorCode,
		Message:    msg,
	}
	return httpErr
}

// GetHttpStatusCode used to get Status code from code provided
func GetHttpStatusCode(c int) int {
	str := strconv.Itoa(c)
	// Geting first 3 digits from ErrorCode (eg : 400001 => 400)
	code := str[:3]
	r, _ := strconv.Atoi(code)
	if r < 100 || r >= 600 {
		return http.StatusInternalServerError
	}
	return r
}
