package main

import (
	"fmt"
	"os"
	"strings"
	"unsafe"

	corev1 "github.com/kubewarden/k8s-objects/api/core/v1"
	"github.com/mailru/easyjson"
	"github.com/tidwall/gjson"
)

// main is required for TinyGo to compile to Wasm.
func main() {}

// eval evaluates the policy
func eval() {
	objectToTest := os.Args[1]
	parameters := os.Args[2]

	// objectToTest = `{"apiVersion":"v1","kind":"Pod","metadata":{"name":"test-pod1"},"spec":{"containers":[{"image":"tomcat","name":"tomcat"}]}}`
	// parameters = `{"imagePrefix":["tom"]}`

	paramResults := gjson.Get(parameters, "imagePrefix")

	pod := &corev1.Pod{}
	if err := easyjson.Unmarshal([]byte(objectToTest), pod); err != nil {
		panic(err)
	}

	decision := false
	for _, container := range pod.Spec.Containers {
		for _, p := range paramResults.Array() {
			decision = strings.HasPrefix(container.Image, p.String())
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
