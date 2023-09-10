package lru

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/xiaobinqt/cat/lru/consistenthash"
	"github.com/xiaobinqt/cat/lru/pb"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/-cat-cache/"
	defaultReplicas = 50
)

type HTTPPool struct {
	self string // self address,record host,port etc.
	// peers communicating address prefix,likes http://example.com:8088/-cat-cache
	// used to communicate peers, and other prefix use other service,likes /api prefix use
	// api service.
	basePath    string
	mu          sync.Mutex // guards peers and httpGetters
	peers       *consistenthash.Map
	httpGetters map[string]*httpGetter // key es: "http://0.0.0.2:8088"
}

type httpGetter struct {
	baseURL string
}

func (h *httpGetter) Get(in *pb.Request, out *pb.Response) error {
	u := fmt.Sprintf("%s%s/%s", h.baseURL, url.QueryEscape(in.Group), url.QueryEscape(in.Key))
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body err:%s", err.Error())
	}

	if err = proto.UnmarshalMerge(bytes, out); err != nil {
		return fmt.Errorf("decoding  response body err: %s ", err.Error())

	}

	return nil
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// Set updates the pool's list of peers
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{
			baseURL: peer + p.basePath,
		}
	}
}

func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 找另一个服务器，如果是自己的话就是 getLocally()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}

	return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil)

func (p *HTTPPool) Log(format string, v ...any) {
	log.Printf("[Server %s ] %s", p.self, fmt.Sprintf(format, v...))
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, p.basePath) == false {
		panic("HTTPPool serving unexcepted path: " + r.URL.Path)
	}

	p.Log("%s %s", r.Method, r.URL.Path)
	// /<basepath>/<groupname>/<key> required
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group:"+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body, err := proto.Marshal(&pb.Response{Value: view.b})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(body)
}
