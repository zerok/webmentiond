# Changes

## Next

- Support for STARTTLS using `EMAIL_USE_STARTTLS` environment variable
- Add `SERVER_AUTH_JWT_SECRET` environment variable ([\#43][gh43])
- Improve detection of likes and comments (in nested `h-like`s)
- JWT secret can now be configured via an environment variable
- Expose metrics only if a `--metrics-addr` is set

[gh43]: https://github.com/zerok/webmentiond/issues/43

## 1.0.0 (2021-04-04)

Initial release
