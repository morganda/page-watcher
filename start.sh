#!/bin/bash

set -e

configfile=$1
container=page-watcher
image=${container}:0.1

if docker ps -a | grep -q "${container}"; then
	echo -n "Stopping running instance of ${container} first... "
	docker stop $container 2>&1 > /dev/null || true
	docker rm $container 2>&1 > /dev/null
	echo "done"
fi

echo -n "Starting... "
docker run -d --name=${container} -v $(pwd)/${configfile}:/etc/page-watcher/$(basename $configfile) $image 2>&1 > /dev/null
echo "done"
