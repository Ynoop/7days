package ycache

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
)

const defaultBasePath = "/_ycache/"

// HttpPool implements PeerPicker for a pool of HTTP peers.
type HttpPool struct {
	self     string
	basePath string
}

// NewHttpPool initializes an HTTP pool of peers.
func NewHTTPPool(self string) *HttpPool {
	return &HttpPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// log info service name
func (p *HttpPool) Log(format string, v ...interface{}) {
	log.Printf("[Service %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP handle all http requests
func (p *HttpPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}

	p.Log("%s, %s", r.Method, r.URL.Path)

	// /<basePath>/<groupname>/<key> required
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]
	ctx := context.Background()

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(ctx, key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}
