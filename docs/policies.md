# Policies

Imagine you know some sites that regularly mention your content and that you
trust to not spam you. By default, these would still have to go through a
manual approval process before their links show up on your site. Policies allow
you to define, for instance, auto-approval for those sites so that their links
show up right away after the verification step.

Webmentiond stores such policies inside the `url_policies` table of the
database which has three primary properties:

1. The *URL pattern* (regular expression) for which a policy should apply.
2. A *policy* that should be executed if the given URL matches.
3. A *weight* in case multiple policies match a certain URL. The lower the
   weight, the earlier it is considered.

Webmentiond checks this table every couple of seconds for changed or new
policies and applies them to webmentions during the verification phase. Right
now, only a single policy is supported: `approve`. This means that a mention's
source that matches a policy's URL pattern and that passes verification is
automatically approved and does not require the administrator to manually
approve it.

At this point, policies cannot be configured through the UI yet. Instead, you
have to manually add policies to the database:

```
INSERT INTO url_policies VALUES
('^https://zerokspot.com/', 'approve', 1);
```
