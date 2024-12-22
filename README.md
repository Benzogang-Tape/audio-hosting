# Audio Hosting Service

## Service startup

To deploy the Audio Hosting service, follow these steps:

- `git clone https://github.com/Benzogang-Tape/audio-hosting.git`

## configs

In configs directory you can find configuration files for the services.
Fill the files based on the examples.

To fill configs with default values run `task default-config` or
`bash -c 'for i in $( ls ./configs/example.* ); do cp $i ./configs/${i#*example.}; done'`
or just remove `example.` prefix on each file in configs directory.

## first launch
Run `docker compose -f docker/prod/docker-compose.yml up minio -d`

Before the first launch of the services you need to browse http://localhost:9001 login with creds from ./configs/minio.env.
Click there on access keys and generate new pair of access and secret keys.
After that you should change `connections.minio.accessKey` and `connections.minio.secretKey`
keys in ./configs/service.yaml and ./configs/songs.yaml.

## docker

- to build and run service in docker run `task full-launch-build` or
`docker compose -f docker/prod/docker-compose.yml up --build -d`

- to run service without building in docker run `task full-launch` or
`docker compose -f docker/prod/docker-compose.yml up -d`


# Common parts of Audio Hosting service

# JWT access tokens Library

This library provides functionality for creating, signing, and verifying JWT (JSON Web Token) access tokens for authentication within an audio hosting service. It uses the Ed25519 signature algorithm for security.

## Usage

### 1. Key Generation

Before using the library, you must generate a pair of Ed25519 public and private keys. These keys are crucial for signing and verifying tokens. You can generate keys using `go run cmd/keygen/main.go`

## 2. Signing (in users service)

A Signer is responsible for creating and signing JWTs. You initialize it with your private key (in base64 encoded string format) and a token expiration duration.

```go
import (
    "time"
    "github.com/google/uuid"
    "github.com/Benzogang-Tape/audio-hosting/auth"
)

privateKey := os.Getenv("JWT_PRIVATE_KEY") // Getting base64 encoded private key
tokenDuration := time.Minute * 15 // Example: 15-minute token lifespan

signer, err := auth.NewSigner(privateKey, tokenDuration)
if err != nil {}

signedToken, err := signer.Sign(auth.Token{Subject: uuid.New(), IsArtist: true})
```

### 3. Verifying (in other services)

```go
import (
	"context"
	"github.com/Benzogang-Tape/audio-hosting/auth"
)

publicKey := os.Getenv("JWT_PUBLIC_KEY") // Getting base64 encoded public key

verifier, err := auth.NewVerifier(publicKey)
if err != nil {}

bearer := "jwt.token.here" // Getting token from request headers or metadata (Authorization: Bearer <token>)
token, err := verifier.Verify(bearer)
if err != nil {}
```
