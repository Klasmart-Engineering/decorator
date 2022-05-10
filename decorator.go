package decorator

import (
	"context"
	"net/http"

	"github.com/KL-Engineering/decorator/als"
	"github.com/KL-Engineering/decorator/tcs"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/tencentyun/scf-go-lib/cloudfunction"
)

type RunningEnv string

const (
	EnvHTTP   RunningEnv = "HTTP"
	EnvSCF    RunningEnv = "SCF"
	EnvLAMBDA RunningEnv = "LAMBDA"
)

type decoratorHandler struct {
	http.Handler
}

func (d *decoratorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d.Handler.ServeHTTP(w, r)
}

var env RunningEnv

// RunWithHTTPHandler 是服务启动入口
func RunWithHTTPHandler(handler http.Handler, addr string) {
	dec := &decoratorHandler{handler}

	switch env {
	case EnvSCF:
		cloudfunction.Start(func(ctx context.Context, req *tcs.Request) (response *tcs.Response, err error) {
			response = &tcs.Response{}
			dec.ServeHTTP(response, req.GetHttpReq(ctx))
			return
		})
	case EnvLAMBDA:
		httpHandler := func(ctx context.Context, req *als.Request) (response *als.Response, err error) {
			response = &als.Response{}
			dec.ServeHTTP(response, req.GetHttpReq(ctx))
			return
		}
		lambda.Start(httpHandler)
	default:
		if err := http.ListenAndServe(addr, dec); err != nil {
			panic(err)
		}
	}
}

func Setenv(e RunningEnv) {
	env = e
}
