package curl

import (
	"net/http"
	"time"
)

type Transport struct {
	Transport *http.Transport

	CAPath         string
	ForceHTTP3     bool
	HTTP3LogEnable bool
	Timeout        int64 // 单位：ms
}

func (t *Transport) RoundTrip(request *http.Request) (*http.Response, error) {

	if t.ForceHTTP3 {
		transport := &http3Transport{
			ResolverList:   nil,
			CAPath:         t.CAPath,
			HTTP3LogEnable: t.HTTP3LogEnable,
			ConnectTimeout: int64(t.Transport.IdleConnTimeout / time.Millisecond),
			Timeout:        t.Timeout,
		}
		return transport.RoundTrip(request)
	} else {
		return t.Transport.RoundTrip(request)
	}
}
