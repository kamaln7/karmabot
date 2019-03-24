#!/bin/bash

cd "${0%/*}"

docker run \
	--rm -ti --privileged \
	-v $(pwd):/go/src/github.com/kamaln7/karmabot \
	-v /var/run/docker.sock:/var/run/docker.sock \
	-v $HOME/.config/goreleaser/github_token:/root/.config/goreleaser/github_token \
    -v $HOME/.docker/config.json:/root/.docker/config.json \
	karmabot-goreleaser $@
