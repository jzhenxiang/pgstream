package router

import "errors"

// Validate checks that the CORSConfig is well-formed.
// It returns an error if AllowedOrigins contains blank entries.
func (c *CORSConfig) Validate() error {
	if c == nil {
		return nil
	}
	for _, o := range c.AllowedOrigins {
		if o == "" {
			return errors.New("cors: allowed origin must not be blank")
		}
	}
	for _, m := range c.AllowedMethods {
		if m == "" {
			return errors.New("cors: allowed method must not be blank")
		}
	}
	for _, h := range c.AllowedHeaders {
		if h == "" {
			return errors.New("cors: allowed header must not be blank")
		}
	}
	if c.MaxAge < 0 {
		return errors.New("cors: max age must not be negative")
	}
	return nil
}
