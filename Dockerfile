FROM node:20 AS frontend

WORKDIR /app

COPY ./frontend ./
RUN npm ci

FROM golang:1.23 AS builder

RUN useradd -u 1001 -m namemyserver

WORKDIR /app

COPY go.* ./

RUN go mod download

COPY ./cmd ./cmd
COPY ./internal ./internal
COPY embed.go .
COPY Makefile ./

COPY --from=frontend /app/dist /app/frontend/dist

RUN make build

FROM scratch

WORKDIR /
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /app/build/namemyserver /namemyserver

USER 1001

ENTRYPOINT ["/namemyserver"]
