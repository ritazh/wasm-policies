package main

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// greetWasm was compiled using `tinygo build -o policy.wasm -scheduler=none --no-debug -target=wasi policy.go`
//
//go:embed policy.wasm
var policyWasm []byte

// main shows how to interact with a WebAssembly function that was compiled
// from TinyGo.
//
// See README.md for a full description.
func main() {
	// Choose the context to use for function calls.
	ctx := context.Background()

	stdout := bytes.NewBuffer(nil)

	// Create a new WebAssembly Runtime.
	c := wazero.NewRuntimeConfig().
		WithFeatureBulkMemoryOperations(true).WithFeatureSignExtensionOps(true) // not sure why we need this but got this error: memory.copy invalid as feature "bulk-memory-operations" is disabled
	r := wazero.NewRuntimeWithConfig(ctx, c)

	defer r.Close(ctx) // This closes everything this Runtime created.

	config := wazero.NewModuleConfig().
		// By default, I/O streams are discarded and there's no file system.
		WithStdout(stdout).WithStderr(os.Stderr)

	// Instantiate a Go-defined module named "env" that exports a function to
	// log to the console.
	_, err := r.NewModuleBuilder("env").
		ExportFunction("log", logString).
		Instantiate(ctx, r)
	if err != nil {
		log.Panicln(err)
	}

	// Note: testdata/greet.go doesn't use WASI, but TinyGo needs it to
	// implement functions such as panic.
	if _, err = wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		log.Panicln(err)
	}

	// Compile the WebAssembly module using the default configuration.
	code, err := r.CompileModule(ctx, policyWasm, wazero.NewCompileConfig())
	if err != nil {
		log.Panicln(err)
	}
	mod, err := r.InstantiateModule(ctx, code, config.WithArgs("wasi", os.Args[1], os.Args[2]))
	if err != nil {
		log.Panicln(err)
	}

	// Get references to WebAssembly functions we'll use in this example.
	eval := mod.ExportedFunction("eval")

	// Now, we can call "eval", which reads the string we wrote to memory!
	_, err = eval.Call(ctx)
	if err != nil {
		log.Panicln(err)
	}
	decision := stdout.Bytes()
	fmt.Println("host getting data from guest stdout: " + string(decision))
}

func logString(ctx context.Context, m api.Module, offset, byteCount uint32) {
	buf, ok := m.Memory().Read(ctx, offset, byteCount)
	if !ok {
		log.Panicf("Memory.Read(%d, %d) out of range", offset, byteCount)
	}
	fmt.Println(string(buf))
}
