package hermes

import (
	"bytes"
	"context"
	"encoding/json"
	"sync"

	"github.com/valyala/fasthttp"
)

var requestPool = &sync.Pool{
	New: func() interface{} {
		return &request{
			validParams: make([]string, 0, 10),
			params:      make([][]byte, 0, 10),
		}
	},
}

type request struct {
	ctx         context.Context
	r           *fasthttp.RequestCtx
	validParams []string
	params      [][]byte
}

func acquireRequest(ctx context.Context, r *fasthttp.RequestCtx) *request {
	req := requestPool.Get().(*request)
	req.r = r
	req.ctx = ctx
	return req
}

func releaseRequest(req *request) {
	req.reset()
	requestPool.Put(req)
}

func (req *request) reset() {
	req.r = nil
	req.ctx = nil
	req.validParams = req.validParams[:0]
	req.params = req.params[:0]
}

func (req *request) Raw() *fasthttp.RequestCtx {
	return req.r
}

func (req *request) Path() []byte {
	return req.r.Path()
}

func (req *request) Method() []byte {
	return req.r.Method()
}

func (req *request) URI() *fasthttp.URI {
	return req.r.URI()
}

func (req *request) Header(name string) []byte {
	return req.r.Request.Header.Peek(name)
}

func (req *request) Host() []byte {
	return req.r.Host()
}

func (req *request) Param(name string) string {
	// req.params is not safe, since its reused over requests
	// but validParams is, so we check if name is one of the
	// valid params, before actually return the value
	for i, p := range req.validParams {
		if p == name {
			return string(req.params[i])
		}
	}
	return ""
}

func (req *request) Query(name string) []byte {
	return req.r.QueryArgs().Peek(name)
}

func (req *request) QueryMulti(name string) [][]byte {
	return req.r.QueryArgs().PeekMulti(name)
}

func (req *request) Data(dst interface{}) error {
	return json.Unmarshal(req.r.PostBody(), dst)
}

func (req *request) Post(name string) []byte {
	return req.r.PostArgs().Peek(name)
}

func (req *request) PostMulti(name string) [][]byte {
	return req.r.PostArgs().PeekMulti(name)
}

func (req *request) Cookie(name string) []byte {
	return req.r.Request.Header.Cookie(name)
}

func (req *request) Context() context.Context {
	return req.ctx
}

func (req *request) WithContext(ctx context.Context) Request {
	req.ctx = ctx
	return req
}

func (req *request) IsJSON() bool {
	ct := req.r.Request.Header.ContentType()
	laj := len(applicationJSON)
	return bytes.Equal(ct, applicationJSON) ||
		((laj < len(ct)) && bytes.Equal(ct[:laj], applicationJSON) && (ct[laj] == ';'))
}

func (req *request) WantsJSON() bool {
	accept := req.r.Request.Header.Peek("Accept")
	laj := len(applicationJSON)
	return bytes.Equal(accept, applicationJSON) ||
		((laj < len(accept)) && bytes.Equal(accept[:laj], applicationJSON))
}
