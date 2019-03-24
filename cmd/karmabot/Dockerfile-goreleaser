FROM alpine:3.6
RUN apk add --no-cache sqlite ca-certificates
COPY karmabot /
ENTRYPOINT ["/karmabot"]
