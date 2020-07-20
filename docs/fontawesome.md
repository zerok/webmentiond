# FontAwesome

Webmentiond supports FontAwesome Pro icons (5.x) for some UI elements inside the admin backend. Because FontAwesome has a proprietary license, though, we cannot ship it with webmentiond. In order to still see those fancy icons in the UI you have to do the following:

1. Download the Pro package for web from [fontawesome.com/download](https://fontawesome.com/download).
2. Extract and upload it to your server and put it into a folder where your webserver (e.g. nginx, Caddy, Apache) can reach it.
3. The last step depends on your configuration. If you serve the admin backend through `https://domain.com/webmentions/ui` then instruct your webserver to serve the FontAwesome folder through `https://domain.com/webmentions/ui/fontawesome`.

With Caddy this would mean, that you need to add something like this:

```
route /webmentions/ui/fontawesome/* {
    uri strip_prefix /webmentions/ui/fontawesome/
    root * /srv/www/path/to/fontawesome-pro-5.13.0-web/
    file_server
}
```