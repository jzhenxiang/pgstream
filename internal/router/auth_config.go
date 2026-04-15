package router

import "errors"

// AuthConfig holds configuration for bearer token authentication middleware.
type AuthConfig struct {
	// Tokens is the set of valid bearer tokens.
	Tokens []string
}

// Validate returns an error if the configuration is invalid.
func (c *AuthConfig) Validate() error {
	if c == nil {
		return errors.New("auth config must not be nil")
	}
	for _, t := range c.Tokens {
		if t != "" {
			return nil
		}
	}
	if len(c.Tokens) > 0 {
		return errors.New("auth config must contain at least one non-blank token")
	}
	return errors.New("auth config must contain at least one token")
}

// validTokens returns a set of non-blank tokens for O(1) lookup.
func (c *AuthConfig) validTokens() map[string]struct{} {
	out := make(map[string]struct{}, len(c.Tokens))
	for _, t := range c.Tokens {
		if t != "" {
			out[t] = struct{}{}
		}
	}
	return out
}
