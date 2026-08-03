//go:build !noads

package backend

import (
	"cursor/internal/ads"
	"cursor/internal/appdata"
	"cursor/internal/backend/server"
)

func adRoutes() server.Option {
	return server.Mount(ads.RoutePrefix, ads.NewHTTPHandler(appdata.AdsRootPath()))
}
