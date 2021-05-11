package curl

import (
	"net/http"
)

type Transport struct {
	Transport *http.Transport

	CAPath         string
	ForceHTTP3     bool
	HTTP3LogEnable bool
}

func (t *Transport) RoundTrip(request *http.Request) (*http.Response, error) {

	if t.ForceHTTP3 {
		transport := &http3Transport{CAPath: t.CAPath, HTTP3LogEnable: t.HTTP3LogEnable}
		return transport.RoundTrip(request)
	} else {
		return t.Transport.RoundTrip(request)
	}
}
