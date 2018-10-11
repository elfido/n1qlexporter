FROM golang:1.11 as builder

COPY . $GOPATH/src/github.com/elfido/n1qlExporter
RUN mv $GOPATH/src/github.com/elfido/n1qlExporter/settings.json /etc/settings.json
WORKDIR $GOPATH/src/github.com/elfido/n1qlExporter
RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure -vendor-only
RUN GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build --installsuffix cgo --ldflags="-s"  -o /n1qlExporter

FROM alpine:3.8
RUN mkdir -p /app
COPY --from=builder /n1qlExporter /app/n1qlExporter
COPY --from=builder /etc/settings.json /app/settings.json
WORKDIR /app
EXPOSE 8380
ENTRYPOINT ["./n1qlExporter"]