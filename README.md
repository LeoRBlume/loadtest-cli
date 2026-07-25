# loadtest-cli

CLI em Go para testes de carga em serviços web. Dispara um número configurável
de requisições HTTP GET com um nível de concorrência definido e apresenta um
relatório detalhado da execução.

## Parâmetros

| Flag            | Obrigatório | Default | Descrição                     |
|-----------------|-------------|---------|-------------------------------|
| `--url`         | sim         | —       | URL do serviço a ser testado  |
| `--requests`    | sim         | —       | Número total de requisições   |
| `--concurrency` | sim         | —       | Número de chamadas simultâneas|
| `--timeout`     | não         | `30s`   | Timeout por requisição HTTP   |

## Execução local

```bash
go run . --url=http://google.com --requests=1000 --concurrency=10
```

Ou compilando o binário:

```bash
go build -o loadtest-cli .
./loadtest-cli --url=http://google.com --requests=1000 --concurrency=10
```

## Execução via Docker

Build da imagem:

```bash
docker build -t loadtest-cli .
```

Execução:

```bash
docker run loadtest-cli --url=http://google.com --requests=1000 --concurrency=10
```

## Exemplo de relatório

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
  500 .......   5
  errors ....   5
```

- **HTTP 200** — quantidade de respostas exatamente com status 200.
- **Successful (2xx)** — total de respostas com status na faixa 2xx.
- **Errors (network)** — falhas antes de obter um status HTTP (timeout, DNS,
  conexão recusada). Contam para o total de requisições.

## Arquitetura

```
main.go                    entrypoint → cmd.Execute()
cmd/root.go                CLI (Cobra): flags, validação, wiring
internal/runner/runner.go  worker pool + agregação de resultados
internal/report/report.go  formatação do relatório
```

O `runner` usa um pool de `--concurrency` goroutines consumindo de um canal com
`--requests` jobs, garantindo **exatamente** o total de requisições solicitado,
com no máximo `--concurrency` em voo simultaneamente.

## Testes

```bash
go test ./...
```
