FROM golang:1.13

WORKDIR /go/src/inreach2aprs
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

ENV MAPSHARE_ID=""
ENV MAPSHARE_INTERVAL=60
ENV APRS_HOST="euro.aprs2.net"
ENV APRS_USER=""
ENV APRS_PASSCODE=""

CMD ["inreach2aprs"]