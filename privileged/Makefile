SOURCE_FILES := $(shell test -e src/ && find src -type f)

policy.wasm: $(SOURCE_FILES) Cargo.*
	# rustup target add wasm32-wasi 
	# rustup target add wasm32-unknown-unknown
	# cargo build --target=wasm32-wasi --release
	# cp target/wasm32-wasi/release/*.wasm policy.wasm
	cargo build --release --target wasm32-wasi
	cp target/wasm32-wasi/release/*.wasm policy.wasm

.PHONY: run
run:
	go get github.com/tetratelabs/wazero
	go run host.go test1 test2 

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