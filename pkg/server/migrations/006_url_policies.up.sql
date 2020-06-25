CREATE TABLE IF NOT EXISTS url_policies (
    url_pattern TEXT,
    policy TEXT,
    weight INTEGER
);

CREATE UNIQUE INDEX IF NOT EXISTS url_policies_key ON url_policies (url_pattern, policy);
