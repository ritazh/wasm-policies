//go:build wasi
// +build wasi

package main

import (
	"fmt"
	"os"
	"strings"
	"unsafe"

	"github.com/tidwall/gjson"
)

// main is required for TinyGo to compile to Wasm.
func main() {}

// eval evaluates the policy
func eval() {
	objectToTest := os.Args[1]
	parameters := os.Args[2]
	decision := true

	paramResults := gjson.Get(parameters, "imagePrefix")
	images := gjson.Get(objectToTest, "spec.containers.#.image")

	for _, image := range images.Array() {
		for _, p := range paramResults.Array() {
			decision = !strings.HasPrefix(image.String(), p.String())
			if !decision {
				break
			}
		}
	}

	fmt.Println(decision)
	log(fmt.Sprint("wasm guest objectToTest >> ", string(objectToTest), ", parameters >> ", string(parameters)))
}

// log a message to the console using _log.
func log(message string) {
	ptr, size := stringToPtr(message)
	_log(ptr, size)
}

// _log is a WebAssembly import which prints a string (linear memory offset,
// byteCount) to the console.
//
// Note: In TinyGo "//export" on a func is actually an import!
//
//go:wasm-module env
//export log
func _log(ptr uint32, size uint32)

// _eval is a WebAssembly export that accepts a string pointer (linear memory
// offset) and calls eval.
//
//export eval
func _eval() {
	eval()
}

// stringToPtr returns a pointer and size pair for the given string in a way
// compatible with WebAssembly numeric types.
func stringToPtr(s string) (uint32, uint32) {
	buf := []byte(s)
	ptr := &buf[0]
	unsafePtr := uintptr(unsafe.Pointer(ptr))
	return uint32(unsafePtr), uint32(len(buf))
}
