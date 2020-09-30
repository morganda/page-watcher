#!/bin/bash

[ -z $1 ] && version=0.1 || version=$1

docker build -t page-watcher:${version} .
