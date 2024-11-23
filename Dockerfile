FROM golang:1.23 AS build

RUN apt-get update && \
    apt-get install -y --no-install-suggests build-essential git cmake  libsdl2-dev clang \
    && rm -rf /var/lib/apt/lists/* /var/cache/apt/archives/*

WORKDIR /whisper.cpp

RUN git clone https://github.com/ggerganov/whisper.cpp.git .

# RUN make WHISPER_SDL2=ON libwhisper.a

RUN cmake -B build -DWHISPER_SDL2=on && cmake --build build --target whisper

ENV LIBRARY_PATH=/whisper.cpp:/whisper.cpp/build/src:/whisper.cpp/build/ggml/src
ENV C_INCLUDE_PATH=/whisper.cpp/include:/whisper.cpp/ggml/include

ENV GOSUMDB=off
WORKDIR /app

COPY . .

RUN go mod download
RUN go build -tags localWhisper  -o app ./cmd/

RUN  mkdir -p /app/lib/  && \
 cp /go/pkg/mod/github.com/k2-fsa/sherpa-onnx-go-linux@*/lib/x86_64-unknown-linux-gnu/*.so /app/lib/ && \
 cp /whisper.cpp/build/ggml/src/*.so /app/lib && \ 
 cp /whisper.cpp/build/src/libwhisper.so.1.7.2 /app/lib
FROM ubuntu:22.04

ENV GIN_MODE=release

RUN  apt-get update  \
    && apt-get install -y  --no-install-suggests ca-certificates  ffmpeg  \
    && rm -rf /var/lib/apt/lists/* /var/cache/apt/archives/*

WORKDIR /app

COPY --from=build  /app/lib/ /app/
COPY --from=build /app/app .
# RUN ln -s /app/libwhisper.so /app/libwhisper.so.1

ENV LD_LIBRARY_PATH=/app
ENV PATH="/app:${PATH}"

ADD resource  resource

COPY resource/config.sample.yaml config.yaml 

CMD  [ "./app"]
