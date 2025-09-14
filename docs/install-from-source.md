# Install webmentiond from source (Linux or macOS)

The default way to run webmentiond is using Docker. If you don't have the
option, you can also build the whole application from source. For this you'll
need to have the following components installed:

- Go 1.24 or newer
- NodeJS 22 or newer


## Build the frontend

```
cd frontend
yarn
yarn run webpack --mode production
```

This will produce production JavaScript files that you can then serve using
your HTTP server or from the backend when using the `--ui-path` setting.
Additionally, the output is embedded into the server binary during the build
process in the next step.


## Build the backend

```
cd cmd/webmentiond
go build -o ../../webmentiond
```

This will produce the `webmentiond` binary file in the root folder of this
project.


## Running the build

Now you have a fully working build of the application including the frontend.
For actually executing it you will also set some environment variables and
flags that are documented in the getting-started guide and the configuration
reference.

