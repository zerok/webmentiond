# Configuration

Webmentiond can be configured through a handful of flags and environment
variables. This page will give you an overview of all of them:


## Server settings

### `--addr ADDR` (flag)

This is the network address the server should listen to for requests.

Default: `127.0.0.1:8080`

### `--metrics-addr ADDR` (flag)

This is the network address the server should listen to for requests exposing
metrics in a Prometheus-compatible format.

If this flag is not set then *no metrics* are exposed. If it equals the value
for the `--addr` flag then the metrics are expose by the same HTTP server as
the request of the API.

Default: ``


### `--public-url URL` (flag)

Especially if you're running webmentiond behind a proxy like Caddy, nginx, or
Apache you need to tell webmentiond under what URL user's should be able to
access it. For instance: If you've instructed your proxy to forward requests
from `https://yoursite.com/webmentions` to webmentiond, set `--public-url
https://yoursite.com/webmentions`.

Default: `http://127.0.0.1:8080`


### `--ui-path PATH` (flag)

Webmentiond comes with a little administration UI. The binary package as well
as the source repository contain a folder called `frontend` that contains this
UI. Webmentiond has to know where this folder is located on your system in
order to present it to you.

Default: `./frontend`


## Database settings

### `--database PATH` (flag)

Path to the file where webmentiond should store received mentions. Note that
webmentiond requires *write-access* to that file *and* the folder that contains
it.

Default: `./webmentiond.sqlite`


### `--database-migrations PATH` (flag)

As features are added or changed to/in webmentiond the database structure has
to change as well. To make these updates as pain-free as possible we ship a
folder with so-called "migrations". This folder is included in the binary
package inside the `pkg/server` folder.

Default: `./pkg/server/migrations`

## E-mail settings

Webmentiond requires an SMTP server to send you login email and also
notifications.

### `MAIL_HOST` (environment)

Hostname of the SMTP server, e.g.: `smtp.mailgun.org`.

### `MAIL_PORT` (environment)

Port the SMTP server accepts requests on, e.g.: `465`.

### `MAIL_USER` (environment)

E.g.: `postmaster@12345.mailgun.org`

### `MAIL_PASSWORD` (environment)
### `MAIL_FROM` (environment)

Through this setting you can define who should show up as the sender of mails.
E.g.: `yourname@yoursite.com`.

### `--send-notifications` (flag)

With this flag you can tell webmentiond to send out email notifications (e.g.
if a new mention has been verified).

Default: `false`


## Authentication settings

### `--auth-jwt-secret SECRET` (flag) / `SERVER_AUTH_JWT_SECRET` (environment)

When you log into the administration UI the server generates a little token for
you and signs it. The secret is necessary for that signing step. What the
secret looks like, though, is completely up to you. Just take any random but
long string (e.g. `52ba8240-b926-11ea-9e38-73b2d46d3547` or
`this-is-a-r3411y-long-PassPhrase-Th4t5_hard-to-GuEsS`).

### `--auth-jwt-ttl DURATION` (flag)

This value determines how long you can stay logged into the administration UI
without having to re-login.

Default: `7d` (7 days)

### `--auth-admin-emails EMAIL` (flag)

You log into the administration UI using your e-mail address. In order for the
system to know who should be able to log in, you have to specify their e-mail
address here.


## Using a configuration file

If you prefer to store your configuration inside a single configuration, you
can also do that. To get started, run `webmentiond config init` which will
generate a `webmentiond.yaml` file inside your current working directory. In
order for webmentiond to also load your settings from this file, pass
`--config-file webmentiond.yaml` to any webmentiond command.
