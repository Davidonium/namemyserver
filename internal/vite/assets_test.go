package vite_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/davidonium/namemyserver/internal/vite"
	"github.com/stretchr/testify/assert"
)

func testManifestJSON() string {
	return `{
    "main.js": {
      "file": "assets/main.123.js",
      "isEntry": true,
      "imports": ["chunk-vendors.456.js"],
      "css": ["assets/main.123.css"]
    },
    "chunk-vendors.456.js": {
      "file": "assets/chunk-vendors.456.js",
      "imports": [],
      "css": ["assets/vendors.456.css"]
    }
  }`
}

func TestNewAssets_RootURL(t *testing.T) {
	tests := []struct {
		name     string
		config   vite.AssetsConfig
		expected string
	}{
		{"with trailing slash", vite.AssetsConfig{RootURL: "/static/", UseManifest: false}, "/static/foo.js"},
		{"without trailing slash", vite.AssetsConfig{RootURL: "/static", UseManifest: false}, "/static/foo.js"},
		{"empty root", vite.AssetsConfig{RootURL: "", UseManifest: false}, "/foo.js"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := vite.NewAssets(tt.config)
			assert.Equal(t, tt.expected, a.RenderAsset("foo.js"))
		})
	}
}

func TestLoadManifest(t *testing.T) {
	a := vite.NewAssets(vite.AssetsConfig{RootURL: "/", UseManifest: true})
	err := a.LoadManifest(strings.NewReader(testManifestJSON()))
	assert.NoError(t, err)
	// Should render CSS and JS from manifest
	assert.Contains(t, a.RenderCSS("main.js"), "main.123.css")
	assert.Contains(t, a.RenderJS("main.js"), "main.123.js")
}

func TestLoadManifestFromFile_NotFound(t *testing.T) {
	a := vite.NewAssets(vite.AssetsConfig{RootURL: "/", UseManifest: true})
	err := a.LoadManifestFromFile("notfound.json")
	assert.Error(t, err)
}

func TestLoadManifestFromFS(t *testing.T) {
	fs := fstest.MapFS{
		"manifest.json": {Data: []byte(testManifestJSON())},
	}
	a := vite.NewAssets(vite.AssetsConfig{RootURL: "/", UseManifest: true})
	err := a.LoadManifestFromFS(fs, "manifest.json")
	assert.NoError(t, err)
	assert.Contains(t, a.RenderCSS("main.js"), "main.123.css")
}

func TestRenderViteClientJS(t *testing.T) {
	a := vite.NewAssets(vite.AssetsConfig{RootURL: "/static/", UseManifest: false})
	assert.Contains(t, a.RenderViteClientJS(), "@vite/client")

	a2 := vite.NewAssets(vite.AssetsConfig{RootURL: "/static/", UseManifest: true})
	assert.Empty(t, a2.RenderViteClientJS())
}

func TestRenderTags(t *testing.T) {
	a := vite.NewAssets(vite.AssetsConfig{RootURL: "/static/", UseManifest: true})
	_ = a.LoadManifest(strings.NewReader(testManifestJSON()))
	out := a.RenderTags([]string{"main.js"})
	assert.Equal(t, `<link type="text/css" rel="stylesheet" href="/static/assets/main.123.css" />
<link type="text/css" rel="stylesheet" href="/static/assets/vendors.456.css" />
<script type="module" src="/static/assets/main.123.js"></script>
<link rel="modulepreload" href="/static/assets/chunk-vendors.456.js" />
`, out)
}

func TestRenderCSS(t *testing.T) {
	a := vite.NewAssets(vite.AssetsConfig{RootURL: "/static/", UseManifest: true})
	_ = a.LoadManifest(strings.NewReader(testManifestJSON()))
	css := a.RenderCSS("main.js")
	expected := `<link type="text/css" rel="stylesheet" href="/static/assets/main.123.css" />
<link type="text/css" rel="stylesheet" href="/static/assets/vendors.456.css" />
`
	assert.Equal(t, expected, css)

	a2 := vite.NewAssets(vite.AssetsConfig{RootURL: "/static/", UseManifest: false})
	assert.Empty(t, a2.RenderCSS("main.js"))
}

func TestRenderJS(t *testing.T) {
	a := vite.NewAssets(vite.AssetsConfig{RootURL: "/static/", UseManifest: true})
	_ = a.LoadManifest(strings.NewReader(testManifestJSON()))
	js := a.RenderJS("main.js")
	expected := `<script type="module" src="/static/assets/main.123.js"></script>
<link rel="modulepreload" href="/static/assets/chunk-vendors.456.js" />
`
	assert.Equal(t, expected, js)

	a2 := vite.NewAssets(vite.AssetsConfig{RootURL: "/static/", UseManifest: false})
	js2 := a2.RenderJS("main.js")
	expected2 := `<script type="module" src="/static/main.js"></script>
`
	assert.Equal(t, expected2, js2)
}

func TestRenderAsset(t *testing.T) {
	a := vite.NewAssets(vite.AssetsConfig{RootURL: "/static/", UseManifest: false})
	assert.Equal(t, "/static/foo.js", a.RenderAsset("foo.js"))
}

func TestContextHelpers(t *testing.T) {
	a := vite.NewAssets(vite.AssetsConfig{RootURL: "/static/", UseManifest: true})
	_ = a.LoadManifest(strings.NewReader(testManifestJSON()))
	ctx := context.Background()
	ctx = vite.NewContextWithAssets(ctx, a)
	assert.Equal(t, "/static/foo.js", vite.AssetURL(ctx, "foo.js"))
	assert.Contains(t, vite.RenderTags(ctx, []string{"main.js"}), "main.123.js")
}

func TestLoadManifest_InvalidJSON(t *testing.T) {
	a := vite.NewAssets(vite.AssetsConfig{RootURL: "/", UseManifest: true})
	err := a.LoadManifest(strings.NewReader("{invalid json"))
	assert.Error(t, err)
}

func TestLoadManifest_EmptyReader(t *testing.T) {
	a := vite.NewAssets(vite.AssetsConfig{RootURL: "/", UseManifest: true})
	err := a.LoadManifest(io.NopCloser(bytes.NewReader([]byte{})))
	assert.Error(t, err)
}
