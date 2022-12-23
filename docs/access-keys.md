# Authentication with Access Keys

For some setups it might make sense to handle admin operations programmatically.
One such situation might be that you want to retrieve a list of all mentions during CI *without* downloading the database from the server.

This is where access keys come in.
These statically configured tokens allow you retrieve a JWT from webmentiond without going through e-mail authentication.

## Setup

To set up access keys, start the webmentiond server with the following flag:

```
--auth-admin-access-keys 12345=ci
```

This adds the access key `12345` which represents a `ci` user.

## Authentication

Next, you need to authenticate against webmentiond using that key in order to retrieve a shortlived JWT which you can then use to fetch data from the `/manage/` API:

```hurl
# Request a JWT
POST http://localhost:8080/authenticate/access-key
[FormParams]
key: 12345

HTTP/1.1 200
[Captures]
jwt: body

# Retrieve a list of all mentions using the JWT
GET http://localhost:8080/manage/mentions
Authorization: Bearer {{jwt}}
```

## Security considerations

Access keys are static so if you're using them, please also consider putting the `authenticate/access-key` endpoint behind some IP restrictions or install Fail2ban to scan your reverse proxy's access logs for potential attacks.

