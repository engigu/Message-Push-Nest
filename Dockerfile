FROM debian:stable
ENV TZ=Asia/Shanghai

COPY ./ /app/

WORKDIR /app
EXPOSE 8000

CMD ["/app/Message-Nest"]

