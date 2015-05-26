package ghostmates

import (
	"net/http"
	"testing"
)

func TestError(t *testing.T) {

	client := NewClient(TestCustomerId, TestAPIKey, TestTimeout)

	_, err := client.GetQuote("yyy", "xxx")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	e, ok := err.(*Error)
	if !ok {
		t.Errorf("Expected type *Error, got type %T", e)
	}

	if e.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected %d, got %d", http.StatusBadRequest, e.StatusCode)
	}
	if len(e.Status) == 0 {
		t.Errorf("Expected a Status")
	}
	if e.Kind != ErrorKind {
		t.Errorf("Expected %q, got %q", ErrorKind, e.Kind)
	}
	if e.Code != ErrorCodeAddressUndeliverable {
		t.Errorf("Expected %q, got %q", ErrorCodeAddressUndeliverable, e.Code)
	}
	if len(e.Message) == 0 {
		t.Errorf("Expected non-zero length message")
	}
	if e.Error() != "Postmates API Error (400 400 BAD REQUEST) Kind: error, Code: address_undeliverable, Message: The specified location is not in a deliverable area., Params: map[]" {
		t.Errorf("Error() output string does not match expected.  Output: %q", e.Error())
	}

}
