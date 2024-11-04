FROM golang:1.23 AS build

RUN apt-get update && \
    apt-get install -y --no-install-suggests build-essential git cmake  libsdl2-dev clang \
    && rm -rf /var/lib/apt/lists/* /var/cache/apt/archives/*

WORKDIR /whisper.cpp

RUN git clone https://github.com/ggerganov/whisper.cpp.git .

# RUN WHISPER_SDL2=ON  make libwhisper.a

RUN  cmake -B build -DWHISPER_SDL2=on && cmake --build build --target whisper 

ENV LIBRARY_PATH=/whisper.cpp:/whisper.cpp/build/src:/whisper.cpp/build/ggml/src
ENV C_INCLUDE_PATH=/whisper.cpp/include:/whisper.cpp/ggml/include

ENV GOSUMDB=off
WORKDIR /app

COPY . .

RUN go mod download
RUN go build -tags localWhisper  -o app ./cmd/


FROM ubuntu:22.04

ENV GIN_MODE=release

RUN  apt-get update  \
    && apt-get install -y  --no-install-suggests ca-certificates  ffmpeg  \
    && rm -rf /var/lib/apt/lists/* /var/cache/apt/archives/*

WORKDIR /app

COPY --from=build /app/app .
COPY --from=build /whisper.cpp/build/ggml/src/libggml.so ./libggml.so
COPY --from=build /whisper.cpp/build/src/libwhisper.so.1.7.1 ./libwhisper.so
RUN ln -s /app/libwhisper.so /app/libwhisper.so.1

ENV PATH="/app:${PATH}"

ADD resource  resource

COPY resource/config.sample.yaml config.yaml 

CMD  [ "./app"]
