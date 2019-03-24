FROM alpine:3.6
RUN apk add --no-cache sqlite ca-certificates
COPY karmabot /
COPY www /www
EXPOSE 4000
ENV KB_WEBUI_LISTENADDR 0.0.0.0:4000
ENV KB_WEBUI_PATH /www
ENTRYPOINT ["/karmabot"]
