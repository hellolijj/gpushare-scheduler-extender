FROM golang:1.10-stretch as build

WORKDIR /go/src/github.com/AliyunContainerService/gpushare-scheduler-extender
COPY . .

RUN go build -o /go/bin/gputopology-sche-extender cmd/*.go

FROM debian:stretch-slim

COPY --from=build /go/bin/gputopology-sche-extender /usr/bin/gputopology-sche-extender