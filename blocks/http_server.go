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
	"strconv"
	"time"

	"github.com/dbolotin/deadmanswitch/bctx"
	"github.com/dbolotin/deadmanswitch/comm"
	"github.com/dbolotin/deadmanswitch/ctyutil"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zclconf/go-cty/cty"
)

type HttpServer struct {
	SingleChannelBlock
	Address string  `hcl:"address"`
	Timeout *string `hcl:"timeout,optional"`

	Endpoints  []Endpoint          `hcl:"endpoint,block"`
	Monitoring *MonitoringEndpoint `hcl:"monitoring_endpoint,block"`
}

type Endpoint struct {
	Name           *string   `hcl:"name,optional"`
	Methods        *[]string `hcl:"methods,optional"`
	Path           string    `hcl:"path"`
	DiscardBody    bool      `hcl:"discard_body,optional"`
	ParseListAsMap bool      `hcl:"parse_list_as_map,optional"`
	MaxBodySize    *int64    `hcl:"max_body_size,optional"`

	SendTo bctx.ChannelPointer `hcl:"send_to"`
}

type MonitoringEndpoint struct {
	Path string `hcl:"path"`
}

var (
	hsHitsVec             = promauto.NewCounterVec(prometheus.CounterOpts{Name: "http_server_hits"}, []string{"block", "endpoint", "path"})
	hsDecodingErrorsVec   = promauto.NewCounterVec(prometheus.CounterOpts{Name: "http_server_decoding_errors"}, []string{"block", "endpoint", "path"})
	hsDownstreamErrorsVec = promauto.NewCounterVec(prometheus.CounterOpts{Name: "http_server_downstream_errors"}, []string{"block", "endpoint", "path"})
	hsTotalErrorsVec      = promauto.NewCounterVec(prometheus.CounterOpts{Name: "http_server_total_errors"}, []string{"block", "endpoint", "path"})
)

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
		if rwTimeout > timeout {
			rwTimeout = timeout
		}
	}

	if h.Monitoring != nil {
		router.Path(h.Monitoring.Path).Handler(promhttp.Handler())
	}

	// Instantiating endpoints
	for i, ep := range h.Endpoints {
		// Monitoring counters
		pLabels := prometheus.Labels{
			"block":    h.Id,
			"endpoint": StrOrDefault(ep.Name, "ep"+strconv.Itoa(i)),
			"path":     ep.Path,
		}
		mHits := hsHitsVec.With(pLabels)
		mDecodingErrors := hsDecodingErrorsVec.With(pLabels)
		mTotalErrors := hsTotalErrorsVec.With(pLabels)
		mDownstreamErrors := hsDownstreamErrorsVec.With(pLabels)

		// Resolving target communication channel
		sendTo := ep.SendTo.SendCh(ctx)

		// Building route
		route := router.Path(ep.Path)
		if ep.Methods != nil {
			route = route.Methods(*ep.Methods...)
		}
		route.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mHits.Inc()

			// Converting variables from path expansion
			valMap := make(map[string]cty.Value)
			vars := mux.Vars(r)
			if vars != nil && len(vars) > 0 {
				valMap["url"] = ctyutil.StrMapValue(vars)
			}

			// Request context; will be passed along with the message
			ctx := r.Context()

			// Limiting body size
			r.Body = http.MaxBytesReader(w, r.Body, Int64OrDefault(ep.MaxBodySize, 1048576))

			// Parsing request body
			body := cty.NilVal
			if !ep.DiscardBody {
				var err error
				body, err = BodyToValue(r.Body, r.Header)
				if err != nil {
					mDecodingErrors.Inc()
					mTotalErrors.Inc()
					log.Println(err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			}

			// And sending communication message
			valMap["body"] = body
			msg, ch := comm.NewMessageC(ctx, cty.ObjectVal(valMap))
			sendTo <- msg

			var onTimeout <-chan time.Time = nil

			if timeout != 0 {
				onTimeout = time.After(timeout)
			}

			//noinspection GoNilness
			select {
			case <-onTimeout:
				mTotalErrors.Inc()
				http.Error(w, "timeout", http.StatusRequestTimeout)
			case <-r.Context().Done():
				mTotalErrors.Inc()
				log.Println("Client disconnected before receiving reply.")
			case rep := <-ch:
				if rep.Type().IsObjectType() && rep.Type().HasAttribute("err") && rep.GetAttr("err").IsNull() {
					mDownstreamErrors.Inc()
					mTotalErrors.Inc()
					http.Error(w, "error processing request", http.StatusBadRequest)
				}
			}

			// In any case, leaving the handler will lead to request context termination, which can be handled by all downstream actions
		})
	}

	// Creating http server
	srv := &http.Server{
		Handler: router,
		Addr:    h.Address,

		WriteTimeout: rwTimeout,
		ReadTimeout:  rwTimeout,
	}

	// Staring server in a separate go routine
	go func() {
		log.Fatal(srv.ListenAndServe())
	}()

	// Initialized without errors
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
