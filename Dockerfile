# Image page: <https://hub.docker.com/_/golang>
FROM golang:1.13-alpine as builder

# UPX parameters help: <https://www.mankier.com/1/upx>
ARG upx_params
ENV upx_params=${upx_params:--7}

RUN apk add --no-cache upx

ADD ./src /src

WORKDIR /src

RUN set -x \
    && upx -V \
    && go version \
    && go build -ldflags='-s -w' -o /tmp/http-proxy-daemon . \
    && upx ${upx_params} /tmp/http-proxy-daemon \
    && /tmp/http-proxy-daemon -V \
    && /tmp/http-proxy-daemon -h

FROM alpine:latest

LABEL \
    org.label-schema.name="http-proxy-daemon" \
    org.label-schema.description="Docker image with http proxy daemon" \
    org.label-schema.url="https://github.com/tarampampam/http-proxy-daemon" \
    org.label-schema.vcs-url="https://github.com/tarampampam/http-proxy-daemon" \
    org.label-schema.vendor="Tarampampam" \
    org.label-schema.schema-version="1.0"

COPY --from=builder /tmp/http-proxy-daemon /bin/http-proxy-daemon

EXPOSE 8080

ENTRYPOINT ["/bin/http-proxy-daemon"]
