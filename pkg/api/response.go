package api

import (
	"encoding/json"
	"net/http"
)

const (
	StatusOk   = "ok"
	StatusFail = "notok"
)

type Response struct {
	Status string          `json:"status"`
	Error  *ResponseError  `json:"error,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
}

type ResponseError struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Details []string `json:"details"`
}

func (e ResponseError) Error() string {
	j, err := json.Marshal(e)
	if err != nil {
		return "ResponseError :" + err.Error()
	}
	return string(j)
}

// Success sends a successful JSON response with the standared success format
func Success(w http.ResponseWriter, status int, result interface{}) {
	resultJson, err := json.Marshal(result)
	if err != nil {
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}
	r := &Response{
		Status: StatusOk,
		Result: resultJson,
	}

	respJson, err := json.Marshal(r)
	if err != nil {
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(respJson)
}

// Fail sends an unsuccesful JSON response with the standared failure format
func Fail(w http.ResponseWriter, status, errCode int, msg string, details ...string) {
	// Give error response to client
	r := &Response{
		Status: StatusFail,
		Error: &ResponseError{
			Code:    errCode,
			Message: msg,
			Details: details,
		},
	}

	respJson, err := json.Marshal(r)
	if err != nil {
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(respJson)
}
