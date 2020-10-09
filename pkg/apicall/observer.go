package apicall

import (
	"fmt"
	"net/http"
	neturl "net/url"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type observer interface {
	observe(controller string, req *http.Request, resp *http.Response, duration float64)
}

// Observer contains a HistogramVec that it knows how to Observe.
type Observer struct {
	Collector *prometheus.HistogramVec
}

// Validate interface implementation
var _ observer = &Observer{}

// NewAPICallObserver returns a prometheus Collector that times API requests.
// Histogram also gives us a _count metric for free.
func NewAPICallObserver(operatorName string) *Observer {
	name := fmt.Sprintf("%s_api_request_duration_seconds", strings.ReplaceAll(operatorName, "-", "_"))
	return &Observer{
		Collector: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        name,
				Help:        "Distribution of the number of seconds an API request takes",
				ConstLabels: prometheus.Labels{"name": operatorName},
				// We really don't care about quantiles, but omitting Buckets results in defaults.
				// This minimizes the number of unused data points we store.
				Buckets: []float64{1},
			},
			[]string{"controller", "method", "resource", "status"},
		),
	}
}

func (o *Observer) observe(controller string, req *http.Request, resp *http.Response, duration float64) {
	o.Collector.With(prometheus.Labels{
		"controller": controller,
		"method":     req.Method,
		"resource":   resourceFrom(req.URL),
		"status":     resp.Status,
	}).Observe(duration)
}

// resourceFrom normalizes an API request URL, including removing individual namespace and
// resource names, to yield a string of the form:
//     $group/$version/$kind[/{NAME}[/...]]
// or
//     $group/$version/namespaces/{NAMESPACE}/$kind[/{NAME}[/...]]
// ...where $foo is variable, {FOO} is actually {FOO}, and [foo] is optional.
// This is so we can use it as a dimension for the apiCallCount metric, without ending up
// with separate labels for each {namespace x name}.
func resourceFrom(url *neturl.URL) (resource string) {
	defer func() {
		// If we can't parse, return a general bucket. This includes paths that don't start with
		// /api or /apis.
		if r := recover(); r != nil {
			// TODO(efried): Should we be logging these? I guess if we start to see a lot of them...
			resource = "{OTHER}"
		}
	}()

	tokens := strings.Split(url.Path[1:], "/")

	// First normalize to $group/$version/...
	switch tokens[0] {
	case "api":
		// Core resources: /api/$version/...
		// => core/$version/...
		tokens[0] = "core"
	case "apis":
		// Extensions: /apis/$group/$version/...
		// => $group/$version/...
		tokens = tokens[1:]
	default:
		// Something else. Punt.
		panic(1)
	}

	// Single resource, non-namespaced (including a namespace itself): $group/$version/$kind/$name
	if len(tokens) == 4 {
		// Factor out the resource name
		tokens[3] = "{NAME}"
	}

	// Kind or single resource, namespaced: $group/$version/namespaces/$nsname/$kind[/$name[/...]]
	if len(tokens) > 4 && tokens[2] == "namespaces" {
		// Factor out the namespace name
		tokens[3] = "{NAMESPACE}"

		// Single resource, namespaced: $group/$version/namespaces/$nsname/$kind/$name[/...]
		if len(tokens) > 5 {
			// Factor out the resource name
			tokens[5] = "{NAME}"
		}
	}

	resource = strings.Join(tokens, "/")

	return
}
