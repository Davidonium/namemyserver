package env

import "net/url"

type Config struct {
	ListenAddr             string   `env:"LISTEN_ADDR"              envDefault:":8080"`
	DatabaseURL            *url.URL `env:"DATABASE_URL,required"`
	Debug                  bool     `env:"DEBUG"                    envDefault:"false"`
	LogFormat              string   `env:"LOG_FORMAT"               envDefault:"text"`
	LogLevel               string   `env:"LOG_LEVEL"                envDefault:"info"`
	AssetsRootURL          *url.URL `env:"ASSETS_ROOT_URL,required"`
	AssetsUseManifest      bool     `env:"ASSETS_MANIFEST_USE"      envDefault:"false"`
	AssetsWatch            bool     `env:"ASSETS_MANIFEST_WATCH"    envDefault:"false"`
	AssetsManifestLocation string   `env:"ASSETS_MANIFEST_LOCATION"`
	AssetsManifestFS       string   `env:"ASSETS_MANIFEST_FS"       envDefault:"os"`
}
