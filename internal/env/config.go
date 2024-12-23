package env

import "net/url"

type Config struct {
	DatabaseURL            *url.URL `env:"DATABASE_URL"`
	AssetsRootURL          *url.URL `env:"ASSETS_ROOT_URL"`
	AssetsUseManifest      bool     `env:"ASSETS_USE_MANIFEST"`
	AssetsManifestLocation string   `env:"ASSETS_MANIFEST_LOCATION"`
}
