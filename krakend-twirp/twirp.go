package krakendtwirp

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/luraproject/lura/config"
	"github.com/luraproject/lura/logging"
	"github.com/luraproject/lura/proxy"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type TwirpLuraClient interface {
	Call(context.Context, string, string, proto.Message) (proto.Message, error)
	Unmarshal(context.Context, string, []byte) (proto.Message, error)
	Name() string
}

type registry struct {
	r  map[string]TwirpLuraClient
	mu *sync.RWMutex
}

var clientRegistry = registry{
	r:  map[string]TwirpLuraClient{},
	mu: new(sync.RWMutex),
}

func RegisterClients(
	clients ...TwirpLuraClient,
) {
	for _, client := range clients {
		clientRegistry.Set(client.Name(), client)
	}
}

func (r *registry) Set(
	name string,
	client TwirpLuraClient,
) {
	r.mu.Lock()
	r.r[name] = client
	r.mu.Unlock()
}

func (r *registry) Get(name string) (TwirpLuraClient, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	v, ok := r.r[name]
	return v, ok
}

func NewTwirpProxy(l logging.Logger, f proxy.BackendFactory) proxy.BackendFactory {
	return func(remote *config.Backend) proxy.Proxy {
		bo := getOptions(remote)
		if bo == nil {
			log.Println("twirp: client factory is not used for", remote)
			return f(remote)
		}

		return func(ctx context.Context, request *proxy.Request) (*proxy.Response, error) {
			resp, err := callService(ctx, request, bo)
			request.Body.Close()
			if err != nil {
				l.Warning("gRPC calling the next mw:", err.Error())
				return nil, err
			}
			return resp, err
		}
	}
}

func callService(ctx context.Context, request *proxy.Request, opts *twirpBackendOptions) (*proxy.Response, error) {
	caller := func(ctx context.Context, req *proxy.Request) (*proxy.Response, error) {
		luraTwirpClient, ok := clientRegistry.Get(opts.protoDef)
		if !ok {
			return nil, fmt.Errorf("Client not found")
		}

		var in proto.Message
		if req.Body != nil {
			payload, err := ioutil.ReadAll(req.Body)
			if err != nil {
				log.Println("Request Payload Error", err)
				return nil, err
			}
			log.Println("Payload", string(payload))
			in, err = luraTwirpClient.Unmarshal(ctx, opts.method, payload)
			if err != nil {
				return nil, err
			}
		}
		log.Println("IN", in)

		resp, err := luraTwirpClient.Call(ctx, opts.serviceName, opts.method, in)
		if err != nil {
			return nil, err
		}

		str := protojson.Format(resp)
		var data map[string]interface{}
		err = json.Unmarshal([]byte(str), &data)
		if err != nil {
			return nil, err
		}
		return &proxy.Response{
			Data:       data,
			IsComplete: true,
			Metadata: proxy.Metadata{
				Headers: make(map[string][]string),
			},
		}, err
	}
	return caller(ctx, request)
}

type twirpBackendOptions struct {
	serviceName string
	method      string
	protoDef    string
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func getOptions(remote *config.Backend) *twirpBackendOptions {
	svc, _ := remote.ExtraConfig["twirp_service_name"].(string)
	return &twirpBackendOptions{
		method:      remote.Method,
		serviceName: strings.TrimPrefix(remote.URLPattern, "/"),
		protoDef:    svc,
	}
}
