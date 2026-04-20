package router

import "errors"

// CORSConfig holds configuration for the CORS middleware.
type CORSConfig struct {
	// AllowedOrigins is the list of origins allowed to make cross-origin requests.
	// Use "*" to allow all origins.
	AllowedOrigins []string

	// AllowedMethods is the list of HTTP methods permitted in CORS requests.
	// Defaults to [GET, HEAD, POST] if empty.
	AllowedMethods []string

	// AllowedHeaders is the list of request headers permitted in CORS requests.
	AllowedHeaders []string

	// AllowCredentials indicates whether the request can include user credentials.
	AllowCredentials bool

	// MaxAge is the value for the Access-Control-Max-Age header in seconds.
	// Zero means the header is omitted.
	MaxAge int
}

// Validate returns an error if the CORSConfig is invalid.
func (c *CORSConfig) Validate() error {
	if c == nil {
		return errors.New("cors: config must not be nil")
	}
	for _, o := range c.AllowedOrigins {
		if o == "" {
			return errors.New("cors: blank origin in AllowedOrigins")
		}
	}
	for _, m := range c.AllowedMethods {
		if m == "" {
			return errors.New("cors: blank method in AllowedMethods")
		}
	}
	for _, h := range c.AllowedHeaders {
		if h == "" {
			return errors.New("cors: blank header in AllowedHeaders")
		}
	}
	return nil
}
