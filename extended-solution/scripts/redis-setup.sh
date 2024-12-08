#!/bin/bash

docker run --rm -d \
    -p 6379:6379 \
    --name redis-server \
    redis:latest