package main

import (
	"testing"
	"github.com/pact-foundation/pact-go/dsl"
	"github.com/pact-foundation/pact-go/types"
	"flag"
	"github.com/go-kit/kit/log"
	"os"
	"net/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	httptransport "github.com/go-kit/kit/transport/http"
	"context"
	"io/ioutil"
	"encoding/json"
)

func TestProvider(t *testing.T) {

	// Create Pact connecting to local Daemon
	pact := &dsl.Pact{
		Port:     6666, // Ensure this port matches the daemon port!
		Consumer: "MyConsumer",
		Provider: "MyProvider",
	}

	go startInstrumentedProvider()

	//// Start provider API in the background
	//go startServer()

	pact.VerifyProvider(t, types.VerifyRequest{
		ProviderBaseURL:        "http://localhost:8080",
		BrokerURL:              "https://pact.halfpipe.io",
		ProviderStatesSetupURL: "http://localhost:8080/setup",
		PublishVerificationResults: true,
		ProviderVersion:            "1.0.2",
	})
}

func startInstrumentedProvider() {
	var (
		listen = flag.String("listen", ":8080", "HTTP listen address")
		proxy  = flag.String("proxy", "", "Optional comma-separated list of URLs to proxy uppercase requests")
	)
	flag.Parse()

	var logger log.Logger
	logger = log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "listen", *listen, "caller", log.DefaultCaller)

	fieldKeys := []string{"method", "error"}
	requestCount := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "my_group",
		Subsystem: "string_service",
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, fieldKeys)
	requestLatency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "my_group",
		Subsystem: "string_service",
		Name:      "request_latency_microseconds",
		Help:      "Total duration of requests in microseconds.",
	}, fieldKeys)
	countResult := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "my_group",
		Subsystem: "string_service",
		Name:      "count_result",
		Help:      "The result of each count method.",
	}, []string{})

	var svc StringService
	svc = stringService{}
	svc = proxyingMiddleware(context.Background(), *proxy, logger)(svc)
	svc = loggingMiddleware(logger)(svc)
	svc = instrumentingMiddleware(requestCount, requestLatency, countResult)(svc)

	uppercaseHandler := httptransport.NewServer(
		makeUppercaseEndpoint(svc),
		decodeUppercaseRequest,
		encodeResponse,
	)
	countHandler := httptransport.NewServer(
		makeCountEndpoint(svc),
		decodeCountRequest,
		encodeResponse,
	)

	http.Handle("/uppercase", uppercaseHandler)
	http.Handle("/count", countHandler)
	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/setup", func(w http.ResponseWriter, req *http.Request) {
		logger.Log("[DEBUG] provider API: states setup")
		w.Header().Add("Content-Type", "application/json")

		var state types.ProviderState

		body, err := ioutil.ReadAll(req.Body)
		logger.Log(string(body))
		req.Body.Close()
		if err != nil {
			return
		}
		json.Unmarshal(body, &state)

		//svc := s.(*loggingMiddleware).next.(*userService)
		//
		//// Setup database for different states
		//if state.State == "User billy exists" {
		//	svc.userDatabase = billyExists
		//} else if state.State == "User billy is unauthorized" {
		//	svc.userDatabase = billyUnauthorized
		//} else {
		//	svc.userDatabase = billyDoesNotExist
		//}

		logger.Log("[DEBUG] configured provider state: ", state.State)
	})

	logger.Log("msg", "HTTP", "addr", *listen)
	logger.Log("err", http.ListenAndServe(*listen, nil))
}