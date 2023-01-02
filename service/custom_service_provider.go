package service

import (
	"context"
	"fmt"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/internal/logger"
	"github.com/jmatsu/splitter/internal/net"
	"github.com/rs/zerolog"
	"strings"
)

var customServiceLogger zerolog.Logger

func init() {
	customServiceLogger = logger.Logger.With().Str("service", "custom").Logger()
}

func NewCustomServiceProvider(ctx context.Context, definition *config.CustomServiceDefinition, conf *config.CustomServiceConfig) *CustomServiceProvider {
	scheme, t, _ := strings.Cut(definition.Endpoint, "://")
	hostname, path, _ := strings.Cut(t, "/")

	return &CustomServiceProvider{
		CustomServiceConfig:     *conf,
		CustomServiceDefinition: *definition,
		ctx:                     ctx,
		client:                  net.NewHttpClient(fmt.Sprintf("%s://%s", scheme, hostname)),
		path:                    path,
	}
}

type CustomServiceProvider struct {
	config.CustomServiceConfig
	config.CustomServiceDefinition
	ctx    context.Context
	client *net.HttpClient
	path   string
}
