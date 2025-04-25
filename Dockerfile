# WARNING: At this point this file is no longer maintained. The Docker image is built within ci/package.go!
FROM golang:1.20-alpine AS gobuilder
RUN apk add --no-cache gcc libc-dev git sqlite-dev
COPY . /src
WORKDIR /src/cmd/webmentiond
RUN go build --tags "libsqlite3 linux"

FROM node:18-alpine3.18 AS nodebuilder
COPY frontend /src/frontend
WORKDIR /src/frontend
RUN yarn && yarn run webpack --mode production

FROM alpine:3.21
RUN apk add --no-cache sqlite-dev
VOLUME ["/data"]
RUN adduser -u 1500 -h /data -H -D webmentiond && \
    mkdir -p /var/lib/webmentiond/frontend
COPY pkg/server/migrations /var/lib/webmentiond/migrations
COPY --from=gobuilder /src/cmd/webmentiond/webmentiond /usr/local/bin/
COPY --from=nodebuilder /src/frontend/dist /var/lib/webmentiond/frontend/dist
COPY --from=nodebuilder /src/frontend/css /var/lib/webmentiond/frontend/css
COPY --from=nodebuilder /src/frontend/index.html /var/lib/webmentiond/frontend/
COPY --from=nodebuilder /src/frontend/demo.html /var/lib/webmentiond/frontend/
WORKDIR /var/lib/webmentiond
USER 1500
ENTRYPOINT ["/usr/local/bin/webmentiond", "serve", "--database-migrations", "/var/lib/webmentiond/migrations", "--database", "/data/webmentiond.sqlite"]
