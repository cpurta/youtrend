FROM golang:latest

MAINTAINER Chris Purta cpurta@gmail.com

RUN apt-get update && \
    mkdir -p /app

ADD ./youtrend /app

ENTRYPOINT /app/youtrend
