# Build the linux executable on a Golang container
FROM golang:latest
WORKDIR /src

COPY go.* .
COPY *.go .
RUN go build -o /threshold-metrics

ENV NEW_RELIC_ACCOUNT=""
ENV METRIC_NAME=""
ENV METRIC_WHERE=""
ENV METRIC_FACET=""
ENV THRESHOLD=""
ENV NEW_RELIC_LICENSE_KEY=""
ENV NEW_RELIC_USER_KEY=""

WORKDIR /

CMD ["/threshold-metrics"]
