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
LABEL Description="Docker image with http proxy daemon" Vendor="Tarampampam"

COPY --from=builder /tmp/http-proxy-daemon /bin/http-proxy-daemon

ENTRYPOINT ["/bin/http-proxy-daemon"]
