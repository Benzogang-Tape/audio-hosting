CREATE TABLE songs
(
  song_id        UUID       ,
  singer_fk      UUID        NOT NULL,
  name           VARCHAR(64) NOT NULL,
  s3_object_name VARCHAR(64) NOT NULL UNIQUE,
  s3_image_name  VARCHAR(64),
  duration       INTERVAL   ,
  weight_bytes   INT        ,
  released       BOOLEAN     NOT NULL,
  uploaded_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  released_at    TIMESTAMPTZ,
  PRIMARY KEY (song_id),
  UNIQUE(singer_fk, name)
);

CREATE TABLE feats
(
  song_fk   UUID NOT NULL REFERENCES songs(song_id),
  artist_fk UUID NOT NULL,
  PRIMARY KEY (song_fk, artist_fk)
);