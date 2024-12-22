package covers

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/google/uuid"
	"strings"
)

func getObjectID(userId, playlistId uuid.UUID, fileExt string) string {
	objectIDHash := sha1.Sum([]byte(userId.String() + "\u0002" + playlistId.String())) //nolint:gosec
	return hex.EncodeToString(objectIDHash[:]) + fileExt
}

func (s *ServiceCovers) CoverURL(playlistID, rawCoverId string) string {
	tmpl := strings.Split(s.coverURLTmpl, "{playlist_id}")

	return tmpl[0] + playlistID + tmpl[1] + rawCoverId
}
