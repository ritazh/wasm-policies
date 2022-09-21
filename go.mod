module hello-wasm

go 1.17

require (
	github.com/bytecodealliance/wasmtime-go v0.40.0
	github.com/kubewarden/k8s-objects v1.24.0-kw3
	github.com/mailru/easyjson v0.7.7
	github.com/tetratelabs/wazero v1.0.0-pre.1
	github.com/tidwall/gjson v1.14.3
	github.com/go-interpreter/wagon v0.6.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
)

require (
	github.com/go-openapi/strfmt v0.0.0-00010101000000-000000000000 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
)

replace github.com/go-openapi/strfmt => github.com/kubewarden/strfmt v0.1.2
