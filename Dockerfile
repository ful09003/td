FROM golang:latest AS build

RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64 && chmod +x /usr/local/bin/dep

RUN mkdir -p /go/src/github.com/ful09003/td
RUN git clone https://github.com/ful09003/td /go/src/github.com/ful09003/td/
WORKDIR /go/src/github.com/ful09003/td

COPY Gopkg.toml Gopkg.lock ./
# copies the Gopkg.toml and Gopkg.lock to WORKDIR

RUN dep ensure -vendor-only
# install the dependencies without checking for go code

WORKDIR /go/src/github.com/ful09003/td/cmd
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o td .


FROM alpine
RUN apk --no-cache add ca-certificates
WORKDIR /r
COPY --from=build /go/src/github.com/ful09003/td/cmd/td /r/
ENTRYPOINT ["./td"]