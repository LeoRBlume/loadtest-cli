// Package report formata e imprime o relatório final do teste de carga.
package report

import (
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/LeoRBlume/loadtest-cli/internal/runner"
)

// Print escreve o relatório do resultado em w.
func Print(w io.Writer, url string, res runner.Result, elapsed time.Duration) {
	fmt.Fprintln(w, "==== Load Test Report ====")
	fmt.Fprintf(w, "URL:                %s\n", url)
	fmt.Fprintf(w, "Total time:         %s\n", elapsed.Round(time.Millisecond))
	fmt.Fprintf(w, "Total requests:     %d\n", res.TotalRequests)
	fmt.Fprintf(w, "HTTP 200:           %d\n", res.StatusCounts[200])
	fmt.Fprintf(w, "Successful (2xx):   %d\n", res.Successful2xx())
	fmt.Fprintf(w, "Errors (network):   %d\n", res.Errors)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Status code distribution:")

	codes := make([]int, 0, len(res.StatusCounts))
	for code := range res.StatusCounts {
		codes = append(codes, code)
	}
	sort.Ints(codes)
	for _, code := range codes {
		fmt.Fprintf(w, "  %-3d ....... %d\n", code, res.StatusCounts[code])
	}
	if res.Errors > 0 {
		fmt.Fprintf(w, "  errors .... %d\n", res.Errors)
	}
}
