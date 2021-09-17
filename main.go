package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"

	gologging "github.com/devopsfaith/krakend-gologging"
	viper "github.com/devopsfaith/krakend-viper"
	"github.com/gin-gonic/gin"
	_ "github.com/kpacha/martian-components/body/querystring2body"
	krakendtwirp "github.com/kyawmyintthein/api-gateway-poc/krakend-twirp"
	_ "github.com/kyawmyintthein/api-gateway-poc/requestbodytransformer"
	svcc "github.com/kyawmyintthein/api-gateway-poc/rpc/svc_c"
	"github.com/luraproject/lura/proxy"
	krakendgin "github.com/luraproject/lura/router/gin"
	"github.com/luraproject/lura/transport/http/client"
	"github.com/luraproject/lura/transport/http/server"
	"github.com/twitchtv/twirp"
)

func main() {
	port := flag.Int("p", 0, "Port of the service")
	debug := flag.Bool("d", false, "Enable the debug")
	configFile := flag.String("c", "./configuration.json", "Path to the configuration filename")
	flag.Parse()

	parser := viper.New()
	serviceConfig, err := parser.Parse(*configFile)
	if err != nil {
		log.Fatal("ERROR:", err.Error())
	}
	serviceConfig.Debug = serviceConfig.Debug || *debug
	if *port != 0 {
		serviceConfig.Port = *port
	}

	logger, err := gologging.NewLogger(serviceConfig.ExtraConfig, os.Stdout)
	if err != nil {
		log.Fatal("ERROR:", err.Error())
	}

	logger.Debug("config:", serviceConfig)

	ctx, cancel := context.WithCancel(context.Background())

	//backendFactory := martian.NewBackendFactory(logger, client.DefaultHTTPRequestExecutor(client.NewHTTPClient))

	svccLuraClient, err := svcc.NewCServiceLuraClient(&serviceConfig, "rpc.svcc.CService", http.DefaultClient, twirp.WithClientPathPrefix("rz"))
	if err != nil {
		panic(err)
	}
	krakendtwirp.RegisterClients(svccLuraClient)

	bf := krakendtwirp.NewTwirpProxy(logger, proxy.CustomHTTPProxyFactory(client.NewHTTPClient))
	routerFactory := krakendgin.NewFactory(krakendgin.Config{
		Engine:         gin.Default(),
		Logger:         logger,
		Middlewares:    []gin.HandlerFunc{},
		HandlerFactory: krakendgin.EndpointHandler,
		ProxyFactory:   proxy.NewDefaultFactory(bf, logger),
		RunServer:      server.RunServer,
	})

	routerFactory.NewWithContext(ctx).Run(serviceConfig)

	cancel()
}
