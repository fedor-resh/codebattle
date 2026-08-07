# syntax=docker/dockerfile:1
FROM golang:1.26.5-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY db ./db
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/problem-seed ./cmd/problem-seed

FROM golang:1.26.5-alpine
WORKDIR /app
COPY --from=build /out/problem-seed /usr/local/bin/problem-seed
COPY problems ./problems
ENTRYPOINT ["problem-seed", "-dir", "/app/problems"]
