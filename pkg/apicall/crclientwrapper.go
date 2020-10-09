package apicall

import (
	"net/http"
	"time"

	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// NewCRClientWithMetrics creates a new controller-runtime client with a wrapper which increments
// metrics for requests by controller name, HTTP method, URL path, and HTTP status. The client will
// re-use the manager's cache. This should be used in all controllers.
// - param mgr: Your controller-runtime Manager
// - param controller: The name of your controller. This will be the value of the `controller`
//       label for API call metrics recorded by the returned Client
// - param observer: An Observer produced by NewAPICallObserver()
func NewCRClientWithMetrics(mgr manager.Manager, controller string, observer *Observer) (client.Client, error) {
	// Copy the rest.Config as we want our round trippers to be controller-specific.
	cfg := rest.CopyConfig(mgr.GetConfig())
	addControllerMetricsTransportWrapper(cfg, controller, observer)

	options := client.Options{
		Scheme: mgr.GetScheme(),
		Mapper: mgr.GetRESTMapper(),
	}
	c, err := client.New(cfg, options)
	if err != nil {
		return nil, err
	}

	return &client.DelegatingClient{
		Reader: &client.DelegatingReader{
			CacheReader:  mgr.GetCache(),
			ClientReader: c,
		},
		Writer:       c,
		StatusClient: c,
	}, nil
}

// addControllerMetricsTransportWrapper adds a transport wrapper to the given rest config which
// exposes metrics based on the requests being made.
func addControllerMetricsTransportWrapper(cfg *rest.Config, controllerName string, observer *Observer) {
	// If the restConfig already has a transport wrapper, wrap it.
	if cfg.WrapTransport != nil {
		origFunc := cfg.WrapTransport
		cfg.WrapTransport = func(rt http.RoundTripper) http.RoundTripper {
			return &controllerMetricsTripper{
				RoundTripper: origFunc(rt),
				controller:   controllerName,
			}
		}
	}

	cfg.WrapTransport = func(rt http.RoundTripper) http.RoundTripper {
		return &controllerMetricsTripper{
			RoundTripper: rt,
			controller:   controllerName,
			observer:     observer,
		}
	}
}

// controllerMetricsTripper is a RoundTripper implementation which tracks our metrics for client requests.
type controllerMetricsTripper struct {
	http.RoundTripper
	controller string
	observer   *Observer
}

// RoundTrip implements the http RoundTripper interface. We simply call the wrapped RoundTripper
// and register the call with our apiCallCount metric.
func (cmt *controllerMetricsTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	// Call the nested RoundTripper.
	resp, err := cmt.RoundTripper.RoundTrip(req)

	// Count this call, if it worked (where "worked" includes HTTP errors).
	if err == nil {
		cmt.observer.observe(cmt.controller, req, resp, time.Since(start).Seconds())
	}

	return resp, err
}
