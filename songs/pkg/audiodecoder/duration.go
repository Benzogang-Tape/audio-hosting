package audiodecoder

import (
	"context"
	"errors"
	"io"
	"time"

	"dev.gaijin.team/go/golib/e"
	"github.com/tcolgate/mp3"
)

type Decoder struct{}

func (Decoder) GetMp3Duration(ctx context.Context, r io.Reader) (time.Duration, error) {
	var (
		frame   mp3.Frame
		skipped int
		dur     time.Duration
		d       = mp3.NewDecoder(r)
	)

	for {
		select {
		case <-ctx.Done():
			return 0, ctx.Err() //nolint:wrapcheck

		default:
		}

		err := d.Decode(&frame, &skipped)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return 0, e.NewFrom("decoding mp3", err)
		}

		dur += frame.Duration()
	}

	return dur, nil
}
