package blocks

import (
	"bytes"
	context2 "context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/dbolotin/deadmanswitch/bctx"
	"github.com/dbolotin/deadmanswitch/comm"
	"github.com/dbolotin/deadmanswitch/ctyutil"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/json"
)

type HttpRequest struct {
	SingleChannelBlock
	Method          string                    `hcl:"method"`
	URL             hcl.Expression            `hcl:"url"`
	Timeout         *string                   `hcl:"timeout,optional"`
	Headers         map[string]hcl.Expression `hcl:"headers,optional"`
	BasicAuth       *BasicAuth                `hcl:"basic_auth,optional"`
	Encoding        string                    `hcl:"encoding,optional"`
	Body            hcl.Expression            `hcl:"body,optional"`
	DiscardResponse bool                      `hcl:"discard_response,optional"`
}

type BasicAuth struct {
	User     string `cty:"user"`
	Password string `cty:"password"`
}

// TODO monitoring

func (h *HttpRequest) Start(env *bctx.BEnv) error {
	// Input channel
	ch0 := h.Ch0(env)

	timeout, err := time.ParseDuration(StrOrDefault(h.Timeout, "15s"))
	if err != nil {
		return err
	}

	// Creating body serializer
	var bodySerializer func(value cty.Value) ([]byte, error)
	var contentType string
	if h.Encoding == "" || h.Encoding == "json" {
		contentType = "application/json"
		bodySerializer = func(value cty.Value) ([]byte, error) {
			marshal, err := json.Marshal(value, value.Type())
			return marshal, err
		}
	} else if h.Encoding == "urlencoded" || h.Encoding == "url" {
		contentType = "application/x-www-form-urlencoded"
		bodySerializer = func(value cty.Value) (res []byte, err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("%s", r)
				}
			}()

			if !value.Type().IsObjectType() && !value.Type().IsMapType() {
				return nil, errors.New("can't decode the value as urlencoded string")
			}

			data := url.Values{}
			for k, v := range value.AsValueMap() {
				vv, err := convert.Convert(v, cty.String)
				if err != nil {
					return nil, err
				}
				data.Add(k, vv.AsString())
			}
			return []byte(data.Encode()), nil
		}
	} else if h.Encoding == "raw" {
		contentType = "text/plain"
		bodySerializer = func(value cty.Value) (res []byte, err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("%s", r)
				}
			}()
			return []byte(value.AsString()), nil
		}
	} else {
		return errors.New("unknown content type \"" + h.Encoding + "\" in block \"" + h.Id + "\"")
	}

	// Spinning up handing goroutine
	env.StartProcessing(ch0, func(msg comm.Msg) error {
		// Creating the evaluation context
		evCtx := env.DefaultEvaluationContext(&msg)

		// Executing body expression
		vBody, err := bctx.EvaluateExpression(h.Body, evCtx)
		if err != nil {
			return err
		}
		bBody, err := bodySerializer(vBody)
		if err != nil {
			return err
		}

		vUrl, err := bctx.EvaluateExpression(h.URL, evCtx)
		if err != nil {
			return err
		}
		sUrl := vUrl.AsString()

		// Setting up request
		cCtx := msg.Ctx
		cCtx, _ = context2.WithTimeout(cCtx, timeout)
		req, err := http.NewRequestWithContext(cCtx, h.Method, sUrl, bytes.NewBuffer(bBody))
		if err != nil {
			return err
		}
		req.Header.Add("Content-Type", contentType)
		req.Header.Add("Content-Length", strconv.Itoa(len(bBody)))

		if h.BasicAuth != nil {
			req.SetBasicAuth(h.BasicAuth.User, h.BasicAuth.Password)
		}

		// Executing request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		// Handing response body
		var responseBody cty.Value
		if !h.DiscardResponse {
			responseBody, err = BodyToValue(resp.Body, resp.Header)
			if err != nil {
				return err
			}
		} else {
			_, err := io.Copy(ioutil.Discard, resp.Body)
			if err != nil {
				return err
			}
			responseBody = ctyutil.StrNullVal
		}

		// Success
		msg.Reply(cty.ObjectVal(map[string]cty.Value{"body": responseBody}))
		return nil
	})

	return nil
}
