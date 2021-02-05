package curl

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/YangSen-qn/go-curl/v2/libcurl"
)

const (
	Version = "2.0.2"
)

var (
	initOnce = sync.Once{}
)

type http3Transport struct {
	ResolverList []string
	CAPath string
}

func (t *http3Transport) RoundTrip(request *http.Request) (response *http.Response, err error) {
	initOnce.Do(func() {
		err = libcurl.GlobalInit(libcurl.GLOBAL_ALL)
	})
	easy := libcurl.EasyInit()
	defer easy.Cleanup()

	fmt.Printf("===== curl ========= \n")

	if easy == nil {
		err = errors.New("create easy handle error")
		return
	}

	// request default
	if t.CAPath != "" {
		err = easy.Setopt(libcurl.OPT_CAPATH, t.CAPath)
	}

	err = easy.Setopt(libcurl.OPT_SSL_VERIFYHOST, 0) // 0 is ok
	if err != nil {
		return
	}

	err = easy.Setopt(libcurl.OPT_SSL_VERIFYPEER, 0) // 0 is ok
	if err != nil {
		return
	}

	err = easy.Setopt(libcurl.OPT_HTTP_VERSION, libcurl.HTTP_VERSION_3)
	if err != nil {
		return
	}

	// request url
	err = easy.Setopt(libcurl.OPT_URL, request.URL.String())
	if err != nil {
		return
	}

	// method
	switch request.Method {
	case http.MethodGet:
		err = easy.Setopt(libcurl.OPT_HTTPGET, 1)
	case http.MethodPost:
		err = easy.Setopt(libcurl.OPT_POST, 1)
	case http.MethodPut:
		err = easy.Setopt(libcurl.OPT_PUT, 1)
	case http.MethodDelete:
	case http.MethodHead:
		err = easy.Setopt(libcurl.OPT_HEADER, 1)
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
	err = easy.Setopt(libcurl.OPT_HTTPHEADER, requestHeader)
	if err != nil {
		return
	}

	// request resolver
	//resolverList := make([]string, 10)
	//if t.ResolverList != nil {
	//	for _, resolver := range t.ResolverList {
	//		resolverList = append(resolverList, resolver)
	//	}
	//}
	//err = easy.Setopt(libcurl.OPT_RESOLVE, resolverList)
	//if err != nil {
	//	return
	//}

	err = easy.Setopt(libcurl.OPT_HTTPHEADER, requestHeader)
	if err != nil {
		return
	}

	responseHeader := make(http.Header)
	responseBody := new(bytes.Buffer)
	err = easy.Setopt(libcurl.OPT_HEADERFUNCTION, func(headField []byte, userData interface{}) bool {
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
	if err != nil {
		return
	}

	err = easy.Setopt(libcurl.OPT_WRITEFUNCTION, func(buff []byte, userData interface{}) bool {
		_, err := responseBody.Write(buff)
		if err != nil {
			return false
		} else {
			return true
		}
	})
	if err != nil {
		return
	}

	err = easy.Setopt(libcurl.OPT_READFUNCTION, func(buff []byte, userData interface{}) int {
		if request.Body == nil {
			return 0
		}

		len, err := request.Body.Read(buff)
		if err == nil {
			return len
		} else {
			return 0
		}
	})
	if err != nil {
		return
	}

	err = easy.Perform()

	if err == nil {
		statusCodeI, _ := easy.Getinfo(libcurl.INFO_HTTP_CODE)
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
