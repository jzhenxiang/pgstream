package router

import (
	"encoding/json"
	"net/http"
	"runtime"
)

// BuildInfo holds version and build metadata exposed via the version endpoint.
type BuildInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
}

// defaultBuildInfo is populated at link time via -ldflags.
var defaultBuildInfo = BuildInfo{
	Version:   "dev",
	Commit:    "unknown",
	BuildTime: "unknown",
	GoVersion: runtime.Version(),
}

// WithVersion returns an http.Handler that serves build info as JSON.
// The caller may override the default build info by passing a non-nil *BuildInfo.
func WithVersion(info *BuildInfo) http.Handler {
	if info == nil {
		info = &defaultBuildInfo
	}
	// Ensure GoVersion is always populated even when overriding.
	if info.GoVersion == "" {
		info.GoVersion = runtime.Version()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(info)
	})
}
