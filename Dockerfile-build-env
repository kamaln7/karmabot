FROM golang:1.11-alpine

# Need to mount /var/run/docker.sock
# Need to mount /root/.config/goreleaser/github_token
# Need to mount /go/src/github.com/kamaln7/karmabot

RUN apk add --no-cache alpine-sdk git docker

RUN mkdir /tmp/goreleaser && \
    cd /tmp/goreleaser && \
    wget -O goreleaser.tgz https://github.com/goreleaser/goreleaser/releases/download/v0.104.0/goreleaser_Linux_x86_64.tar.gz && \
    tar vxf goreleaser.tgz && \
    mv goreleaser /bin && \
    rm -r /tmp/goreleaser

WORKDIR /go/src/github.com/kamaln7/karmabot
ENTRYPOINT ["goreleaser"]
