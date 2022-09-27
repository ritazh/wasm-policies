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

// greetWasm was compiled using `tinygo build -o greet.wasm -scheduler=none --no-debug -target=wasi greet.go`
//
//go:embed testdata/greet.wasm
var greetWasm []byte

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

	// Instantiate a WebAssembly module that imports the "log" function defined
	// in "env" and exports "memory" and functions we'll use in this example.
	// mod, err := r.InstantiateModuleFromBinary(ctx, greetWasm)
	// if err != nil {
	// 	log.Panicln(err)
	// }

	// Compile the WebAssembly module using the default configuration.
	code, err := r.CompileModule(ctx, greetWasm, wazero.NewCompileConfig())
	if err != nil {
		log.Panicln(err)
	}
	mod, err := r.InstantiateModule(ctx, code, config.WithArgs("wasi", os.Args[1], os.Args[2]))
	if err != nil {
		log.Panicln(err)
	}

	// Get references to WebAssembly functions we'll use in this example.
	greet := mod.ExportedFunction("greet")
	//greeting := mod.ExportedFunction("greeting")
	// These are undocumented, but exported. See tinygo-org/tinygo#2788
	// malloc := mod.ExportedFunction("malloc")
	// free := mod.ExportedFunction("free")

	// Let's use the argument to this main function in Wasm.
	// name := os.Args[1]
	// nameSize := uint64(len(name))

	// Instead of an arbitrary memory offset, use TinyGo's allocator. Notice
	// there is nothing string-specific in this allocation function. The same
	// function could be used to pass binary serialized data to Wasm.
	// results, err := malloc.Call(ctx)
	// if err != nil {
	// 	log.Panicln(err)
	// }
	// namePtr := results[0]
	// // This pointer is managed by TinyGo, but TinyGo is unaware of external usage.
	// // So, we have to free it when finished
	// defer free.Call(ctx, namePtr)

	// The pointer is a linear memory offset, which is where we write the name.
	// if !mod.Memory().Write(ctx, uint32(namePtr), []byte(name)) {
	// 	log.Panicf("Memory.Write(%d, %d) out of range of memory size %d",
	// 		namePtr, nameSize, mod.Memory().Size(ctx))
	// }

	// Now, we can call "greet", which reads the string we wrote to memory!
	_, err = greet.Call(ctx)
	if err != nil {
		log.Panicln(err)
	}
	decision := stdout.Bytes()
	fmt.Println("host getting data from guest stdout: " + string(decision))

	// // Finally, we get the greeting message "greet" printed. This shows how to
	// // read-back something allocated by TinyGo.
	// ptrSize, err := greeting.Call(ctx, namePtr, nameSize)
	// if err != nil {
	// 	log.Panicln(err)
	// }
	// // Note: This pointer is still owned by TinyGo, so don't try to free it!
	// greetingPtr := uint32(ptrSize[0] >> 32)
	// greetingSize := uint32(ptrSize[0])
	// // The pointer is a linear memory offset, which is where we write the name.
	// if bytes, ok := mod.Memory().Read(ctx, greetingPtr, greetingSize); !ok {
	// 	log.Panicf("Memory.Read(%d, %d) out of range of memory size %d",
	// 		greetingPtr, greetingSize, mod.Memory().Size(ctx))
	// } else {
	// 	fmt.Println("go >>", string(bytes))
	// }
}

func logString(ctx context.Context, m api.Module, offset, byteCount uint32) {
	buf, ok := m.Memory().Read(ctx, offset, byteCount)
	if !ok {
		log.Panicf("Memory.Read(%d, %d) out of range", offset, byteCount)
	}
	fmt.Println(string(buf))
}
