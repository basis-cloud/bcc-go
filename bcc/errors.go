package bcc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ApiError struct {
	msg          string
	code         int
	body         []byte
	errorAliases []string
}

func NewApiError(url string, resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	msg := fmt.Sprintf("HTTP request failure on %s:\n%d: %s", url, resp.StatusCode, string(body))
	var parsedBody struct {
		ErrorAliases []string `json:"error_alias"`
	}
	json.Unmarshal(body, &parsedBody)
	return &ApiError{
		msg:          msg,
		code:         resp.StatusCode,
		body:         body,
		errorAliases: parsedBody.ErrorAliases,
	}
}

func (e *ApiError) Error() string          { return e.msg }
func (e *ApiError) Message() string        { return e.msg }
func (e *ApiError) Code() int              { return e.code }
func (e *ApiError) Body() []byte           { return e.body }
func (e *ApiError) ErrorAliases() []string { return e.errorAliases }
