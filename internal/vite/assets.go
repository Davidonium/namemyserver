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
	"time"
)

const (
	AssetManifestFSOS    = "os"
	AssetManifestFSEmbed = "embed"
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

type AssetsConfig struct {
	RootURL          string
	UseManifest      bool
	ManifestLocation string
	WatchForChanges  bool
	AssetsFS         fs.FS
}

type Logger interface {
	Info(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
}

type Assets struct {
	logger Logger

	rootURL          string
	useManifest      bool
	manifestLocation string
	watchEnabled     bool
	assetsFS         fs.FS

	manifest     map[string]ManifestEntry
	manifestLock *sync.RWMutex
	watchCancel  context.CancelFunc
	lastModTime  time.Time
}

func NewAssets(logger Logger, config AssetsConfig) *Assets {
	// root url should always end with / because entries are declared without leading slash
	rootURL := config.RootURL
	if !strings.HasSuffix(config.RootURL, "/") {
		rootURL += "/"
	}

	if config.AssetsFS == nil {
		config.AssetsFS = os.DirFS(".")
	}

	a := &Assets{
		rootURL:          rootURL,
		useManifest:      config.UseManifest,
		watchEnabled:     config.WatchForChanges,
		manifestLocation: config.ManifestLocation,
		assetsFS:         config.AssetsFS,
		logger:           logger,

		manifestLock: &sync.RWMutex{},
	}

	return a
}

func (v *Assets) LoadManifestFromReader(r io.Reader) error {
	v.manifestLock.Lock()
	defer v.manifestLock.Unlock()

	return json.NewDecoder(r).Decode(&v.manifest)
}

func (v *Assets) LoadManifest() error {
	return v.LoadManifestFromFS(v.assetsFS, v.manifestLocation)
}

func (v *Assets) LoadManifestFromFS(fs fs.FS, location string) error {
	fd, err := fs.Open(location)
	if err != nil {
		return fmt.Errorf("failed to open assets manifest file: %w", err)
	}

	defer fd.Close()

	return v.LoadManifestFromReader(fd)
}

// WatchManifest begins watching the manifest file for changes
func (v *Assets) WatchManifest(ctx context.Context) {
	if !v.watchEnabled {
		return
	}

	watchCtx, cancel := context.WithCancel(ctx)
	v.watchCancel = cancel

	go v.watchManifestFile(watchCtx)
}

// StopWatching stops the file watching
func (v *Assets) Close() {
	if v.watchCancel != nil {
		v.watchCancel()
		v.watchCancel = nil
	}
}

// watchManifestFile polls the manifest file for changes
func (v *Assets) watchManifestFile(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:

			if stat, err := fs.Stat(v.assetsFS, v.manifestLocation); err == nil {
				if stat.ModTime().After(v.lastModTime) {
					v.logger.Info(
						"detected change in assets manifest, reloading",
						"file", v.manifestLocation,
					)
					v.lastModTime = stat.ModTime()
					if err := v.LoadManifest(); err != nil {
						v.logger.Error("failed to reload manifest", "error", err)
					}
				}
			}
		}
	}
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

	imported := v.ImportedChunks(entry)

	for _, cssEntry := range manifestEntry.CSS {
		fmt.Fprintf(
			buf,
			`<link type="text/css" rel="stylesheet" href=%q />`+"\n",
			v.rootURL+cssEntry,
		)
	}

	for _, imp := range imported {
		for _, cssEntry := range imp.CSS {
			fmt.Fprintf(
				buf,
				`<link type="text/css" rel="stylesheet" href=%q />`+"\n",
				v.rootURL+cssEntry,
			)
		}
	}

	return buf.String()
}

func (v *Assets) RenderJS(entry string) string {
	if !v.useManifest {
		return fmt.Sprintf(`<script type="module" src=%q></script>`+"\n", v.rootURL+entry)
	}

	buf := &strings.Builder{}

	manifestEntry, ok := v.manifestEntry(entry)
	if !ok {
		// TODO log this and handle differently
		return ""
	}

	file := v.rootURL + manifestEntry.File
	fmt.Fprintf(buf, `<script type="module" src=%q></script>`+"\n", file)

	imported := v.ImportedChunks(entry)
	for _, imp := range imported {
		fmt.Fprintf(buf, `<link rel="modulepreload" href=%q />`+"\n", v.rootURL+imp.File)
	}

	return buf.String()
}

func (v *Assets) RenderAsset(location string) string {
	return v.rootURL + location
}

// ImportedChunks uses an iterative depth-first search (DFS) with a stack.
// It preserves the post-order traversal logic of the original recursive function,
// where a chunk is added to the result after all its transitive imports.
func (v *Assets) ImportedChunks(name string) []ManifestEntry {
	type StackItem struct {
		Chunk          ManifestEntry
		ChildrenPushed bool
	}

	seen := make(map[string]struct{})
	stack := []StackItem{}
	imported := []ManifestEntry{}

	initialChunk, ok := v.manifestEntry(name)
	if !ok {
		return []ManifestEntry{}
	}

	// Only add children (imports) to the stack, not the initial chunk itself
	for i := len(initialChunk.Imports) - 1; i >= 0; i-- {
		file := initialChunk.Imports[i]
		if _, visited := seen[file]; !visited {
			if importee, ok := v.manifestEntry(file); ok {
				stack = append(stack, StackItem{Chunk: importee, ChildrenPushed: false})
				seen[file] = struct{}{}
			}
		}
	}

	for len(stack) > 0 {
		topIndex := len(stack) - 1
		currentItem := stack[topIndex]

		if !currentItem.ChildrenPushed {
			stack[topIndex].ChildrenPushed = true

			foundUnvisitedChild := false
			for i := len(currentItem.Chunk.Imports) - 1; i >= 0; i-- {
				file := currentItem.Chunk.Imports[i]
				if _, visited := seen[file]; !visited {
					if importee, ok := v.manifestEntry(file); ok {
						stack = append(stack, StackItem{Chunk: importee, ChildrenPushed: false})
						seen[file] = struct{}{}
						foundUnvisitedChild = true
					}
				}
			}
			if foundUnvisitedChild {
				continue
			}
		}

		stack = stack[:topIndex]
		imported = append(imported, currentItem.Chunk)
	}

	return imported
}

func (v *Assets) manifestEntry(name string) (ManifestEntry, bool) {
	v.manifestLock.RLock()
	defer v.manifestLock.RUnlock()
	manifestEntry, ok := v.manifest[name]
	return manifestEntry, ok
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
