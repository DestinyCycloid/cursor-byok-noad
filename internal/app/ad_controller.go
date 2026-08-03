//go:build !noads

package app

import (
	"context"
	"strings"
	"time"

	"cursor/internal/ads"
	"cursor/internal/appdata"
	serverconfig "cursor/internal/backend/server/config"
	bridge "cursor/internal/bridge"
	"cursor/internal/buildinfo"
	"cursor/internal/cursor"
	"cursor/internal/historymetrics"
	"cursor/internal/netproxy"

	"github.com/wailsapp/wails/v3/pkg/application"
)

const adRefreshInterval = 3 * time.Minute

type adController struct {
	core   *ads.Service
	app    *application.App
	cancel context.CancelFunc
}

func registerAdEvents() {
	application.RegisterEvent[bridge.AdRuntime](ads.EventUpdated)
}

func newAdController(proxyService *bridge.ProxyService, baseURL string) *adController {
	core := ads.NewService(ads.Options{
		StoreRoot:    appdata.AdsRootPath(),
		HTTPClient:   netproxy.NewHTTPClient(30 * time.Second),
		AppVersion:   buildinfo.CurrentVersion(),
		AssetBaseURL: baseURL + ads.RoutePrefix,
		DeviceID:     cursor.GetDeviceID,
		Metrics: func(context.Context) (ads.MetricsSnapshot, error) {
			if err := appdata.EnsureAssistantHome(); err != nil {
				return ads.MetricsSnapshot{}, err
			}
			summary, err := historymetrics.LoadUsageSummary(appdata.UsageFilePath())
			if err != nil {
				return ads.MetricsSnapshot{}, err
			}
			return ads.MetricsSnapshot{TurnsTotal: summary.TurnsTotal, RequestTokensTotal: summary.RequestTokensTotal, PromptTokensTotal: summary.PromptTokensTotal, CacheReadTokens: summary.CacheReadTokens, CacheWriteTokens: summary.CacheWriteTokens}, nil
		},
		ProviderCount: func(context.Context) (int, error) {
			cfg, err := proxyService.LoadUserConfig()
			if err != nil {
				return 0, err
			}
			return len(cfg.ModelAdapters), nil
		},
	})
	return &adController{core: core}
}

func (c *adController) Services() []application.Service {
	return []application.Service{application.NewService(bridge.NewAdService(c.core))}
}
func (c *adController) SetApp(app *application.App) { c.app = app }
func (c *adController) Stop() {
	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
	}
}
func (c *adController) RefreshAssetBaseURL(proxyService *bridge.ProxyService) bool {
	state := proxyService.GetState()
	addr := strings.TrimSpace(state.BackendListenAddr)
	if addr == "" {
		addr = serverconfig.DefaultBackendListenAddr
	}
	return c.core.SetAssetBaseURL(browserReachableLoopbackBaseURL(addr) + ads.RoutePrefix)
}
func (c *adController) RefreshRuntime() {
	if c.app == nil {
		return
	}
	state, err := c.core.GetRuntime(context.Background())
	if err == nil {
		c.app.Event.Emit(ads.EventUpdated, state)
	}
}
func (c *adController) Refresh(ctx context.Context) {
	if c.app == nil {
		return
	}
	state, changed, err := c.core.Refresh(ctx)
	if err == nil && changed {
		c.app.Event.Emit(ads.EventUpdated, state)
	}
}
func (c *adController) RefreshAsync() { go c.Refresh(context.Background()) }
func (c *adController) Start(ctx context.Context) {
	c.Stop()
	refreshCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	go func() {
		c.Refresh(refreshCtx)
		ticker := time.NewTicker(adRefreshInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				c.Refresh(ctx)
			}
		}
	}()
}
