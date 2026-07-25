package report

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/LeoRBlume/loadtest-cli/internal/runner"
)

func TestPrint(t *testing.T) {
	res := runner.Result{
		TotalRequests: 1000,
		StatusCounts:  map[int]int{200: 980, 404: 10, 500: 5},
		Errors:        5,
	}

	var buf bytes.Buffer
	Print(&buf, "http://google.com", res, 1234*time.Millisecond)
	out := buf.String()

	wantLines := []string{
		"==== Load Test Report ====",
		"URL:                http://google.com",
		"Total time:         1.234s",
		"Total requests:     1000",
		"HTTP 200:           980",
		"Successful (2xx):   980",
		"Errors (network):   5",
		"Status code distribution:",
		"200 ....... 980",
		"404 ....... 10",
		"500 ....... 5",
		"errors .... 5",
	}
	for _, want := range wantLines {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n---\n%s", want, out)
		}
	}
}

func TestPrint_NoErrorsOmitsErrorLine(t *testing.T) {
	res := runner.Result{
		TotalRequests: 3,
		StatusCounts:  map[int]int{200: 3},
		Errors:        0,
	}

	var buf bytes.Buffer
	Print(&buf, "http://x", res, time.Second)
	out := buf.String()

	if strings.Contains(out, "errors ....") {
		t.Errorf("expected no error line when Errors=0, got:\n%s", out)
	}
}
