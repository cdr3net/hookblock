module github.com/dbolotin/deadmanswitch

go 1.14

require (
	github.com/hashicorp/hcl/v2 v2.6.0
	github.com/prometheus/client_golang v1.6.0 // indirect
	github.com/zclconf/go-cty v1.2.0
	gopkg.in/yaml.v2 v2.2.5 // indirect
)

replace github.com/hashicorp/hcl/v2 => ./hcl
