FROM ubuntu:18.04
EXPOSE 8080/tcp
ENV DATA_DIR "/data"
USER nobody
COPY api /usr/local/search/api
ENTRYPOINT ["/usr/local/search/api"]