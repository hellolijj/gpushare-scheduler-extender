FROM golang:1.10-stretch as build

WORKDIR /go/src/github.com/AliyunContainerService/gpushare-scheduler-extender
COPY . .

RUN go build -o /go/bin/gputopology-schd-extender cmd/*.go

FROM debian:stretch-slim

COPY --from=build /go/bin/gputopology-schd-extender /usr/bin/gputopology-schd-extender

CMD ["gputopology-schdextender","-logtostderr"]