-- name: SaveSong :exec
WITH inserted_song AS (
    INSERT INTO songs (
    song_id,
    singer_fk,
    name,
    s3_object_name,
    image_url,
    duration,
    weight_bytes,
    uploaded_at,
    released_at) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)
    RETURNING song_id
)
INSERT INTO feats (song_fk, artist_fk, order_num)
SELECT
    (SELECT song_id FROM inserted_song) AS song_fk,
    artist AS artist_fk,
    ROW_NUMBER() OVER () AS order_num
FROM UNNEST(@artists_ids::UUID[]) AS artist;

-- name: Song :one
SELECT
    sqlc.embed(songs),
    ARRAY_AGG(feats.artist_fk ORDER BY feats.order_num)::UUID[] AS artists_ids
FROM songs
LEFT JOIN feats ON feats.song_fk = songs.song_id
WHERE song_id = $1 AND released_at IS NOT NULL
GROUP BY songs.song_id;

-- name: UpdateSong :one
UPDATE songs SET
    singer_fk = $2,
    name = $3,
    s3_object_name = $4,
    image_url = $5,
    duration = $6,
    weight_bytes = $7,
    released_at = $8,
    uploaded_at = $9
WHERE song_id = $1
RETURNING *;

-- name: PatchSong :one
UPDATE songs SET
    singer_fk = COALESCE(sqlc.narg('singer_fk'), singer_fk),
    name = COALESCE(sqlc.narg('name'), name),
    s3_object_name = COALESCE(sqlc.narg('s3_object_name'), s3_object_name),
    image_url = COALESCE(sqlc.narg('image_url'), image_url),
    duration = COALESCE(sqlc.narg('duration'), duration),
    weight_bytes = COALESCE(sqlc.narg('weight_bytes'), weight_bytes),
    released_at = COALESCE(sqlc.narg('released_at'), released_at),
    uploaded_at = COALESCE(sqlc.narg('uploaded_at'), uploaded_at)
WHERE song_id = @id
RETURNING *;

-- name: PatchSongs :exec
UPDATE songs SET
    singer_fk = COALESCE(sqlc.narg('singer_fk'), singer_fk),
    name = COALESCE(sqlc.narg('name'), name),
    s3_object_name = COALESCE(sqlc.narg('s3_object_name'), s3_object_name),
    image_url = COALESCE(sqlc.narg('image_url'), image_url),
    duration = COALESCE(sqlc.narg('duration'), duration),
    weight_bytes = COALESCE(sqlc.narg('weight_bytes'), weight_bytes),
    released_at = COALESCE(sqlc.narg('released_at'), released_at),
    uploaded_at = COALESCE(sqlc.narg('uploaded_at'), uploaded_at)
WHERE song_id = ANY(@ids::UUID[]);

-- name: DeleteSongs :exec
DELETE FROM songs WHERE song_id = ANY(@ids::UUID[]);

-- name: ReleasedSongs :many
SELECT
    sqlc.embed(songs),
    ARRAY_AGG(feats.artist_fk ORDER BY feats.order_num)::UUID[] AS artists_ids
FROM songs
LEFT JOIN feats ON feats.song_fk = songs.song_id
WHERE
    -- Only released songs!
    released_at IS NOT NULL AND (
    -- All songs created by some artists
    ((@by_singer::BOOLEAN OR @with_artist::BOOLEAN) AND singer_fk = ANY(@singers_ids::UUID[])) OR
    -- All songs created by or featured some artists
    (@with_artist::BOOLEAN AND @singers_ids::UUID[] <@ (
        SELECT ARRAY_AGG(feats.artist_fk)
        FROM feats
        WHERE song_fk = songs.song_id
        GROUP BY song_fk
    )) OR
    -- All songs matching name
    (@by_name::BOOLEAN AND name ILIKE CONCAT('%', @match_name::TEXT, '%')) OR
    -- Songs by IDs
    (@by_ids::BOOLEAN AND song_id = ANY(@ids::UUID[]))
    )
GROUP BY songs.song_id
ORDER BY songs.uploaded_at DESC, name
LIMIT @limitv
OFFSET @offsetv;

-- name: MySongs :many
SELECT
    sqlc.embed(songs),
    ARRAY_AGG(feats.artist_fk ORDER BY feats.order_num)::UUID[] AS artists_ids
FROM songs
LEFT JOIN feats ON feats.song_fk = songs.song_id
WHERE singer_fk = @singer_id::UUID AND (NOT @by_ids::BOOLEAN OR song_id = ANY(@ids::UUID[]))
GROUP BY songs.song_id
ORDER BY songs.uploaded_at DESC, name
LIMIT @limitv
OFFSET @offsetv;

-- name: MySong :one
SELECT sqlc.embed(songs)
FROM songs
WHERE singer_fk = @singer_id::UUID AND song_id = @song_id::UUID;

-- name: CountMySongs :one
SELECT COUNT(*)::INT
FROM songs
WHERE singer_fk = @singer_id::UUID;

-- name: CountSongsWithArtistsIds :one
WITH cte AS (
    SELECT COUNT(*)
    FROM songs
    WHERE singer_fk = ANY(@singers_ids::UUID[])
)
SELECT COUNT(*)::INT + (SELECT * FROM cte)
FROM feats
WHERE artist_fk = ANY(@singers_ids::UUID[]);

-- name: CountSongsMatchName :one
SELECT COUNT(*)::INT
FROM songs
WHERE name ILIKE CONCAT('%', @match_name::TEXT, '%');