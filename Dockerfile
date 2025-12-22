FROM docker.io/library/node:24-slim AS frontend

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"

RUN corepack enable
COPY ./frontend /app/frontend
COPY ./internal/templates/ /app/internal/templates

WORKDIR /app/frontend

RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --frozen-lockfile
RUN pnpm run build

FROM docker.io/library/golang:1.25 AS builder

WORKDIR /app

COPY go.* ./

RUN go mod download

COPY ./cmd ./cmd
COPY ./internal ./internal
COPY ./db/ ./db
COPY --from=frontend /app/frontend/dist /app/frontend/dist
COPY embed.go .
COPY please .

RUN ./please build

FROM gcr.io/distroless/static-debian13:nonroot

WORKDIR /app
COPY --from=builder /app/build/namemyserver /app/namemyserver

ENTRYPOINT ["/app/namemyserver", "server"]
