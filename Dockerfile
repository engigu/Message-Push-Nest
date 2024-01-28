FROM debian:latest

ENV TZ=Asia/Shanghai

RUN apt update \
        && apt install -y ca-certificates \
        && update-ca-certificates \
        && rm -rf /var/lib/apt/lists/*

COPY ./ /app/

WORKDIR /app
EXPOSE 8000

CMD ["/app/Message-Nest"]

