//go:build noads

package backend

import "cursor/internal/backend/server"

func adRoutes() server.Option { return nil }
