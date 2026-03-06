FROM golang:1.26-bookworm AS builder

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /out/intake ./cmd/intake

FROM debian:bookworm-slim

RUN useradd -m intakeuser
USER intakeuser

WORKDIR /
COPY --from=builder /out/intake /intake

EXPOSE 8080

ENTRYPOINT ["/intake"]
