package runner

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func newClient() *http.Client {
	return &http.Client{Timeout: 5 * time.Second}
}

func TestRun_AllSuccess(t *testing.T) {
	var hits int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&hits, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	res := Run(context.Background(), newClient(), Config{URL: srv.URL, Requests: 50, Concurrency: 10})

	if res.TotalRequests != 50 {
		t.Fatalf("TotalRequests = %d, want 50", res.TotalRequests)
	}
	if got := atomic.LoadInt64(&hits); got != 50 {
		t.Fatalf("server hits = %d, want 50", got)
	}
	if res.StatusCounts[200] != 50 {
		t.Fatalf("StatusCounts[200] = %d, want 50", res.StatusCounts[200])
	}
	if res.Successful2xx() != 50 {
		t.Fatalf("Successful2xx = %d, want 50", res.Successful2xx())
	}
	if res.Errors != 0 {
		t.Fatalf("Errors = %d, want 0", res.Errors)
	}
}

func TestRun_MixedStatuses(t *testing.T) {
	var n int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Alterna determinísticamente: a cada 3 requests → 200, 404, 500.
		switch atomic.AddInt64(&n, 1) % 3 {
		case 1:
			w.WriteHeader(http.StatusOK)
		case 2:
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	res := Run(context.Background(), newClient(), Config{URL: srv.URL, Requests: 30, Concurrency: 3})

	if res.TotalRequests != 30 {
		t.Fatalf("TotalRequests = %d, want 30", res.TotalRequests)
	}
	total := res.StatusCounts[200] + res.StatusCounts[404] + res.StatusCounts[500]
	if total != 30 {
		t.Fatalf("sum of status counts = %d, want 30", total)
	}
	if res.StatusCounts[200] != 10 || res.StatusCounts[404] != 10 || res.StatusCounts[500] != 10 {
		t.Fatalf("distribution = 200:%d 404:%d 500:%d, want 10/10/10",
			res.StatusCounts[200], res.StatusCounts[404], res.StatusCounts[500])
	}
}

func TestRun_ConcurrencyGreaterThanRequests(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	res := Run(context.Background(), newClient(), Config{URL: srv.URL, Requests: 5, Concurrency: 100})

	if res.TotalRequests != 5 {
		t.Fatalf("TotalRequests = %d, want 5", res.TotalRequests)
	}
	if res.StatusCounts[200] != 5 {
		t.Fatalf("StatusCounts[200] = %d, want 5", res.StatusCounts[200])
	}
}

func TestRun_NetworkErrors(t *testing.T) {
	// Servidor fechado: a conexão é recusada, gerando erros de rede.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close()

	res := Run(context.Background(), newClient(), Config{URL: url, Requests: 8, Concurrency: 4})

	if res.TotalRequests != 8 {
		t.Fatalf("TotalRequests = %d, want 8", res.TotalRequests)
	}
	if res.Errors != 8 {
		t.Fatalf("Errors = %d, want 8", res.Errors)
	}
	if len(res.StatusCounts) != 0 {
		t.Fatalf("StatusCounts = %v, want empty", res.StatusCounts)
	}
}
