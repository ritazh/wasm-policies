# hello-wasm


To test:
```console
tinygo build -o greet.wasm -scheduler=none -target=wasi greet.go && go run greet-host.go obj param

wasm guest objectToTest >> obj, parameters >> param
host getting data from guest stdout: true
```