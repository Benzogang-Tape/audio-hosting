CREATE TABLE playlists
(
    id          UUID         NOT NULL,
    title       VARCHAR(128) NOT NULL,
    author_id   UUID         NOT NULL,
    cover_url   VARCHAR(255),
    track_ids   UUID[]       DEFAULT '{}',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ,
    released_at TIMESTAMPTZ,
    is_album    BOOLEAN      NOT NULL DEFAULT False,
    is_public   BOOLEAN      NOT NULL DEFAULT False,
    PRIMARY KEY (id)
);

CREATE TABLE liked_playlists
(
    liked_playlist UUID        NOT NULL REFERENCES playlists(id) ON DELETE CASCADE,
    user_id        UUID        NOT NULL,
    liked_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (liked_playlist, user_id)
);

CREATE TABLE liked_tracks
(
    user_id  UUID        NOT NULL,
    track_id UUID        NOT NULL,
    liked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, track_id)
);