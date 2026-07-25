// Package cmd define a interface de linha de comando (Cobra).
package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/LeoRBlume/loadtest-cli/internal/report"
	"github.com/LeoRBlume/loadtest-cli/internal/runner"
	"github.com/spf13/cobra"
)

var (
	flagURL         string
	flagRequests    int
	flagConcurrency int
	flagTimeout     time.Duration
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "loadtest-cli",
		Short:         "CLI para testes de carga em serviços web",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          run,
	}

	cmd.Flags().StringVar(&flagURL, "url", "", "URL do serviço a ser testado (obrigatório)")
	cmd.Flags().IntVar(&flagRequests, "requests", 0, "número total de requisições (obrigatório)")
	cmd.Flags().IntVar(&flagConcurrency, "concurrency", 0, "número de chamadas simultâneas (obrigatório)")
	cmd.Flags().DurationVar(&flagTimeout, "timeout", 30*time.Second, "timeout por requisição HTTP")

	return cmd
}

func run(cmd *cobra.Command, _ []string) error {
	if err := validate(); err != nil {
		return err
	}

	concurrency := flagConcurrency
	if concurrency > flagRequests {
		concurrency = flagRequests
	}

	client := &http.Client{Timeout: flagTimeout}
	cfg := runner.Config{
		URL:         flagURL,
		Requests:    flagRequests,
		Concurrency: concurrency,
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Iniciando teste de carga: %d requisições, %d simultâneas → %s\n\n",
		flagRequests, concurrency, flagURL)

	start := time.Now()
	res := runner.Run(context.Background(), client, cfg)
	elapsed := time.Since(start)

	report.Print(cmd.OutOrStdout(), flagURL, res, elapsed)
	return nil
}

func validate() error {
	if flagURL == "" {
		return fmt.Errorf("--url é obrigatório")
	}
	u, err := url.ParseRequestURI(flagURL)
	if err != nil {
		return fmt.Errorf("--url inválida: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("--url deve usar http ou https, recebido %q", u.Scheme)
	}
	if flagRequests < 1 {
		return fmt.Errorf("--requests deve ser >= 1")
	}
	if flagConcurrency < 1 {
		return fmt.Errorf("--concurrency deve ser >= 1")
	}
	return nil
}

// Execute é o ponto de entrada da CLI.
func Execute() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
}
