FROM golang:1.11 as builder

WORKDIR /code
RUN set -xe
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY cmd/ ./cmd
COPY constants/ ./constants
COPY pkg/ ./pkg
RUN GO111MODULE=on go build -a -ldflags '-extldflags "-static"' -o /tmp/html2image cmd/web/main.go

FROM sempr/chrome-headless:latest-notofont
ENV SCREENSHOT_CHROME_PATH /headless-shell/headless-shell
COPY --from=builder /tmp/html2image /usr/bin/html2image
ENTRYPOINT []
USER root
EXPOSE 8080
CMD /usr/bin/html2image
ENV SCREENSHOT_SERVER_PORT 8080
