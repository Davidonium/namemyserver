package vite

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"sync"
)

type ManifestEntry struct {
	File           string   `json:"file"`
	Src            string   `json:"src"`
	IsEntry        bool     `json:"isEntry"`
	IsDynamicEntry bool     `json:"isDynamicEntry"`
	Imports        []string `json:"imports"`
	DynamicImports []string `json:"dynamicImports"`
	CSS            []string `json:"css"`
	Assets         []string `json:"assets"`
}

type Assets struct {
	rootURL      string
	useManifest  bool
	manifest     map[string]ManifestEntry
	manifestLock *sync.RWMutex
}

type AssetsConfig struct {
	RootURL     string
	UseManifest bool
}

func NewAssets(config AssetsConfig) *Assets {
	// root url should always end with / because entries are declared without leading slash
	rootURL := config.RootURL
	if !strings.HasSuffix(config.RootURL, "/") {
		rootURL += "/"
	}

	return &Assets{
		rootURL:      rootURL,
		useManifest:  config.UseManifest,
		manifestLock: &sync.RWMutex{},
	}
}

func (v *Assets) LoadManifest(r io.Reader) error {
	v.manifestLock.Lock()
	defer v.manifestLock.Unlock()

	return json.NewDecoder(r).Decode(&v.manifest)
}

func (v *Assets) LoadManifestFromFile(location string) error {
	fd, err := os.Open(location)
	if err != nil {
		return fmt.Errorf("failed to open assets manifest file: %w", err)
	}

	defer fd.Close()

	return v.LoadManifest(fd)
}

func (v *Assets) LoadManifestFromFS(fs fs.FS, location string) error {
	fd, err := fs.Open(location)
	if err != nil {
		return fmt.Errorf("failed to open assets manifest file: %w", err)
	}

	defer fd.Close()

	return v.LoadManifest(fd)
}

func (v *Assets) RenderViteClientJS() string {
	if !v.useManifest {
		return fmt.Sprintf(`<script type="module" src=%q></script>`+"\n", v.rootURL+"@vite/client")
	}

	return ""
}

func (v *Assets) RenderTags(entries []string) string {
	buf := &strings.Builder{}

	for _, e := range entries {
		buf.WriteString(v.RenderCSS(e))
	}

	buf.WriteString(v.RenderViteClientJS())

	for _, e := range entries {
		buf.WriteString(v.RenderJS(e))
	}

	return buf.String()
}

func (v *Assets) RenderCSS(entry string) string {
	// avoid rendering css tags in dev, they are loaded by javascript in vite
	if !v.useManifest {
		return ""
	}

	v.manifestLock.RLock()
	manifestEntry, ok := v.manifest[entry]
	v.manifestLock.RUnlock()
	if !ok {
		// TODO log this and handle differently
		return ""
	}

	buf := &strings.Builder{}

	for _, cssEntry := range manifestEntry.CSS {
		buf.WriteString(fmt.Sprintf(`<link type="text/css" rel="stylesheet" href=%q />`+"\n", v.rootURL+cssEntry))
	}

	imports := manifestEntry.Imports
	imports = append(imports, manifestEntry.DynamicImports...)

	for _, im := range imports {
		v.manifestLock.RLock()
		chunkEntry, ok := v.manifest[im]
		v.manifestLock.RUnlock()

		if !ok {
			continue
		}

		for _, cssEntry := range chunkEntry.CSS {
			buf.WriteString(fmt.Sprintf(`<link type="text/css" rel="stylesheet" href=%q />`+"\n", v.rootURL+cssEntry))
		}
	}

	return buf.String()
}

func (v *Assets) RenderJS(entry string) string {
	if !v.useManifest {
		return fmt.Sprintf(`<script type="module" src=%q></script>`+"\n", v.rootURL+entry)
	}

	buf := &strings.Builder{}

	v.manifestLock.RLock()
	manifestEntry, ok := v.manifest[entry]
	v.manifestLock.RUnlock()
	if !ok {
		// TODO log this and handle differently
		return ""
	}

	file := v.rootURL + manifestEntry.File

	buf.WriteString(fmt.Sprintf(`<script type="module" src=%q></script>`+"\n", file))

	imports := manifestEntry.Imports
	imports = append(imports, manifestEntry.DynamicImports...)

	for _, im := range imports {
		v.manifestLock.RLock()
		chunkEntry, ok := v.manifest[im]
		v.manifestLock.RUnlock()

		if !ok {
			continue
		}

		// not sure if modulepreload can be set for static imports, it's not clarified or exemplified in the docs
		if chunkEntry.IsDynamicEntry {
			buf.WriteString(fmt.Sprintf(`<links rel="modulepreload" href=%q />`+"\n", file))
		}
	}

	return buf.String()
}

func (v *Assets) RenderAsset(location string) string {
	return v.rootURL + location
}

type contextKey struct{}

var assetsContextKey = contextKey{}

func AssetURL(ctx context.Context, location string) string {
	return GetAssets(ctx).RenderAsset(location)
}

func RenderTags(ctx context.Context, entries []string) string {
	return GetAssets(ctx).RenderTags(entries)
}

func GetAssets(ctx context.Context) *Assets {
	return ctx.Value(assetsContextKey).(*Assets)
}

func NewContextWithAssets(ctx context.Context, a *Assets) context.Context {
	return context.WithValue(ctx, assetsContextKey, a)
}
