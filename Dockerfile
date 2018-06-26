FROM golang:1.10.2 as builder
WORKDIR /
RUN go get -d -v github.com/eug48/fhir-server
WORKDIR /go/src/github.com/eug48/fhir-server
RUN CGO_ENABLED=0 GOOS=linux go build
RUN cp fhir-server /

FROM alpine:3.7
RUN apk add --no-cache ca-certificates tini
COPY --from=builder /fhir-server /
CMD ["/fhir-server"]