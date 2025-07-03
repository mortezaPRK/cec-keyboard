ARG BASE_IMAGE_VERSION
FROM debian:${BASE_IMAGE_VERSION} AS base

RUN apt update && \
    apt install -y libcec-dev libp8-platform-dev libudev-dev ca-certificates pkg-config gcc g++ && \
    rm -rf /var/lib/apt/lists/*

COPY --from=golang:1.24 /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"

FROM base AS builder
    
WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build .

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