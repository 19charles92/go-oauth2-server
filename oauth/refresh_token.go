package oauth

import (
	"errors"
	"time"

	"github.com/RichardKnop/go-oauth2-server/util"
)

// GetOrCreateRefreshToken retrieves an existing token or creates a new one
// If the retrieved token is expired, it is deleted and a new one created
func (s *Service) GetOrCreateRefreshToken(client *Client, user *User, scope string) (*RefreshToken, error) {
	// Try to fetch an existing refresh token first
	refreshToken := new(RefreshToken)
	found := !s.db.Where(RefreshToken{
		ClientID: util.IntOrNull(client.ID),
		UserID:   util.IntOrNull(user.ID),
	}).Preload("Client").Preload("User").First(refreshToken).RecordNotFound()

	// Check if the token is expired, if found
	var expired bool
	if found {
		expired = time.Now().After(refreshToken.ExpiresAt)
	}

	// If the refresh token has expired, delete it
	if expired {
		s.db.Delete(refreshToken)
	}

	// Create a new refresh token if it expired or was not found
	if expired || !found {
		refreshToken = newRefreshToken(
			s.cnf.Oauth.RefreshTokenLifetime, // expires in
			client, // client
			user,   // user
			scope,  // scope
		)
		if err := s.db.Create(refreshToken).Error; err != nil {
			return nil, errors.New("Error saving refresh token")
		}
	}

	return refreshToken, nil
}

// GetValidRefreshToken return a valid non expired refresh token
func (s *Service) GetValidRefreshToken(token string, client *Client) (*RefreshToken, error) {
	// Fetch the refresh token from the database
	refreshToken := new(RefreshToken)
	if s.db.Where(RefreshToken{
		Token:    token,
		ClientID: util.IntOrNull(client.ID),
	}).Preload("Client").Preload("User").First(refreshToken).RecordNotFound() {
		return nil, errors.New("Refresh token not found")
	}

	// Check the refresh token hasn't expired
	if time.Now().After(refreshToken.ExpiresAt) {
		return nil, errors.New("Refresh token expired")
	}

	return refreshToken, nil
}
