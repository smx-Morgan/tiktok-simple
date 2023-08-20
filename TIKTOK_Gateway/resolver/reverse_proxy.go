package resolver

import (
	"TIKTOK_Gateway/configs"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/hertz-contrib/reverseproxy"
)

func CreateProxy(myConfig *configs.Config) map[string]app.HandlerFunc {
	ret := make(map[string]app.HandlerFunc)

	// 创建服务发现cli
	cli := CreateDiscoveryClient(myConfig)

	for _, route := range myConfig.Routes {
		if _, exist := ret[route.ServiceName]; !exist {
			proxy, _ := reverseproxy.NewSingleHostReverseProxy(route.ServiceName)
			proxy.SetClient(cli)
			proxy.SetDirector(func(req *protocol.Request) {
				req.SetRequestURI(string(reverseproxy.JoinURLPath(req, proxy.Target)))
				req.Header.SetHostBytes(req.URI().Host())
				req.Options().Apply([]config.RequestOption{config.WithSD(true)})
			})
			ret[route.ServiceName] = proxy.ServeHTTP
		}
	}

	return ret
}
