#!/bin/bash

__dirname=$(dirname "$(readlink -f "$0")")

docker run -d --rm --name redis6 -v $__dirname"/redis.conf":/etc/redis/redis.conf -p 6379:6379 redis:6-alpine redis-server /etc/redis/redis.conf