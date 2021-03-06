FROM golang:1.13-alpine

WORKDIR /go/src/inreach2aprs
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

ENV MAPSHARE_ID=""
ENV MAPSHARE_PASSWORD=""
ENV MAPSHARE_INTERVAL=60
ENV APRS_HOST="euro.aprs2.net"
ENV APRS_USER=""
ENV APRS_PASSCODE=""
ENV APRS_DEFAULT_CALLSIGN=""
ENV APRS_DEFAULT_COMMENT=""
ENV APRS_DEFAULT_SYMBOL=""

CMD ["inreach2aprs"]