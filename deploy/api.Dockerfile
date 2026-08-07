# syntax=docker/dockerfile:1
FROM golang:1.26.5-alpine AS build
ARG TARGETOS
ARG TARGETARCH

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY db ./db
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -ldflags="-s -w" -o /out/codebattle-api ./cmd/api

FROM alpine:3.22
RUN addgroup -S codebattle && adduser -S -G codebattle codebattle
COPY --from=build /out/codebattle-api /usr/local/bin/codebattle-api
USER codebattle
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/codebattle-api"]
