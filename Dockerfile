FROM golang:1.13-alpine AS gobuilder
RUN apk add --no-cache gcc libc-dev git
COPY . /src
WORKDIR /src/cmd/webmentiond
RUN go build

FROM node:12-alpine AS nodebuilder
COPY frontend /src/frontend
WORKDIR /src/frontend
RUN yarn && yarn run webpack

FROM alpine:3.11
VOLUME ["/data"]
RUN mkdir -p /var/lib/webmentiond/frontend
COPY pkg/server/migrations /var/lib/webmentiond/migrations
COPY --from=gobuilder /src/cmd/webmentiond/webmentiond /usr/local/bin/
COPY --from=nodebuilder /src/frontend/dist /var/lib/webmentiond/frontend/dist
COPY --from=nodebuilder /src/frontend/css /var/lib/webmentiond/frontend/css
COPY --from=nodebuilder /src/frontend/fontawesome /var/lib/webmentiond/frontend/fontawesome
COPY --from=nodebuilder /src/frontend/index.html /var/lib/webmentiond/frontend/
WORKDIR /var/lib/webmentiond
ENTRYPOINT ["/usr/local/bin/webmentiond", "serve", "--database-migrations", "/var/lib/webmentiond/migrations", "--database", "/data/webmentiond.sqlite"]
