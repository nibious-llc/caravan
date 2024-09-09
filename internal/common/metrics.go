package common

import (
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_model/go"
	"github.com/prometheus/common/version"
	"github.com/prometheus/node_exporter/collector"
	"github.com/rs/zerolog/log"
	"net"
	"net/http"
	"sort"
)

var r *prometheus.Registry

type LoggerStub struct {
}

// This function is called each time a metric is requested from each
// processor. It will produce a lot of output if you add anything here
func (l LoggerStub) Log(keyvals ...interface{}) error {
	return nil
}

func (client ClientHub) Gather() ([]*io_prometheus_client.MetricFamily, error) {
	client.Writer <- generateMetricsRequestMsg()

	metrics, ok := <-client.MetricsChan

	if !ok {
		return nil, errors.New("Operation Cancelled")
	}

	return metrics, nil
}

func HandleMetrics(client *ClientHub) int {
	handler := promhttp.HandlerFor(
		client,
		promhttp.HandlerOpts{
			ErrorHandling: promhttp.ContinueOnError,
		},
	)

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}

	fmt.Println("Using port:", listener.Addr().(*net.TCPAddr).Port)

	sm := http.NewServeMux()
	sm.Handle("/metrics", handler)

	client.MetricsServer = &http.Server{Handler: sm}

	go func() {
		if err := client.MetricsServer.Serve(listener); err != nil {
			// handle err
			return
		}
	}()

	return listener.Addr().(*net.TCPAddr).Port
}

func initMetricsReport() bool {

	var l LoggerStub

	// NewCollector task a logger and a string for filters. Don't want to filter
	// anything right now
	nc, err := collector.NewNodeCollector(l)
	if err != nil {
		log.Error().Err(err).Msg("Could not create collector")
		return false
	}

	// Only log the creation of an unfiltered handler, which should happen
	// only once upon startup.
	collectors := []string{}
	for n := range nc.Collectors {
		collectors = append(collectors, n)
		log.Info().Msg("Adding Collectors")
		log.Info().Msg(n)
	}
	sort.Strings(collectors)
	for _, c := range collectors {
		log.Info().Msg(fmt.Sprintf("collector enabled: %s", c))
	}

	r = prometheus.NewRegistry()
	r.MustRegister(version.NewCollector("node_exporter"))
	if err := r.Register(nc); err != nil {
		log.Error().Err(err).Msg("couldn't register node collector")
		return false
	}
	return true
}

func generateMetricsRequestMsg() []byte {

	msg := Message{
		Type:    RequestMetricsMsgType,
		Content: nil,
	}

	d, err := MarshalObject(msg)
	if err != nil {
		log.Error().Err(err).Msg("Could not marshal object")
		return nil
	}

	return d
}

func generateMetricsReportMsg() []byte {

	metrics, gatherErr := r.Gather()
	if gatherErr != nil {
		log.Error().Err(gatherErr).Msg("couldn't gather metrics")
		return nil
	}

	m, metricsMarshalErr := MarshalObject(metrics)
	if metricsMarshalErr != nil {
		log.Error().Err(metricsMarshalErr).Msg("couldn't marhsal metrics")
		return nil
	}

	msg := Message{
		Type:    ResponseMetricsReportMsgType,
		Content: m,
	}

	d, err := MarshalObject(msg)
	if err != nil {
		log.Error().Err(err).Msg("Could not marshal object")
		return nil
	}

	return d
}
