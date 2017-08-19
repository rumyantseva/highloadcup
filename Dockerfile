FROM busybox

COPY bin/travels /

EXPOSE 80

CMD ["/travels"]
