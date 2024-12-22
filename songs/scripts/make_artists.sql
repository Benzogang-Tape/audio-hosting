DO $$
DECLARE
    hash_12345678 TEXT := '$2a$12$2tWCfH/XSY0sAE3kiy3wfegkJeuHDfyWxpCBUk2BH5hgspXDQaFLe';
    avatar_url TEXT := 'https://example.com';
BEGIN
-- inserting special artists
INSERT INTO users (id, name, email, password_hash, avatar_url) VALUES
('bc5f2d0f-1c29-42e6-b04a-aa3c43789d42', 'Estella Weber', 'verniemccullough@cole.org', hash_12345678, avatar_url),
('afe1d17c-fa31-4014-a3b0-77b8635c741d', 'Dustin Hilll', 'trishavolkman@wintheiser.name', hash_12345678, avatar_url),
('b1cb2661-0be5-4b38-bffd-604a7328ac90', 'Muriel Wilkinson', 'derickullrich@corwin.name', hash_12345678, avatar_url),
('e337e0ed-a629-4503-8249-c43cb5366aa1', 'Marilou Mohr', 'leannkunde@swift.net', hash_12345678, avatar_url);
-- inserting random artists
FOR i IN 1..1000 LOOP
    INSERT INTO users (id, name, email, password_hash, avatar_url)
    VALUES (gen_random_uuid(), 'artist' || i, 'artist' || i || '@artist.team02.com', hash_12345678, avatar_url);
END LOOP;
-- making them all artist with random labels
FOR i IN 1..100 LOOP
    INSERT INTO artists (label, user_id)
    SELECT (SELECT name FROM users ORDER BY random() LIMIT 1) AS label, id
    FROM users
    LIMIT 10
    OFFSET (i-1)*10;
END LOOP;
-- inserting usual users
FOR i IN 1..100000 LOOP
    INSERT INTO users (id, name, email, password_hash, avatar_url)
    VALUES (gen_random_uuid(), 'user' || i, 'user' || i || '@user.team02.com', hash_12345678, avatar_url);
END LOOP;
END; $$;