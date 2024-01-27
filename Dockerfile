FROM ubuntu:latest

COPY ./ /app/

WORKDIR /app
EXPOSE 8000

CMD ["/app/Message-Nest"]

