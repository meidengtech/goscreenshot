FROM golang:1.11 as builder

WORKDIR /go/src/github.com/sempr/goscreenshot/
RUN set -xe
RUN go get github.com/Masterminds/glide
COPY glide.* ./
RUN glide install
COPY cmd/ ./cmd
COPY constants/ ./constants
COPY pkg/ ./pkg
RUN go build -o /tmp/html2image github.com/sempr/goscreenshot/cmd/web

FROM sempr/chrome-headless:latest-notofont
ENV SCREENSHOT_CHROME_PATH /chrome/headless_shell
COPY --from=builder /tmp/html2image /usr/bin/html2image
ENTRYPOINT []
USER root
EXPOSE 8080
CMD /usr/bin/html2image
ENV SCREENSHOT_SERVER_PORT 8080
