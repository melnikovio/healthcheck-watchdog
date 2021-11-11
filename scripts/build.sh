#!/usr/bin/env bash
tag=healthcheck-exporter:0.2.1

echo Building $tag
# docker build --no-cache -t ziiot-docker.dp.nlmk.com/digital-plant/$tag . -f ./Dockerfile
# docker push ziiot-docker.dp.nlmk.com/digital-plant/$tag
docker build --no-cache -t melnikovio/$tag . -f ./Dockerfile
docker push melnikovio/$tag