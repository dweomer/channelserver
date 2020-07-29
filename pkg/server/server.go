package server

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rancher/channelserver/pkg/config"
	"github.com/rancher/channelserver/pkg/model"
	"github.com/rancher/channelserver/pkg/server/store"
	"github.com/rancher/channelserver/pkg/server/store/release"
	"github.com/rancher/steve/pkg/schemaserver/server"
	"github.com/rancher/steve/pkg/schemaserver/store/apiroot"
	"github.com/rancher/steve/pkg/schemaserver/types"
)

func ListenAndServe(ctx context.Context, address string, configs []*config.Config, pathPrefix []string) error {
	server := server.DefaultAPIServer()
	var err error
	for index, config := range configs {
		server.Schemas.MustImportAndCustomize(model.Channel{}, func(schema *types.APISchema) {
			schema.Store = store.New(config)
			schema.CollectionMethods = []string{http.MethodGet}
			schema.ResourceMethods = []string{http.MethodGet}
		})
		server.Schemas.MustImportAndCustomize(model.Release{}, func(schema *types.APISchema) {
			schema.Store = release.New(config)
			schema.CollectionMethods = []string{http.MethodGet}
		})

		pathPrefix[index] = strings.TrimPrefix(pathPrefix[index], "/")
		pathPrefix[index] = strings.TrimSuffix(pathPrefix[index], "/")

		router := mux.NewRouter()

		apiroot.Register(server.Schemas, []string{pathPrefix[index]}, nil)
		router.MatcherFunc(setType("apiRoot", pathPrefix[index])).Path("/").Handler(server)
		router.MatcherFunc(setType("apiRoot", pathPrefix[index])).Path("/{name}").Handler(server)
		router.Path("/{prefix:" + pathPrefix[index] + "}/{type}").Handler(server)
		router.Path("/{prefix:" + pathPrefix[index] + "}/{type}/{name}").Handler(server)

		fmt.Println(config, pathPrefix[index], config.ChannelsConfig())
		next := handlers.LoggingHandler(os.Stdout, router)
		handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			user := req.Header.Get("X-SUC-Cluster-ID")
			if user != "" && req.URL != nil {
				req.URL.User = url.User(user)
			}
			next.ServeHTTP(rw, req)
		})

		err = http.ListenAndServe(address, handler)

	}

	return err
}
func setType(t string, pathPrefix string) mux.MatcherFunc {
	return func(request *http.Request, match *mux.RouteMatch) bool {
		if match.Vars == nil {
			match.Vars = map[string]string{}
		}
		match.Vars["type"] = t
		match.Vars["prefix"] = pathPrefix
		return true
	}
}
