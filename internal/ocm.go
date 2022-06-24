package internal

import (
	"fmt"

	"github.com/golang-jwt/jwt/v4"
	"github.com/openshift-online/ocm-cli/pkg/config"
	sdk "github.com/openshift-online/ocm-sdk-go"
)

func OcmLogin(ocmToken, gatewayURL string) (*config.Config, error) {
	parser := new(jwt.Parser)
	token, _, err := parser.ParseUnverified(ocmToken, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse the provided token: %w", err)
	}

	tokenURL := sdk.DefaultTokenURL
	clientID := sdk.DefaultClientID

	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load the config file: %w", err)
	}
	if cfg == nil {
		cfg = new(config.Config)
	}

	// Update the configuration with appropriate values
	cfg.TokenURL = tokenURL
	cfg.ClientID = clientID
	cfg.ClientSecret = ""
	cfg.Scopes = sdk.DefaultScopes
	cfg.URL = gatewayURL
	cfg.User = ""
	cfg.Password = ""
	cfg.Insecure = false
	cfg.AccessToken = ""
	cfg.RefreshToken = ""

	typ, err := tokenType(token)
	if err != nil {
		return nil, fmt.Errorf("failed to extract type from 'typ' claim of token '%s': %v", ocmToken, err)
	}

	switch typ {
	case "Bearer":
		cfg.AccessToken = ocmToken
	case "Refresh", "Offline":
		cfg.RefreshToken = ocmToken
	case "":
		return nil, fmt.Errorf("no ocm token found to be provided")
	default:
		return nil, fmt.Errorf("unknown data type of the ocm token found '%s'", typ)
	}

	// Create a connection and get the token to verify that the crendentials are correct
	connection, err := cfg.Connection()
	if err != nil {
		return nil, fmt.Errorf("failed to create connection with the processed config: %w", err)
	}
	accessToken, refreshToken, err := connection.Tokens()
	if err != nil {
		return nil, fmt.Errorf("failed to get access and refresh tokens from the established OCM connection: %w", err)
	}

	cfg.AccessToken = accessToken
	cfg.RefreshToken = refreshToken

	err = config.Save(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to save the config file: %w", err)
	}

	return cfg, nil
}

// tokenType extracts the value of the `typ` claim. It returns the value as a string, or the empty
// string if there is no such claim.
func tokenType(token *jwt.Token) (typ string, err error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		err = fmt.Errorf("expected map claims but got %T", claims)
		return
	}
	claim, ok := claims["typ"]
	if !ok {
		return
	}
	value, ok := claim.(string)
	if !ok {
		err = fmt.Errorf("expected string 'typ' but got %T", claim)
		return
	}
	typ = value
	return
}
