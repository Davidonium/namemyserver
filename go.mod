module github.com/davidonium/serverplate

go 1.26

require (
	github.com/a-h/templ v0.3.1001
	github.com/amacneil/dbmate/v2 v2.26.0
	github.com/caarlos0/env/v11 v11.3.1
	github.com/doug-martin/goqu/v9 v9.19.0
	github.com/dustin/go-humanize v1.0.1
	github.com/getkin/kin-openapi v0.133.0
	github.com/jmoiron/sqlx v1.4.0
	github.com/joho/godotenv v1.5.1
	github.com/mattn/go-sqlite3 v1.14.37
	github.com/oapi-codegen/runtime v1.3.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/stretchr/testify v1.11.1
	golang.org/x/text v0.33.0
)

require (
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-openapi/jsonpointer v0.22.4 // indirect
	github.com/go-openapi/swag/jsonname v0.25.4 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/mailru/easyjson v0.9.1 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/oasdiff/yaml v0.0.0-20250309154309-f31be36b4037 // indirect
	github.com/oasdiff/yaml3 v0.0.0-20250309153720-d2182401db90 // indirect
	github.com/perimeterx/marshmallow v1.1.5 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/woodsbury/decimal128 v1.4.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// Avoid YAML v3/v4 incompatibility.
// The newer version (v0.0.0-20251217220025-0b8845c5554e) requires go.yaml.in/yaml/v4,
// which conflicts with gopkg.in/yaml.v3 used by vmware-labs/yaml-jsonpath (required by oapi-codegen).
// This causes compilation errors in yaml-jsonpath. Keep this replace directive until
// yaml-jsonpath is updated to support yaml v4 or the dependency chain is resolved upstream.
replace github.com/dprotaso/go-yit => github.com/dprotaso/go-yit v0.0.0-20220510233725-9ba8df137936
