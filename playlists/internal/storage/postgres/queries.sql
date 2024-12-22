-- name: SavePlaylist :exec
INSERT INTO playlists (
    id,
    title,
    author_id,
    track_ids,
    cover_url,
    is_album,
    created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: DeletePlaylists :exec
DELETE FROM playlists WHERE id = ANY(@ids::UUID[]) and author_id = @user_id::UUID;

-- name: Playlist :one
SELECT
    sqlc.embed(playlists)
FROM playlists
WHERE id = $1;

-- name: UserPlaylists :many
SELECT
    id,
    title,
    author_id,
    cover_url,
    created_at,
    updated_at,
    released_at,
    is_album,
    is_public
FROM playlists
WHERE playlists.author_id = @user_id::UUID
UNION
SELECT
    id,
    title,
    author_id,
    cover_url,
    created_at,
    updated_at,
    released_at,
    is_album,
    is_public
FROM playlists
LEFT JOIN liked_playlists ON playlists.id = liked_playlists.liked_playlist
WHERE liked_playlists.user_id = @user_id::UUID AND (playlists.is_public)
ORDER BY created_at DESC;

-- name: MyCollection :many
SELECT
    track_id,
    liked_at
FROM liked_tracks
WHERE user_id = @user_id::UUID
ORDER BY liked_at DESC;

-- name: LikePlaylist :one
WITH inserted_row AS (
    INSERT INTO liked_playlists (liked_playlist, user_id)
    SELECT id, @user_id::UUID
    FROM playlists
    WHERE id = @playlist_id::UUID AND (is_public OR author_id = @user_id::UUID)
    RETURNING liked_playlist)
SELECT liked_playlist FROM inserted_row;

-- name: DislikePlaylist :exec
DELETE FROM liked_playlists WHERE liked_playlist = @playlist_id::UUID AND user_id = @user_id::UUID;

-- name: UpdatePlaylist :one
UPDATE playlists SET
                 title = $2,
                 cover_url = $3,
                 track_ids = $4,
                 is_album = $5,
                 is_public = $6,
                 released_at = $7,
                 updated_at = $8
WHERE id = $1 AND author_id = @user_id::UUID
RETURNING *;

-- name: PatchPlaylist :one
UPDATE playlists SET
                 title = COALESCE(sqlc.narg('title'), title),
                 cover_url = COALESCE(sqlc.narg('cover_url'), cover_url),
                 track_ids = COALESCE(sqlc.narg('track_ids'), track_ids),
                 released_at = COALESCE(sqlc.narg('released_at'), released_at),
                 is_album = COALESCE(sqlc.narg('is_album'), is_album),
                 is_public = COALESCE(sqlc.narg('is_public'), is_public),
                 updated_at = @updated_at::TIMESTAMPTZ
WHERE playlists.id = @id::UUID AND author_id = @user_id::UUID
RETURNING *;

-- name: CopyPlaylist :one
WITH row_for_copy AS (
    SELECT *
    FROM playlists
    WHERE playlists.id = @playlist_id::UUID AND
        playlists.is_public AND
        playlists.is_album = FALSE AND
        playlists.author_id != @user_id
)
INSERT INTO playlists (
    id,
    title,
    author_id,
    track_ids,
    cover_url,
    is_album,
    is_public,
    created_at
) SELECT
    @new_playlist_id::UUID as new_playlis_id,
    title,
    @user_id::UUID,
    track_ids,
    cover_url,
    FALSE,
    FALSE,
    NOW()
FROM row_for_copy
RETURNING id;

-- name: LikeTrack :one
WITH inserted_row AS (
    INSERT INTO liked_tracks (track_id, user_id)
    VALUES (@track_id::UUID, @user_id::UUID) ON CONFLICT (track_id, user_id) DO NOTHING
    RETURNING track_id)
SELECT track_id FROM inserted_row;

-- name: DislikeTrack :exec
DELETE FROM liked_tracks WHERE track_id = @track_id::UUID AND user_id = @user_id::UUID;

-- name: PublicPlaylists :many
SELECT
    sqlc.embed(playlists)
FROM playlists
WHERE
  -- Only public playlists!
    is_public AND (
    -- All albums created by some artists
    ((@by_artist_id::BOOLEAN) AND author_id = @artist_id::UUID AND is_album) OR
        -- All playlists matching name
    (@by_title::BOOLEAN AND title ILIKE CONCAT('%', @match_name::TEXT, '%')) OR
        -- Playlists by IDs
    (@by_ids::BOOLEAN AND id = ANY(@ids::UUID[]))
    )
ORDER BY created_at DESC, title
LIMIT @limitv
OFFSET @offsetv;