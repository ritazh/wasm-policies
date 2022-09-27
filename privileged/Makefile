SOURCE_FILES := $(shell test -e src/ && find src -type f)

policy.wasm: $(SOURCE_FILES) Cargo.*
	rustup target add wasm32-wasi 
	cargo build --release --target wasm32-wasi
	cp target/wasm32-wasi/release/*.wasm policy.wasm

.PHONY: run
run:
	go get github.com/tetratelabs/wazero
	go run host.go '{"apiVersion":"v1","kind":"Pod","metadata":{"name":"nginx","labels":{"app":"nginx"}},"spec":{"containers":[{"name":"nginx","image":"nginx","securityContext":{"privileged":true}}]}}' '' 

.PHONY: fmt
fmt:
	cargo fmt --all -- --check

.PHONY: lint
lint:
	cargo clippy -- -D warnings

.PHONY: test
test: fmt lint
	cargo test

.PHONY: clean
clean:
	cargo clean