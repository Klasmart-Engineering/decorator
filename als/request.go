package als

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"io"
	"net/http"
	"strings"
)

type Request struct {
	events.APIGatewayV2HTTPRequest
	refinedHeader http.Header
	url string
}

func (req *Request) GetHttpReq(ctx context.Context) (httpRequest *http.Request) {
	ctx = context.WithValue(ctx, "originAWSRequest", req)
	req.RebuildUrlAndHeader()
	var bodyReader io.Reader = bytes.NewBufferString(req.Body)
	if req.IsBase64Encoded {
		bodyReader = base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(req.Body))
	}


	httpRequest, err := http.NewRequestWithContext(ctx, req.RequestContext.HTTP.Method, req.url, bodyReader)
	if err != nil {
		panic(fmt.Errorf("failed to convert lambda event to http request for gin, err: %v", err))
	}
	httpRequest.Header = req.refinedHeader
	return
}

func (req *Request) RebuildUrlAndHeader() {
	req.refinedHeader = http.Header{}
	for k, v := range req.Headers {
		req.refinedHeader.Add(k, v)
	}

	// special dealing with cookies
	cookie := strings.Join(req.Cookies, ";")
	req.refinedHeader.Add("Cookie", cookie)

	proto := req.refinedHeader.Get("x-forwarded-proto")
	host := req.refinedHeader.Get("Host")
	rawPath := req.RawPath
	rawQueryString := req.RawQueryString

	req.url = strings.Join([]string{proto, "://", host, rawPath, "?", rawQueryString}, "")
}
