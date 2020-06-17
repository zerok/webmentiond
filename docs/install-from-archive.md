# Install webmentiond from an archive (Linux-only)

The default way to run webmentiond is using Docker. If you don't have the
option, you can also download a snapshot build which contains everything you
need to run webmentiond on Linux (amd64).

You can download the packaged version of webmentiond from the following URL
(after replacing $COMMITID with for instance
`25e7595472ddf0812407c5905bd881abc339b5fa`:

```
https://files.zerokspot.com/releases/webmentiond/snapshots/$COMMITID/webmentiond_$COMMITID_linux_amd64.zip
```

Once extracted, you should find a new folder with the same name as the zip
file. In there is the webmentiond binary, frontend files, and database
migrations. To start the server, run the following command:

```
./webmentiond serve \
  --database-migrations $PWD/pkg/server/migrations \
  --ui-path $PWD/frontend \
  $OTHER_OPTIONS
```

Please run `./webmentiond serve --help` to find out about all the other options that might be useful in your particular setup.
