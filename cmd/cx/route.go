package main

import (
	"github.com/tommac/contextless/internal/config"
	"github.com/tommac/contextless/internal/route"
	"github.com/tommac/contextless/internal/store"
)

func routeNew(cfg config.Config, h *store.Store) (*route.Router, error) { return route.New(cfg, h) }
