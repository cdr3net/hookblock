package blocks

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/dbolotin/deadmanswitch/bctx"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/json"
)

type HttpRequest struct {
	SingleChannelBlock
	Method          string                    `hcl:"method"`
	URL             hcl.Expression            `hcl:"url"`
	ContentType     string                    `hcl:"content_type,optional"`
	Headers         map[string]hcl.Expression `hcl:"headers,optional"`
	Body            hcl.Expression            `hcl:"body,optional"`
	DiscardResponse bool                      `hcl:"discard_response,optional"`
}

func (h *HttpRequest) Start(ctx *bctx.BCtx) error {
	// context := &hcl.EvalContext{
	// 	Variables: map[string]cty.Value{
	// 		"env": cty.MapVal(map[string]cty.Value{
	// 			"MY_SECRET_1": cty.StringVal("uggugug"),
	// 		}),
	// 	},
	// 	Functions: nil,
	// }
	// value, diagnostics := h.Body.Value(context)
	// log.Println(value.IsKnown())
	// log.Println(value.IsWhollyKnown())
	// log.Println(value)
	// log.Println(diagnostics)

	ch0 := h.Ch0(ctx)

	var bodySerializer func(value cty.Value) ([]byte, error)

	contentType := "application/json" // "application/x-www-form-urlencoded"

	if h.ContentType == "" || h.ContentType == "json" {
		bodySerializer = func(value cty.Value) ([]byte, error) {
			marshal, err := json.Marshal(value, value.Type())
			return marshal, err
		}
	} else {
		return errors.New("unknown content type \"" + h.ContentType + "\" in block \"" + h.Id + "\"")
	}

	go func() {
		for m := range ch0 {
			msg := m

			context := &hcl.EvalContext{
				Variables: map[string]cty.Value{
					"msg": msg.Value(),
				},
			}

			value, diag := h.Body.Value(context)
			if diag.HasErrors() || !value.IsWhollyKnown() {
				ctx.WriteDiagnostics(diag)
				msg.ReplyWithError()
				continue
			}

			bBody, err := bodySerializer(value)
			if err != nil {
				log.Println(err)
				msg.ReplyWithError()
				continue
			}

			urlV, diag := h.URL.Value(context)

			if diag.HasErrors() {
				ctx.WriteDiagnostics(diag)
				msg.ReplyWithError()
				continue
			}

			url := urlV.AsString()
			req, err := http.NewRequestWithContext(msg.Ctx, h.Method, url, bytes.NewBuffer(bBody))
			if err != nil {
				log.Println(err)
				msg.ReplyWithError()
				continue
			}
			req.Header.Add("Content-Type", contentType)
			req.Header.Add("Content-Length", strconv.Itoa(len(bBody)))

			go func() {
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					log.Println(err)
					msg.ReplyWithError()
					return
				}

				var body cty.Value
				if !h.DiscardResponse {
					body, err = BodyToValue(resp.Body, resp.Header)
					if err != nil {
						log.Println(err)
						return
					}
				} else {

					body = cty.NilVal
				}
				msg.Reply(cty.ObjectVal(map[string]cty.Value{"body": body}))
			}()
		}
	}()
	return nil
}
