package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/audio/auth"

	"github.com/google/uuid"
)

var (
	pubStr  = "Gvbo6JyyS410wg87Gq0N9kphc67A5Lb1VS1wupzwYTU=" //nolint:unused
	privStr = "YerVSsnvOioynes/DkZ8cftHYvlSGimZVz/MRMi23l0a9ujonLJLjXTCDzsarQ32SmFzrsDktvVVLXC6nPBhNQ=="
)

func main() {
	userId := flag.String("uuid", uuid.New().String(), "user's uuid to generate token for")
	isArtist := flag.Bool("artist", true, "is user an artist")
	exp := flag.Duration("exp", time.Hour*24*365*100, "token expiration")

	flag.Parse()

	signer := must(auth.NewSigner(privStr, *exp))

	tkn := must(signer.Sign(auth.Token{ //nolint:exhaustruct
		Subject:  must(uuid.Parse(*userId)),
		IsArtist: *isArtist,
	}))

	_, _ = fmt.Println("user id:", *userId)     //nolint:forbidigo
	_, _ = fmt.Println("is artist:", *isArtist) //nolint:forbidigo
	_, _ = fmt.Println("exp:", *exp)            //nolint:forbidigo
	_, _ = fmt.Println(tkn)                     //nolint:forbidigo
}

func must[T any](v T, err error) T {
	if err != nil {
		_, _ = fmt.Println(err) //nolint:forbidigo
		os.Exit(1)              //nolint:revive,wsl
	}

	return v
}
