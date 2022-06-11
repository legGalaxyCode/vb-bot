FROM alpine:latest

RUN mkdir /app

COPY /build/authApp /app

CMD ["/app/authApp"]