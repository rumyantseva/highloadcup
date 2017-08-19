FROM busybox

COPY bin/travels /

RUN mkdir -p /tmp/data
COPY data/data.zip /tmp/data

EXPOSE 80

CMD ["/travels"]
