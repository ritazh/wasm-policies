package main

import (
    "fmt"
    "github.com/bytecodealliance/wasmtime-go"
)

func main() {
    // to run 
    // tinygo build -o main.wasm -target wasi main.go
    // go run main-tinygo.go 
    engine := wasmtime.NewEngine()
    store := wasmtime.NewStore(engine)
    module, err := wasmtime.NewModuleFromFile(engine, "main.wasm")
    check(err)
    linker := wasmtime.NewLinker(store.Engine)
    err = linker.DefineWasi()
    check(err) 
    wasiConfig := wasmtime.NewWasiConfig()
	wasiConfig.InheritEnv()
	wasiConfig.PreopenDir(".", ".")
    store.SetWasi(wasiConfig)

    instance, err := linker.Instantiate(store, module)
    check(err)

    main := instance.GetExport(store, "_start").Func()
    _, err = main.Call(store)
    check(err)
    fmt.Printf("call finished \n")
}

func check(err error) {
    if err != nil {
        panic(err)
    }
}

