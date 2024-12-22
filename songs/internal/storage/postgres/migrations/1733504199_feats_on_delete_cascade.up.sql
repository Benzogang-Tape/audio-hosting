ALTER TABLE feats DROP CONSTRAINT feats_song_fk_fkey;
ALTER TABLE feats
ADD CONSTRAINT feats_song_fk_fkey
FOREIGN KEY (song_fk)
REFERENCES songs(song_id)
ON DELETE CASCADE;