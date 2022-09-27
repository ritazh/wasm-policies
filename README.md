# wasm-policies

Wouldn't it be great if you could write cloud native policies as code in your favorite programming language? Let's create flexible, secure, and portable policies with WebAssembly!

To run these policies with Gatekeeper, please refer to the experimental Gatekeeper (Kubernetes admission webhook) with Wasm support:
https://github.com/ritazh/gatekeeper/tree/wasm

## To test policies written in Go (allowed repos or required labels):

### Required Labels
Allow k8s pod creation if owner label prefix is admin

```console
# build the policy
$ cd requiredlabels
$ tinygo build -o policy.wasm -scheduler=none -target=wasi -no-debug policy.go 

# test from host, if owner label prefix is admin, decision is true
$ go run host.go '{"apiVersion":"v1","kind":"Pod","metadata":{"labels":{"owner":"admin.agilebank.com"},"name":"test-pod1"},"spec":{"containers":[{"image":"tomcat","name":"tomcat"}]}}' param

host getting data from guest stdout: true

# if owner label prefix is NOT admin, decision is false
$ go run host.go '{"apiVersion":"v1","kind":"Pod","metadata":{"labels":{"owner":"admin.agilebank.com"},"name":"test-pod1"},"spec":{"containers":[{"image":"tomcat","name":"tomcat"}]}}' param

host getting data from guest stdout: false
```

## To test policy written in Rust (privileged):

### Privileged
Allow k8s pod creation if pod is not privileged

```console
# build the policy
$ cd privileged
$ make
...
Finished release [optimized] target(s) in 18.67s
cp target/wasm32-wasi/release/*.wasm policy.wasm

# testing from host, if securityContext of container is privileged, decision is false
$ make run
go get github.com/tetratelabs/wazero
go run host.go '{"apiVersion":"v1","kind":"Pod","metadata":{"name":"nginx","labels":{"app":"nginx"}},"spec":{"containers":[{"name":"nginx","image":"nginx","securityContext":{"privileged":true}}]}}' '' 

host getting data from guest stdout: false
```
