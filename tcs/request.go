package tcs

import (
	"bytes"
	"context"
	"net/http"
	"strings"

	"github.com/tencentyun/scf-go-lib/cloudevents/scf"
)

type Request struct {
	scf.APIGatewayProxyRequest
	Headers       map[string]interface{} `json:"headers"`
	QueryString   map[string]interface{} `json:"queryString"`
	refinedHeader http.Header
}

func (req *Request) GetQueryRawStr() string {
	var sb strings.Builder
	sb.WriteString("?")

	for k, v := range req.QueryString {
		if str, ok := v.(string); ok {
			writeKV(&sb, k, str)
		} else {
			strArray, _ := v.([]string)
			for _, str := range strArray {
				writeKV(&sb, k, str)
			}
		}
	}

	return sb.String()[:sb.Len()-1]
}

func writeKV(sb *strings.Builder, k, v string) {
	sb.WriteString(k)
	sb.WriteString("=")
	sb.WriteString(v)
	sb.WriteString("&")
}

func (req *Request) GetUrl() string {
	var sb strings.Builder
	//stageStr := req.RequestContext.Stage
	queryRawStr := req.GetQueryRawStr()

	sb.WriteString("https://")
	sb.WriteString(req.refinedHeader.Get("host"))
	//sb.WriteString(fmt.Sprint("/", stageStr))

	// knock off first knot of req.Path, eg: "/test/path/to/api" -> /path/to/api
	//knocker := regexp.MustCompile(`(?U)^/[\W\w]+/`)
	//knockedReqPath := knocker.ReplaceAllString(req.Path, "/")

	sb.WriteString(req.Path)
	sb.WriteString(queryRawStr)

	return sb.String()
}

func (req *Request) GetHttpReq(ctx context.Context) (httpRequest *http.Request) {
	req.RefineHeader()
	urlStr := req.GetUrl()
	body := bytes.NewBufferString(req.Body)

	if _, exist := ctx.Value("origScfReq").(*Request); !exist {
		ctx = context.WithValue(ctx, "origScfReq", req)
	}

	httpRequest, _ = http.NewRequestWithContext(ctx, req.HTTPMethod, urlStr, body)
	httpRequest.Header = req.refinedHeader

	return
}

func (req *Request) RefineHeader() {
	if req.refinedHeader == nil {
		req.refinedHeader = http.Header{}
	}
	for key, val := range req.Headers {
		switch val := val.(type) {
		case string:
			req.refinedHeader.Add(key, val)
		case []string:
			for _, v := range val {
				req.refinedHeader.Add(key, v)
			}
		}
	}
}

func ExtractOrigScfReq(ctx context.Context) (req *Request) {
	// scf original request sure must be exist
	// because req.GetHttpReq will insert into ctx

	req, exist := ctx.Value("origScfReq").(*Request)
	if !exist {
		panic("original scf req not exist!!!")
	}
	return
}
