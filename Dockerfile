FROM golang:1.15-alpine3.12 as builder
WORKDIR /go/src/git-semver
COPY . .
RUN go build -o ./bin/git-semver .

FROM alpine:3.12
RUN apk add --update --no-cache git
COPY --from=builder /go/src/git-semver/bin/git-semver /usr/local/bin/
ENTRYPOINT ["git-semver"]
