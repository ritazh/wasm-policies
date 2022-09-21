SOURCE_FILES := $(shell test -e src/ && find src -type f)

policy.wasm: $(SOURCE_FILES) Cargo.*
	# cargo build --target=wasm32-wasi --release
	# cp target/wasm32-wasi/release/*.wasm policy.wasm
	cargo build --release --target wasm32-unknown-unknown
	cp target/wasm32-unknown-unknown/release/*.wasm policy.wasm

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