FROM apache/apisix:3.6.0-debian

USER root

RUN apt-get update && apt-get install -y sudo && apt-get install -y procps

USER apisix

COPY ./go-runner /usr/local/apisix/apisix-go-plugin-runner/go-runner
COPY ./resource /usr/local/apisix/apisix-go-plugin-runner/resource