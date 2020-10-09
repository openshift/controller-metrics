# pkg/apicall

Metrics and wrappers for counting external API calls from your controllers.

Currently instruments only kube API calls made through the controller-runtime Client.

- [pkg/apicall](#pkgapicall)
  - [Metric: `{prefix}_api_request_duration`](#metric-prefix_api_request_duration)
  - [Usage](#usage)
    - [Setup](#setup)
    - [Measuring Kube API Calls](#measuring-kube-api-calls)
    - [Scraping](#scraping)

## Metric: `{prefix}_api_request_duration`

**Type:** [HistogramVec](https://godoc.org/github.com/prometheus/client_golang/prometheus#HistogramVec)

**Labels:**
- `name` (constant): The name of your operator. Corresponds to the `operatorName` parameter to
  `NewAPICallObserver`.
- `controller`: The name of the controller making the API request. Corresponds to the
  `controller` argument to `NewCRClientWithMetrics`.
- `method`: The HTTP method of the observed API request. E.g. `GET`, `PUT`, etc.
- `resource`: A normalized representation of the API request URL.
  - Core kubernetes resources are prefixed with `core/`; custom resources are prefixed with their
    group name.
  - Unique namespace and resource names are replaced with the constants `{NAMESPACE}` and
    `{NAME}`, respectively.
- `status`: The HTTP response status code and name. E.g. `200 OK`, `409 Conflict`, etc.

`{prefix}` is the `operatorName` passed into `NewAPICallObserver`, normalized to replace hyphens with underscores.
For example, an `operatorName` of `my-operator` would produce metrics named `my_operator_api_request_duration*`.

## Usage

### Setup

- Import me

```go
import "github.com/openshift/controller-metrics/pkg/apicall"
```

- Create an Observer. You should only need one per executable.

```go
var APICallObserver = apicall.NewAPICallObserver("my-operator")
```

- The Observer includes a Collector. Register it with Prometheus.

```go
prometheus.Register(APICallObserver.Collector)
```

### Measuring Kube API Calls

- In each controller, where you have an implementation of `Reconciler` in the usual shape:

```go
// ReconcileFoo reconciles a Foo object
type ReconcileFoo struct {
	client    client.Client
	scheme    *runtime.Scheme
	...
}
```

...use `NewCRClientWithMetrics` to produce the `client`.
Pass in the name of your controller, and the Observer you created above.

```go
client, err := apicall.NewCRClientWithMetrics(mgr, "wizbang-controller", APICallObserver)
if err != nil {
    log.Error("Couldn't initialize metrics-wrapped client.")
    os.Exit(1)
}

reconciler := &ReconcileFoo{
    client: client,
    scheme: mgr.GetScheme(),
    ...
}
...
```

### Scraping

Using the example above, the metrics of interest would be:
- `my_operator_api_request_duration_sum`: The accumulated round trip time taken by each API call.
- `my_operator_api_request_duration_count`: The number of API calls.
