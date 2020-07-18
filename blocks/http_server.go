package blocks

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/url"
	"time"

	"github.com/dbolotin/deadmanswitch/bctx"
	"github.com/dbolotin/deadmanswitch/comm"
	"github.com/dbolotin/deadmanswitch/ctyutil"
	"github.com/gorilla/mux"
	"github.com/zclconf/go-cty/cty"
)

type Endpoint struct {
	Name           *string   `hcl:"name,optional"`
	Methods        *[]string `hcl:"methods,optional"`
	Path           string    `hcl:"path"`
	DiscardBody    bool      `hcl:"discard_body,optional"`
	ParseListAsMap bool      `hcl:"parse_list_as_map,optional"`

	SendTo bctx.ChannelPointer `hcl:"send_to"`
}

type HttpServer struct {
	SingleChannelBlock
	Address string  `hcl:"address"`
	Timeout *string `hcl:"timeout,optional"`

	Endpoints []Endpoint `hcl:"endpoint,block"`
}

// Inspired by: https://gist.github.com/rjz/fe283b02cbaa50c5991e1ba921adf7c9
//
// Determine whether the request `content-type` includes a
// server-acceptable mime-type
//
// Failure should yield an HTTP 415 (`http.StatusUnsupportedMediaType`)
func HasContentType(header http.Header, mimetype string) bool {
	contentType := header.Get("Content-Type")
	if contentType == "" {
		return false
	}
	t, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	return t == mimetype
}

func (h *HttpServer) Start(ctx *bctx.BCtx) error {
	router := mux.NewRouter()

	timeout := 0 * time.Second
	rwTimeout := 15 * time.Second

	if h.Timeout != nil {
		var err error
		timeout, err = time.ParseDuration(*h.Timeout)
		if err != nil {
			return err
		}
	}

	for _, ep := range h.Endpoints {
		sendTo := ep.SendTo.SendCh(ctx)

		route := router.Path(ep.Path)

		if ep.Methods != nil {
			route = route.Methods(*ep.Methods...)
		}

		route.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			valMap := make(map[string]cty.Value)

			vars := mux.Vars(r)
			if vars != nil && len(vars) > 0 {
				valMap["url"] = ctyutil.StrMapValue(vars)
			}

			ctx := r.Context()

			// TODO Parametrize
			// Limiting body size
			r.Body = http.MaxBytesReader(w, r.Body, 1048576)

			var body cty.Value = cty.NilVal

			if !ep.DiscardBody {
				var err error
				body, err = BodyToValue(r.Body, r.Header)
				if err != nil {
					log.Println(err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			}

			val := cty.MapVal(map[string]cty.Value{
				"body": body,
			})

			msg, ch := comm.NewMessageC(ctx, val)

			sendTo <- msg

			var onTimeout <-chan time.Time = nil

			if timeout != 0 {
				onTimeout = time.After(timeout)
			}

			select {
			case <-onTimeout:
				http.Error(w, "timeout", http.StatusRequestTimeout)
			case <-r.Context().Done():
				log.Println("Client disconnected before receiving reply.")
			case rep := <-ch:
				if rep.Type().IsObjectType() && rep.Type().HasAttribute("err") && rep.GetAttr("err").IsNull() {
					http.Error(w, "error processing request", http.StatusBadRequest)
					return
				}
			}
		})
	}

	srv := &http.Server{
		Handler: router,
		Addr:    h.Address,

		WriteTimeout: rwTimeout,
		ReadTimeout:  rwTimeout,
	}

	go func() {
		log.Fatal(srv.ListenAndServe())
	}()

	return nil
}

func BodyToValue(body io.ReadCloser, header http.Header) (cty.Value, error) {
	if HasContentType(header, "application/json") {
		dec := json.NewDecoder(body)
		var v interface{}
		err := dec.Decode(&v)
		if err != nil {
			log.Println(err)
			return cty.Value{}, errors.New("error decoding request body")
		}
		body, err := ctyutil.Convert(v)
		if err != nil {
			log.Println(err)
			return cty.Value{}, errors.New("error converting request body")
		}
		return body, nil
	} else if HasContentType(header, "application/x-www-form-urlencoded") {
		bytes, err := ioutil.ReadAll(body)
		if err != nil {
			log.Println(err)
			return cty.Value{}, errors.New("error parsing request")
		}
		query, err := url.ParseQuery(string(bytes))
		if err != nil {
			log.Println(err)
			return cty.Value{}, errors.New("error parsing request")
		}
		bb := make(map[string]string)
		for k, v := range query {
			bb[k] = v[0]
		}
		return ctyutil.StrMapValue(bb), nil
	} else {
		return cty.NilVal, nil
	}
}
