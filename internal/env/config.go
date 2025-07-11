package env

import "net/url"

type Config struct {
	DatabaseURL            *url.URL `env:"DATABASE_URL,required"`
	Debug                  bool     `env:"DEBUG"                    envDefault:"false"`
	LogFormat              string   `env:"LOG_FORMAT"               envDefault:"text"`
	LogLevel               string   `env:"LOG_LEVEL"                envDefault:"info"`
	AssetsRootURL          *url.URL `env:"ASSETS_ROOT_URL,required"`
	AssetsUseManifest      bool     `env:"ASSETS_USE_MANIFEST"      envDefault:"false"`
	AssetsManifestLocation string   `env:"ASSETS_MANIFEST_LOCATION"`
}
