package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthAndVersionEndpoints(t *testing.T) {
	handler := newHandler("test-version")

	health := httptest.NewRecorder()
	handler.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if health.Code != http.StatusOK || strings.TrimSpace(health.Body.String()) != "ok" {
		t.Fatalf("health endpoint returned %d and %q", health.Code, health.Body.String())
	}

	version := httptest.NewRecorder()
	handler.ServeHTTP(version, httptest.NewRequest(http.MethodGet, "/version", nil))
	if version.Code != http.StatusOK || !strings.Contains(version.Body.String(), `"version":"test-version"`) {
		t.Fatalf("version endpoint returned %d and %q", version.Code, version.Body.String())
	}
}

func TestMetricsEndpointExposesRequestCounterAndHistogram(t *testing.T) {
	handler := newHandler("test")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	metrics := httptest.NewRecorder()
	handler.ServeHTTP(metrics, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := metrics.Body.String()
	if !strings.Contains(body, `fleetops_http_requests_total{method="GET",path="GET /healthz",status="200"}`) {
		t.Fatalf("request counter missing from metrics: %s", body)
	}
	if !strings.Contains(body, `fleetops_http_request_duration_seconds_bucket`) {
		t.Fatalf("latency histogram missing from metrics: %s", body)
	}
}
