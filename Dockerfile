# Stage 1: Build dell'eseguibile
FROM golang:1.26.5-alpine AS builder

WORKDIR /app

# Copia i file delle dipendenze per sfruttare il caching dei layer di Docker
COPY go.mod go.sum ./
RUN go mod download

# Copia il resto del codice sorgente
COPY . .

# Compila il binario statico, ottimizzato e senza informazioni di debug non necessarie
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o shpl cmd/shpl/main.go

# Stage 2: Immagine finale di esecuzione
FROM alpine:3.19

WORKDIR /root/

# Copia il binario dallo stage di build
COPY --from=builder /app/shpl .

# Comando di default (mostra l'help della CLI)
ENTRYPOINT ["./shpl"]
CMD ["--help"]