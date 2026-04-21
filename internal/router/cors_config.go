package router

import "errors"

// CORSConfig holds configuration for the CORS middleware.
type CORSConfig struct {
	// AllowedOrigins is the list of origins that are allowed.
	// Use "*" to allow all origins.
	AllowedOrigins []string

	// AllowedMethods is the list of HTTP methods that are allowed.
	// Defaults to GET, POST, PUT, DELETE, OPTIONS if empty.
	AllowedMethods []string

	// AllowedHeaders is the list of HTTP headers that are allowed.
	AllowedHeaders []string

	// AllowCredentials indicates whether the request can include
	// user credentials such as cookies or HTTP authentication.
	AllowCredentials bool

	// MaxAge indicates how long (in seconds) the results of a
	// preflight request can be cached.
	MaxAge int
}

// Validate returns an error if the CORSConfig is invalid.
func (c *CORSConfig) Validate() error {
	if c == nil {
		return errors.New("cors: config must not be nil")
	}
	for _, o := range c.AllowedOrigins {
		if o == "" {
			return errors.New("cors: origin must not be blank")
		}
	}
	for _, m := range c.AllowedMethods {
		if m == "" {
			return errors.New("cors: method must not be blank")
		}
	}
	for _, h := range c.AllowedHeaders {
		if h == "" {
			return errors.New("cors: header must not be blank")
		}
	}
	return nil
}
