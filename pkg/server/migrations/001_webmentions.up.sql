create table if not exists webmentions (
       id text primary key not null,
       source text not null,
       target text not null,
       created_at text not null,
       status text not null default "new"
);

create unique index webmentions_source_target on webmentions(source, target);
