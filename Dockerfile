FROM debian:buster-slim

RUN mkdir -p '/mnt/output' && \
    mkdir -p '/mnt/pkgs' && \
    mkdir -p '/build/src' && \
    apt-get update && \
    apt-get upgrade -y && \
    apt-get -y install live-build cpio wget ca-certificates apt-utils

COPY buildimage /usr/local/bin

WORKDIR /build/src

ENTRYPOINT /usr/local/bin/buildimage
