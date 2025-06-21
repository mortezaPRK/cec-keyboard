FROM golang:1.24 AS base

RUN apt update && \
    apt install -y libcec-dev libp8-platform-dev libudev-dev && \
    rm -rf /var/lib/apt/lists/*

    
FROM base AS builder
    
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build .

FROM base AS devenv

RUN apt update && \
    apt install -y sudo && \
    rm -rf /var/lib/apt/lists/*

RUN groupadd -r developer && \
    useradd -r -g developer developer && \
    echo "developer ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/developer && \
    mkdir -p /home/developer && \
    chown developer:developer /home/developer

USER developer

SHELL ["/bin/bash", "-c"]