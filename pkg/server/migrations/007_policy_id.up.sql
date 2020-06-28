CREATE TEMPORARY TABLE tmp_url_policies AS SELECT * FROM url_policies;
DROP TABLE url_policies;
CREATE TABLE url_policies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    url_pattern TEXT,
    policy TEXT,
    weight INT
);
INSERT INTO url_policies (url_pattern, policy, weight) SELECT url_pattern, policy, weight FROM tmp_url_policies;
DROP TABLE IF EXISTS tmp_url_policies;
