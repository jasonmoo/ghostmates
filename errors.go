package ghostmates

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type (
	Error struct {
		StatusCode int    `json:"-"`
		Status     string `json:"-"`

		Kind    string                 `json:"kind"`
		Code    string                 `json:"code"`
		Message string                 `json:"message"`
		Params  map[string]interface{} `json:"params"`
	}
)

const (
	ErrorCodeInvalidParams            = "invalid_params"              // The indicated parameters were missing or invalid.
	ErrorCodeUnknownLocation          = "unknown_location"            // We weren't able to understand the provided address. This usually indicates the address is wrong, or perhaps not exact enough.
	ErrorCodeRequestRateLimitExceeded = "request_rate_limit_exceeded" // This API key has made too many requests.
	ErrorCodeAccountSuspended         = "account_suspended"           //
	ErrorCodeNotFound                 = "not_found"                   //
	ErrorCodeServiceUnavailable       = "service_unavailable"         //
	ErrorCodeDeliveryLimitExceeded    = "delivery_limit_exceeded"     // You have hit the maximum amount of ongoing deliveries allowed.
	ErrorCodeAddressUndeliverable     = "address_undeliverable"

	ErrorKind = "error"
)

func NewError(resp *http.Response) *Error {
	e := &Error{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
	}
	json.NewDecoder(resp.Body).Decode(e)
	return e
}

func (e *Error) Error() string {
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "Postmates API Error (%d %s)", e.StatusCode, e.Status)
	if len(e.Kind) > 0 {
		fmt.Fprintf(buf, " Kind: %s, Code: %s, Message: %s, Params: %v", e.Kind, e.Code, e.Message, e.Params)
	}
	return buf.String()
}
