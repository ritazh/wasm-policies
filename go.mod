module hello-wasm

go 1.17

require (
	github.com/bytecodealliance/wasmtime-go v0.40.0
	github.com/tetratelabs/wazero v1.0.0-pre.1
	github.com/tidwall/gjson v1.14.3
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
)

replace github.com/go-openapi/strfmt => github.com/kubewarden/strfmt v0.1.2
