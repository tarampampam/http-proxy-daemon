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
    && go build -ldflags='-s -w' -o /tmp/app . \
    && upx ${upx_params} /tmp/app \
    && /tmp/app -V \
    && /tmp/app -h

FROM alpine:latest
LABEL Description="Docker image with app" Vendor="Tarampampam"

COPY --from=builder /tmp/app /bin/app

ENTRYPOINT ["/bin/app"]
