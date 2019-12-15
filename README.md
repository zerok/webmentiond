# Webmention server

This repository contains the source code for the Webmention
backend used at <https://zerokspot.com>.

## Authentication

1. Go to /login
2. Enter e-mail
3. Token-link sent to e-mail
4. Redirect to /authenticate page which sets a cookie and marks the
   user as logged in.

One authenticated the user can access the /admin area.
