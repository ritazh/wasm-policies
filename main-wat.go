package main

import (
    "fmt"
    "github.com/bytecodealliance/wasmtime-go"
)

func main() {
    engine := wasmtime.NewEngine()
    store := wasmtime.NewStore(engine)
    module, err := wasmtime.NewModuleFromFile(engine, "gcd.wat")
    check(err)
    instance, err := wasmtime.NewInstance(store, module, []wasmtime.AsExtern{})
    check(err)

    gcd := instance.GetExport(store, "gcd").Func()
    val, err := gcd.Call(store, 6, 27)
    check(err)
    fmt.Printf("gcd(6, 27) = %d\n", val.(int32))
}

func check(err error) {
    if err != nil {
        panic(err)
    }
}

