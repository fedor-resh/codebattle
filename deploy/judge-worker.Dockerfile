# syntax=docker/dockerfile:1
FROM golang:1.26.5-alpine AS build
ARG TARGETOS
ARG TARGETARCH

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -ldflags="-s -w" -o /out/judge-worker ./cmd/judge-worker

FROM alpine:3.22
RUN apk add --no-cache docker-cli \
    && addgroup -S codebattle \
    && adduser -S -G codebattle codebattle \
    && mkdir -p /judge-source /judge-bin \
    && chown -R codebattle:codebattle /judge-source /judge-bin
COPY --from=build /out/judge-worker /usr/local/bin/judge-worker
EXPOSE 8081
ENTRYPOINT ["/usr/local/bin/judge-worker"]
