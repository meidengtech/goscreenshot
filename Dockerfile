FROM golang:1.8.3 as builder

WORKDIR /go/src/github.com/sempr/goscreenshot/
RUN set -xe
RUN go get github.com/Masterminds/glide
COPY cmd/ ./cmd
COPY constants/ ./constants
COPY pkg/ ./pkg
COPY glide.* ./
RUN glide install
RUN go build -o /tmp/html2image github.com/sempr/goscreenshot/cmd/web

FROM sempr/chrome-headless:62.0.3194.2-notofont
COPY --from=builder /tmp/html2image /usr/bin/html2image
ENTRYPOINT []
USER root
EXPOSE 8080
CMD /usr/bin/html2image
