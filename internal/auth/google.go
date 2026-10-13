package auth

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"

	"google.golang.org/api/idtoken"
)

type GoogleAuthenticator struct {
	clientID     string
	allowedUsers map[string]struct{}
}

func NewGoogleAuthenticator() (*GoogleAuthenticator, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")

	if clientID == "" {
		return nil, errors.New("GOOGLE CLIENT_ID is missing")
	}

	allowedUsers := make(map[string]struct{})

	for _, sub := range strings.Split(os.Getenv("GOOGLE_ALLOWED_SUBS"), ",") {
		sub = strings.TrimSpace(sub)

		if sub != "" {
			allowedUsers[sub] = struct{}{}
		}
	}

	return &GoogleAuthenticator{
		clientID:     clientID,
		allowedUsers: allowedUsers,
	}, nil
}

func (g *GoogleAuthenticator) Verify(
	ctx context.Context,
	credential string,
) error {
	payload, err := idtoken.Validate(
		ctx,
		credential,
		g.clientID,
	)

	if err != nil {
		return err
	}
	log.Printf("Google account subject: %s", payload.Subject)

	if payload.Issuer != "accounts.google.com" &&
		payload.Issuer != "https://accounts.google.com" {
		return errors.New("Invalid Google issuer")
	}

	if _, ok := g.allowedUsers[payload.Subject]; !ok {
		return errors.New("Google account not authorised")
	}

	return nil
}
