package svcc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	krakendtwirp "github.com/kyawmyintthein/api-gateway-poc/krakend-twirp"
	"github.com/luraproject/lura/config"
	twirp "github.com/twitchtv/twirp"
	"google.golang.org/protobuf/proto"
)

type luraClient struct {
	name    string
	service CService
}

func NewCServiceLuraClient(config *config.ServiceConfig, name string, client HTTPClient, opts ...twirp.ClientOption) (krakendtwirp.TwirpLuraClient, error) {
	baseURL, err := getBaseURLByClientID(config)
	if err != nil {
		return nil, err
	}
	protobufClient := NewCServiceProtobufClient(baseURL, client, opts...)
	return &luraClient{
		name:    name,
		service: protobufClient,
	}, nil
}

func getBaseURLByClientID(config *config.ServiceConfig) (string, error) {
	for _, endpoint := range config.Endpoints {
		_, ok := endpoint.ExtraConfig["twirp_service_name"].(string)
		if ok {
			for _, backend := range endpoint.Backend {
				_, ok := backend.ExtraConfig["twirp_service_name"].(string)
				if ok {
					return backend.Host[0], nil
				}
			}
		}
	}
	return "", fmt.Errorf("invalid service")
}

func (luraClient *luraClient) Call(ctx context.Context, service string, method string, in proto.Message) (proto.Message, error) {
	switch method {
	case "CallServiceC":
		req, ok := in.(*GetServiceCRequest)
		if !ok {
			return nil, twirp.InternalError("invalid payload")
		}
		log.Println("Req", req, in)
		resp, err := luraClient.service.CallServiceC(ctx, req)
		if err != nil {
			return resp, err
		}
		log.Println("Resp", resp)
		return resp, nil
	}
	return nil, twirp.InternalError("Invalid service method")
}

func (luraClient *luraClient) Name() string {
	return luraClient.name
}

func (luraClient *luraClient) Unmarshal(ctx context.Context, method string, data []byte) (proto.Message, error) {
	switch method {
	case "CallServiceC":
		out := new(GetServiceCRequest)
		err := json.Unmarshal(data, out)
		if err != nil {
			return out, err
		}
		return out, nil
	}
	return nil, fmt.Errorf("invalid method")
}
