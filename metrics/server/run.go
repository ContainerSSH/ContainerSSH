package server

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
)

func (s *MetricsServer) Handle(writer http.ResponseWriter, _ *http.Request) {
	var buffer bytes.Buffer
	for _, metricName := range s.collector.GetMetricNames() {
		help := s.collector.GetHelp(metricName)
		if help != "" {
			buffer.Write([]byte(fmt.Sprintf("# HELP %s %s\n", metricName, strings.ReplaceAll(help, "\n", ""))))
		}
		t := s.collector.GetType(metricName)
		if t != "" {
			buffer.Write([]byte(fmt.Sprintf("# TYPE %s %s\n", metricName, t)))
		}
		for metric, value := range s.collector.GetMetrics(metricName) {
			buffer.Write([]byte(fmt.Sprintf("%s %f\n", metric, value)))
		}
	}
	_, err := writer.Write(buffer.Bytes())
	if err != nil {
		s.logger.NoticeF("failed to write metrics output (%v)", err)
	}
}

func (s *MetricsServer) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if request.URL.Path == s.config.Path {
		s.Handle(response, request)
	} else {
		response.WriteHeader(404)
		_, _ = response.Write([]byte("Not found"))
	}
}

func (s *MetricsServer) Run(ctx context.Context) error {
	s.logger.InfoF("starting metrics server on %s", s.config.Listen)
	server := &http.Server{Addr: s.config.Listen, Handler: s}
	errorChannel := make(chan error, 1)
	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errorChannel <- err
			return
		}
		errorChannel <- nil
	}()

	select {
	case err := <-errorChannel:
		return err
	case <-ctx.Done():
		return server.Shutdown(ctx)
	}
}
