package tcs

import (
	"github.com/tencentyun/scf-go-lib/cloudevents/scf"
	"net/http"
	"strings"
	"unsafe"
)

// Response wraps APIGatewayProxyResponse to http.ResponseWriter
//
// Response.Header return http.Header, Response.httpHeader is the httpHeader map
// provided by APIGatewayProxyResponse.httpHeader, which is actually send
// by scf at last.
type Response struct {
	scf.APIGatewayProxyResponse
	httpHeader  http.Header
	bodyBuilder strings.Builder // in WriteHeader, APIGatewayProxyResponse.Body would be re-point to bodyBuilder
}

func (resp *Response) Header() http.Header {
	if resp.httpHeader == nil {
		resp.httpHeader = http.Header{}
		resp.Headers = map[string]string{}
	}
	return resp.httpHeader
}

func (resp *Response) Write(p []byte) (n int, err error) {
	if resp.httpHeader == nil {
		resp.WriteHeader(http.StatusOK)
	}

	n, err = resp.bodyBuilder.Write(p)

	// re-point resp.Body to string address inside resp.bodyBuilder
	resp.Body = resp.bodyBuilder.String()
	return
}

func (resp *Response) WriteHeader(statusCode int) {
	// detect Content-Type
	httpHeader := resp.Header()
	if _, hasContentType := httpHeader["Content-Type"]; !hasContentType {
		bodyStr := resp.bodyBuilder.String()
		bodyBuf := *(*[]byte)(unsafe.Pointer(&bodyStr))
		contentType := http.DetectContentType(bodyBuf)
		httpHeader.Add("Content-Type", contentType)
	}

	// write into APIGatewayProxyResponse.Header
	// only use first element in array (just like http.Header
	scfHeaders := resp.Headers
	for key := range resp.httpHeader {
		hKey := key
		hVal := resp.httpHeader.Get(key)
		scfHeaders[hKey] = hVal
	}

	resp.StatusCode = statusCode
}
