#!/bin/bash

DIR="$( cd "$( dirname "$BASH_SOURCE[0]" )" && pwd )"
cd "$DIR"

DATE=`date +%Y%m%d`
IMG_BASE_NAME=sempr/goscreenshot
IMG_NEW_NAME=$IMG_BASE_NAME:$DATE
IMG_LATEST_NAME=$IMG_BASE_NAME:latest

docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"
docker build -t $IMG_NEW_NAME .
docker tag $IMG_NEW_NAME $IMG_LATEST_NAME
docker push $IMG_NEW_NAME
docker push $IMG_LATEST_NAME
