#!/bin/bash

GIT_TAG="${GITHUB_REF##refs/tags/}"
GIT_BRANCH="${GITHUB_REF##refs/heads/}"
GIT_SHA=$(git rev-parse --short HEAD)

DOCKER_IMAGE_TAG="--tag ${DOCKER_IMAGE}:sha-${GIT_SHA}"

if [[ $GITHUB_REF != $GIT_TAG ]]; then
    DOCKER_IMAGE_TAG="${DOCKER_IMAGE_TAG} --tag ${DOCKER_IMAGE}:${GIT_TAG#v}  --tag ${DOCKER_IMAGE}:latest"
elif [[ $GITHUB_REF == "refs/heads/master" ]]; then
    DOCKER_IMAGE_TAG="${DOCKER_IMAGE_TAG} --tag ${DOCKER_IMAGE}:develop"
elif [[ $GIT_BRANCH = feature/* ]]; then
    DOCKER_IMAGE_TAG="${DOCKER_IMAGE_TAG} --tag ${DOCKER_IMAGE}:$(echo $GIT_BRANCH | tr / -)"
fi

echo ${DOCKER_IMAGE_TAG}