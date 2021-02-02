package curl

import "C"
import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	curl "github.com/YangSen-qn/go-curl/libcurl"
)

var (
	initOnce = sync.Once{}
)

type http3Transport struct {
	CAPath string
}

func (t *http3Transport) RoundTrip(request *http.Request) (response *http.Response, err error) {
	initOnce.Do(func() {
		err = curl.GlobalInit(curl.GLOBAL_ALL)
	})

	easy := curl.EasyInit()
	defer easy.Cleanup()

	if easy == nil {
		err = errors.New("create easy handle error")
		return
	}

	// request default
	if t.CAPath != "" {
		err = easy.Setopt(curl.OPT_CAPATH, t.CAPath)
	}
	err = easy.Setopt(curl.OPT_SSL_VERIFYPEER, false) // 0 is ok
	err = easy.Setopt(curl.OPT_HTTP_VERSION, curl.HTTP_VERSION_2)

	// request url
	err = easy.Setopt(curl.OPT_URL, request.URL.String())

	// method
	switch request.Method {
	case http.MethodGet:
		err = easy.Setopt(curl.OPT_HTTPGET, 1)
	case http.MethodPost:
		err = easy.Setopt(curl.OPT_POST, 1)
	case http.MethodPut:
		err = easy.Setopt(curl.OPT_PUT, 1)
	case http.MethodDelete:
	case http.MethodHead:
		err = easy.Setopt(curl.OPT_HEADER, 1)
	default:
	}

	if err != nil {
		return
	}

	// request header
	requestHeader := make([]string, len(request.Header))
	for key, _ := range request.Header {
		requestHeader = append(requestHeader, key+":"+request.Header.Get(key))
	}
	err = easy.Setopt(curl.OPT_HTTPHEADER, requestHeader)

	if err != nil {
		return
	}

	responseHeader := make(http.Header)
	responseBody := new(bytes.Buffer)
	err = easy.Setopt(curl.OPT_HEADERFUNCTION, func(headField []byte, userData interface{}) bool {
		keyValue := string(headField)
		keyValueList := strings.SplitN(keyValue, ":", 2)
		if len(keyValueList) != 2 {
			return true
		}
		key := keyValueList[0]
		value := keyValueList[1]
		value = strings.ReplaceAll(value, " ", "")
		value = strings.ReplaceAll(value, "\r", "")
		value = strings.ReplaceAll(value, "\n", "")
		responseHeader.Set(key, value)

		return true
	})

	err = easy.Setopt(curl.OPT_WRITEFUNCTION, func(buff []byte, userData interface{}) bool {
		_, err := responseBody.Write(buff)
		if err != nil {
			return false
		} else {
			return true
		}
	})

	err = easy.Setopt(curl.OPT_READFUNCTION, func(buff []byte, userData interface{}) int {
		len, err := request.Body.Read(buff)
		if err == nil {
			return len
		} else {
			return 0
		}
	})

	err = easy.Perform()

	if err == nil {
		statusCodeI, _ := easy.Getinfo(curl.INFO_HTTP_CODE)
		statusCode, _ := statusCodeI.(int)

		response = &http.Response{
			Status:           "",
			StatusCode:       statusCode,
			Proto:            "HTTP/3",
			ProtoMajor:       0,
			ProtoMinor:       0,
			Header:           responseHeader,
			Body:             ioutil.NopCloser(responseBody),
			ContentLength:    int64(responseBody.Len()),
			TransferEncoding: nil,
			Close:            false,
			Uncompressed:     false,
			Trailer:          nil,
			Request:          request,
			TLS:              nil,
		}
	}

	return
}
