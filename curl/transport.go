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
		transport := &http3Transport{
			ResolverList:   nil,
			CAPath:         t.CAPath,
			HTTP3LogEnable: t.HTTP3LogEnable,
			Timeout:        t.Transport.IdleConnTimeout.Seconds(),
		}
		return transport.RoundTrip(request)
	} else {
		return t.Transport.RoundTrip(request)
	}
}
