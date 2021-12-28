package ycache

import (
	"7days/ycache/consistenthash"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_ycache/"
	defaultReplicas = 50
)

// HTTPPool implements PeerPicker for a pool of HTTP peers.
type HTTPPool struct {
	// this peer's base URL, e.g. "https://example.com:9999"
	self        string
	basePath    string
	mu          sync.Mutex
	peers       *consistenthash.Map
	httpGetters map[string]*httpGetter
}

// NewHttpPool initializes an HTTP pool of peers.
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// log info service name
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Service %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP handle all http requests
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}

	// /<basePath>/<groupname>/<key> required
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(r.Context(), key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}

// Set 更新HTTP池节点列表
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

// PickPeer 根据key选择一个节点
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer: %s", peer)
		return p.httpGetters[peer], true
	}

	return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil)

type httpGetter struct {
	baseURL string
}

func (h *httpGetter) Get(ctx context.Context, group string, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)

	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", resp.Status)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, nil
}

// do what?
var _ PeerGetter = (*httpGetter)(nil)
