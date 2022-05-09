FROM golang:1.18.1 AS builder
LABEL AUTHOR Esther Kim (estherk@codebrick.co)

ADD . /go/src/github.com/codebrick-corp/aws-dms-task-exporter
WORKDIR /go/src/github.com/codebrick-corp/aws-dms-task-exporter

RUN go mod tidy && go mod vendor
RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/aws-dms-task-exporter .

FROM alpine:3.14
LABEL AUTHOR Esther Kim (estherk@codebrick.co)

COPY --chown=0:0 --from=builder /go/bin/aws-dms-task-exporter /bin/

EXPOSE 8080
ENTRYPOINT ["/bin/aws-dms-task-exporter"]