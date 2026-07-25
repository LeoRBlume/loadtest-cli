// Package runner executa o teste de carga: dispara N requisições HTTP GET
// respeitando o nível de concorrência e agrega os resultados.
package runner

import (
	"context"
	"io"
	"net/http"
	"sync"
)

// Config descreve os parâmetros de um teste de carga.
type Config struct {
	URL         string
	Requests    int
	Concurrency int
}

// Result agrega o resultado da execução.
type Result struct {
	TotalRequests int         // total de requisições efetivamente realizadas
	StatusCounts  map[int]int // código HTTP → quantidade
	Errors        int         // falhas de rede (sem status HTTP)
}

// Successful2xx retorna a quantidade de respostas com status 2xx.
func (r Result) Successful2xx() int {
	n := 0
	for code, count := range r.StatusCounts {
		if code >= 200 && code < 300 {
			n += count
		}
	}
	return n
}

// Run executa o teste de carga descrito por cfg usando o client fornecido.
// Garante que exatamente cfg.Requests requisições sejam realizadas, com no
// máximo cfg.Concurrency em voo simultaneamente.
func Run(ctx context.Context, client *http.Client, cfg Config) Result {
	concurrency := cfg.Concurrency
	if concurrency < 1 {
		concurrency = 1
	}
	if concurrency > cfg.Requests {
		concurrency = cfg.Requests
	}

	jobs := make(chan struct{})
	var mu sync.Mutex
	result := Result{StatusCounts: make(map[int]int)}

	record := func(status int, err error) {
		mu.Lock()
		defer mu.Unlock()
		result.TotalRequests++
		if err != nil {
			result.Errors++
			return
		}
		result.StatusCounts[status]++
	}

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				status, err := doRequest(ctx, client, cfg.URL)
				record(status, err)
			}
		}()
	}

	for i := 0; i < cfg.Requests; i++ {
		jobs <- struct{}{}
	}
	close(jobs)
	wg.Wait()

	return result
}

// doRequest faz um GET e retorna o status code. O corpo é descartado e fechado
// para permitir a reutilização de conexões keep-alive.
func doRequest(ctx context.Context, client *http.Client, rawURL string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return 0, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode, nil
}
