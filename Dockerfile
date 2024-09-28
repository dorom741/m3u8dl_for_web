FROM golang:1.22 AS build

# ENV GOPROXY=https://goproxy.cn,direct
# 设置工作目录
WORKDIR /app

# 将项目文件复制到容器中
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 go build -o app

FROM alpine:latest

ENV GIN_MODE=release
RUN set -ex  && apk add --no-cache ca-certificates bash

WORKDIR /app


COPY --from=build /app/app .

ADD resource  resource

COPY resource/config.sample.yaml config.yaml 

CMD  [ "./app"]
