# wasm-policies

## To test allowed repos or required labels (policies written in Go):

```console
$ cd <policy folder name such as allowedrepos or requiredlabels>
$ tinygo build -o policy.wasm -scheduler=none -target=wasi policy.go 

# if owner label prefix is admin, decision is true
$ go run host.go '{"apiVersion":"v1","kind":"Pod","metadata":{"labels":{"owner":"admin.agilebank.com"},"name":"test-pod1"},"spec":{"containers":[{"image":"tomcat","name":"tomcat"}]}}' param

wasm guest objectToTest >> {"apiVersion":"v1","kind":"Pod","metadata":{"labels":{"owner":"admin.agilebank.com"},"name":"test-pod1"},"spec":{"containers":[{"image":"tomcat","name":"tomcat"}]}}, parameters >> param, value >> admin.agilebank.com

host getting data from guest stdout: true

# if owner label prefix is NOT admin, decision is false
$ go run host.go '{"apiVersion":"v1","kind":"Pod","metadata":{"labels":{"owner":"admin.agilebank.com"},"name":"test-pod1"},"spec":{"containers":[{"image":"tomcat","name":"tomcat"}]}}' param

wasm guest objectToTest >> {"apiVersion":"v1","kind":"Pod","metadata":{"labels":{"owner":"user.agilebank.com"},"name":"test-pod1"},"spec":{"containers":[{"image":"tomcat","name":"tomcat"}]}}, parameters >> param, value >> admin.agilebank.com

host getting data from guest stdout: false
```

## To test privileged (policy written in Rust):


```console
# build the policy
$ make
...
Finished release [optimized] target(s) in 18.67s
cp target/wasm32-wasi/release/*.wasm policy.wasm

# testing
$ make run
go get github.com/tetratelabs/wazero
go run host.go '{"apiVersion":"v1","kind":"Pod","metadata":{"name":"nginx","labels":{"app":"nginx"}},"spec":{"containers":[{"name":"nginx","image":"nginx","securityContext":{"privileged":true}}]}}' '' 
host getting data from guest stdout: false
```
