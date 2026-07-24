# loadtest-cli — Design

**Data:** 2026-07-23
**Módulo:** `github.com/LeoRBlume/loadtest-cli`
**Go:** 1.26

## Objetivo

CLI em Go para testes de carga em serviços web. Recebe URL, número total de
requisições e nível de concorrência; executa as requisições HTTP e imprime um
relatório detalhado no console. Empacotado em imagem Docker.

## Parâmetros (CLI — Cobra)

| Flag            | Tipo   | Obrigatório | Default | Descrição                          |
|-----------------|--------|-------------|---------|------------------------------------|
| `--url`         | string | sim         | —       | URL do serviço a testar            |
| `--requests`    | int    | sim         | —       | Número total de requisições        |
| `--concurrency` | int    | sim         | —       | Chamadas simultâneas               |
| `--timeout`     | duration | não       | `30s`   | Timeout por requisição HTTP        |

### Validação
- `--url` obrigatória e parseável (`url.ParseRequestURI`), com scheme http/https.
- `--requests` ≥ 1.
- `--concurrency` ≥ 1. Se `--concurrency > --requests`, ajusta para `--requests`.
- Erros de validação → mensagem no stderr + exit code ≠ 0.

## Arquitetura

```
loadtest-cli/
├── main.go                  # entrypoint → cmd.Execute()
├── cmd/
│   └── root.go              # Cobra: flags, validação, wiring
├── internal/
│   ├── runner/
│   │   ├── runner.go        # worker pool + agregação
│   │   └── runner_test.go
│   └── report/
│       ├── report.go        # formatação do relatório
│       └── report_test.go
├── Dockerfile
├── .dockerignore
├── go.mod / go.sum
└── README.md
```

### Responsabilidades
- **cmd** — parsing/validação de flags e wiring. Sem lógica de negócio.
- **runner** — recebe `Config` e `*http.Client`, executa e retorna `Result`.
  Injeção do client permite testar contra `httptest.Server`.
- **report** — recebe `Result` + duração e imprime relatório. Isolado e testável.

### Tipos

```go
// runner
type Config struct {
    URL         string
    Requests    int
    Concurrency int
}

type Result struct {
    TotalRequests int
    StatusCounts  map[int]int // código HTTP → contagem
    Errors        int         // falhas de rede (sem status HTTP)
}
```

## Modelo de concorrência

Worker pool:
1. Canal `jobs` recebe `Requests` sinais.
2. `Concurrency` goroutines consomem `jobs`, cada uma faz `GET` na URL.
3. Resultado de cada request atualiza contadores protegidos por `sync.Mutex`
   (ou canal de resultados agregado por uma goroutine coletora).
4. `sync.WaitGroup` aguarda o término.

Garante **exatamente `Requests`** requisições, com no máximo `Concurrency` em voo.

### Classificação de resposta
- Resposta HTTP obtida → incrementa `StatusCounts[statusCode]`.
- Erro de rede (timeout, DNS, conexão recusada) → incrementa `Errors`.
- Sucesso = qualquer **2xx**. A contagem de **200** é reportada explicitamente
  (exigência do enunciado) além do resumo 2xx.
- `resp.Body` é lido/descartado (`io.Copy(io.Discard, ...)`) e fechado para
  reutilizar conexões keep-alive.

## Relatório (stdout)

```
==== Load Test Report ====
URL:                http://google.com
Total time:         1.234s
Total requests:     1000
HTTP 200:           980
Successful (2xx):   985
Errors (network):   5

Status code distribution:
  200 ....... 980
  201 .......   5
  404 .......  10
  errors ....   5
```

- `Total time`: medido com `time.Since(start)` em torno da execução do pool.
- Distribuição ordenada por código; `errors` listado ao final quando > 0.

## Docker

Multi-stage:
- Stage build: `golang:1.26` → `CGO_ENABLED=0 go build` (binário estático).
- Stage final: `scratch` + `ca-certificates.crt` copiado do stage de build
  (necessário para HTTPS).
- `ENTRYPOINT ["/loadtest-cli"]` para que:
  ```
  docker run <img> --url=http://google.com --requests=1000 --concurrency=10
  ```
  funcione diretamente.

`.dockerignore` exclui `.git`, `.idea`, docs, README do contexto de build.

## Testes

- **runner_test.go**
  - `httptest.Server` retornando 200 → total exato e `StatusCounts[200]` corretos.
  - Server alternando 200/404/500 → contagem por status.
  - `concurrency > requests` → ainda executa exatamente `requests`.
  - Server que fecha a conexão / URL inválida → incrementa `Errors`.
- **report_test.go**
  - `Result` fixo → verifica strings do relatório (contém linhas esperadas).

## Fora de escopo (YAGNI)

- Métodos HTTP além de GET; corpo/headers customizados.
- Percentis de latência, saída em JSON, rampas de carga.
- Retry/backoff.
