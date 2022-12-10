#!/bin/bash

__dirname=$(dirname "$(readlink -f "$0")")

old=`pwd`
docker stack deploy -c $__dirname"/swarm.yaml" test
cd $old