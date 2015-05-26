package ghostmates

import "net/http"

type (
	postmatesTransport func(req *http.Request) (*http.Response, error)
)

func (pt postmatesTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return pt(req)
}

func (pt postmatesTransport) CancelRequest(req *http.Request) {
	type canceler interface {
		CancelRequest(*http.Request)
	}
	http.DefaultTransport.(canceler).CancelRequest(req)
}
