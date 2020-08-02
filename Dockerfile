FROM golang:1.14-alpine AS build
ENV GCO_ENABLED=0
RUN mkdir /app
ADD . /app
WORKDIR /app

RUN go build -o /bin/balancer main.go

FROM alpine:latest
COPY --from=build /bin/balancer /bin/balancer
COPY ./backends.txt /bin/backends.txt

ENTRYPOINT ["/bin/balancer", "-f", "/bin/backends.txt"]
