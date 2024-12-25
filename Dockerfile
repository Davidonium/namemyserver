FROM node:20-slim AS frontend

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"

RUN corepack enable
COPY ./frontend /app

WORKDIR /app

RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --frozen-lockfile
RUN pnpm run build

FROM golang:1.23 AS builder

WORKDIR /app

COPY go.* ./

RUN go mod download

RUN go install github.com/a-h/templ/cmd/templ@latest

COPY ./cmd ./cmd
COPY ./internal ./internal
COPY ./db/ ./db
COPY --from=frontend /app/dist /app/frontend/dist
COPY embed.go .
COPY Makefile ./

RUN make build

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /
COPY --from=builder /app/build/namemyserver /namemyserver

USER 1001

ENTRYPOINT ["/namemyserver", "server"]
