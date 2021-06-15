package curl

// #include <pthread.h>
// #include "../libcurl/include/curl.h"
import "C"

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/YangSen-qn/go-curl/v2/libcurl"
)

var (
	initOnce = sync.Once{}
	easyLock sync.Mutex
	rander   = rand.New(rand.NewSource(time.Now().UnixNano()))
)

type http3Transport struct {
	ResolverList   []string
	CAPath         string
	HTTP3LogEnable bool
}

func (t *http3Transport) RoundTrip(request *http.Request) (response *http.Response, err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	initOnce.Do(func() {
		err = libcurl.GlobalInit(libcurl.GLOBAL_ALL)
	})

	easyLock.Lock()
	easy := libcurl.EasyInit()
	randNum := rander.Uint64()
	easyLock.Unlock()

	defer func() {
		easyLock.Lock()
		easy.Cleanup()
		easyLock.Unlock()
	}()

	if easy == nil {
		err = errors.New("create easy handle error")
		return
	}

	// request default
	if t.CAPath != "" {
		err = easy.Setopt(libcurl.OPT_CAPATH, t.CAPath)
		if err != nil {
			return
		}
	}

	if t.HTTP3LogEnable {
		err = easy.Setopt(libcurl.OPT_VERBOSE, 1)
		if err != nil {
			return
		}
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
	query := request.URL.Query()
	query.Add("q", strconv.FormatUint(randNum, 10))
	request.URL.RawQuery = query.Encode()
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
		err = easy.Setopt(libcurl.OPT_UPLOAD, 1)
	case http.MethodDelete:
		err = easy.Setopt(libcurl.OPT_CUSTOMREQUEST, "DELETE")
	case http.MethodHead:
		err = easy.Setopt(libcurl.OPT_NOBODY, 1)
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
	if request.Header.Get("Expect") == "" {
		requestHeader = append(requestHeader, "Expect:")
	}
	err = easy.Setopt(libcurl.OPT_HTTPHEADER, requestHeader)
	if err != nil {
		return
	}

	switch request.Method {
	case http.MethodPut:
		err = easy.Setopt(libcurl.OPT_INFILESIZE, request.ContentLength)
	case http.MethodPost:
		err = easy.Setopt(libcurl.OPT_POSTFIELDSIZE, request.ContentLength)
	}
	if err != nil {
		return
	}

	// request resolver
	resolverList := make([]string, 10)
	if len(t.ResolverList) > 0 {
		for _, resolver := range t.ResolverList {
			resolverList = append(resolverList, resolver)
		}
		err = easy.Setopt(libcurl.OPT_RESOLVE, resolverList)
		if err != nil {
			return
		}
	}
	err = easy.Setopt(libcurl.OPT_NOSIGNAL, 1)
	if err != nil {
		return
	}
	err = easy.Setopt(libcurl.OPT_TRANSFER_ENCODING, 0)
	if err != nil {
		return
	}
	err = easy.Setopt(libcurl.OPT_TCP_KEEPALIVE, 1)
	if err != nil {
		return
	}
	err = easy.Setopt(libcurl.OPT_TIMEOUT, 180)
	if err != nil {
		return
	}
	err = easy.Setopt(libcurl.OPT_CONNECTTIMEOUT, 64)
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

	err = easy.Setopt(libcurl.OPT_DEBUGFUNCTION, func(buff []byte, userData interface{}) int {
		fmt.Printf("*** [%s] [%d] %s", time.Now().Format(time.RFC3339Nano), randNum, buff)
		return 0
	})
	if err != nil {
		return
	}

	err = easy.Perform()

	if curlErr, ok := err.(libcurl.CurlError); ok {
		if curlErr == C.CURLE_OPERATION_TIMEDOUT {
			fmt.Printf("*** [%s] FATAL ERROR: CURL TIMEOUT, URL = %s, randNum = %d\n", time.Now().Format(time.RFC3339Nano), request.URL, randNum)
			os.Exit(1)
		}
	}

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
