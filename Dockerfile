# ---- build stage ----
FROM golang:1.26 AS build

WORKDIR /src

# Cache de dependências.
COPY go.mod go.sum ./
RUN go mod download

# Código-fonte e build estático.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /loadtest-cli .

# ---- final stage ----
FROM scratch

# Certificados para conexões HTTPS.
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /loadtest-cli /loadtest-cli

ENTRYPOINT ["/loadtest-cli"]
