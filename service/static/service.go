package static

import (
	rrttp "github.com/spiral/roadrunner/service/http"
	"net/http"
	"path"
	"strings"
)

// ID contains default service name.
const ID = "static"

// Service serves static files. Potentially convert into middleware?
type Service struct {
	// server configuration (location, forbidden files and etc)
	cfg *Config

	// root is initiated http directory
	root http.Dir
}

// Init must return configure service and return true if service hasStatus enabled. Must return error in case of
// misconfiguration. Services must not be used without proper configuration pushed first.
func (s *Service) Init(cfg *Config, r *rrttp.Service) (bool, error) {
	if r == nil {
		return false, nil
	}

	s.cfg = cfg
	s.root = http.Dir(s.cfg.Dir)
	r.AddMiddleware(s.middleware)

	return true, nil
}

// middleware must return true if request/response pair is handled within the middleware.
func (s *Service) middleware(f http.HandlerFunc) http.HandlerFunc {
	// Define the http.HandlerFunc
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.handleStatic(w, r) {
			f(w, r)
		}
	}
}

func (s *Service) handleStatic(w http.ResponseWriter, r *http.Request) bool {
	fPath := r.URL.Path

	if !strings.HasPrefix(fPath, "/") {
		fPath = "/" + fPath
	}
	fPath = path.Clean(fPath)

	if s.cfg.Forbids(fPath) {
		return false
	}

	f, err := s.root.Open(fPath)
	if err != nil {
		return false
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		return false
	}

	// do not serve directories
	if d.IsDir() {
		return false
	}

	http.ServeContent(w, r, d.Name(), d.ModTime(), f)
	return true
}
