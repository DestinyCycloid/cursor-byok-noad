//go:build noads

package app

import (
	"context"

	bridge "cursor/internal/bridge"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type adController struct{}

func registerAdEvents()                                               {}
func newAdController(*bridge.ProxyService, string) *adController      { return &adController{} }
func (c *adController) Services() []application.Service               { return nil }
func (c *adController) SetApp(*application.App)                       {}
func (c *adController) Stop()                                         {}
func (c *adController) RefreshAssetBaseURL(*bridge.ProxyService) bool { return false }
func (c *adController) RefreshRuntime()                               {}
func (c *adController) Refresh(context.Context)                       {}
func (c *adController) RefreshAsync()                                 {}
func (c *adController) Start(context.Context)                         {}
