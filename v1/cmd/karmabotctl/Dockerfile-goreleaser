FROM alpine:3.6
RUN apk add --no-cache sqlite ca-certificates
COPY karmabotctl /
ENTRYPOINT ["/karmabotctl"]
