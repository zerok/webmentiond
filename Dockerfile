# WARNING: At this point this file is no longer maintained. The Docker image is built within ci/package.go!
FROM node:22-alpine3.18 AS nodebuilder
COPY frontend /src/frontend
WORKDIR /src/frontend
RUN yarn && yarn run webpack --mode production

FROM golang:1.25-alpine AS gobuilder
RUN apk add --no-cache gcc libc-dev git sqlite-dev
COPY . /src
COPY --from=nodebuilder /src/frontend /src/
WORKDIR /src/cmd/webmentiond
RUN go build --tags "libsqlite3 linux"

FROM alpine:3.23
RUN apk add --no-cache sqlite-dev
VOLUME ["/data"]
RUN adduser -u 1500 -h /data -H -D webmentiond && \
    mkdir -p /var/lib/webmentiond/frontend
COPY pkg/server/migrations /var/lib/webmentiond/migrations
COPY --from=gobuilder /src/cmd/webmentiond/webmentiond /usr/local/bin/
WORKDIR /var/lib/webmentiond
USER 1500
ENTRYPOINT ["/usr/local/bin/webmentiond", "serve", "--database-migrations", "/var/lib/webmentiond/migrations", "--database", "/data/webmentiond.sqlite"]
