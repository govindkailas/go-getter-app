# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

FROM golang:1.21-alpine AS build
COPY  . /app
WORKDIR /app
RUN go build -o go-getter-app

FROM alpine:latest
COPY --from=build /app/go-getter-app .
EXPOSE 8080
ENTRYPOINT [ "./go-getter-app" ]

RUN apk add --no-cache bash curl

HEALTHCHECK \
    --start-period=1s \
    --interval=1s \
    --timeout=1s \
    --retries=30 \
        CMD curl --fail -s http://localhost:8080/healthcheck || exit 1