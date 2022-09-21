# hello-wasm


To test:
```console
$ tinygo build -o greet.wasm -scheduler=none -target=wasi greet.go 

# if owner label is admin, decision is true
$ go run greet-host.go '{"apiVersion":"v1","kind":"Pod","metadata":{"labels":{"owner":"admin.agilebank.com"},"name":"test-pod1"},"spec":{"containers":[{"image":"tomcat","name":"tomcat"}]}}' param

wasm guest objectToTest >> {"apiVersion":"v1","kind":"Pod","metadata":{"labels":{"owner":"admin.agilebank.com"},"name":"test-pod1"},"spec":{"containers":[{"image":"tomcat","name":"tomcat"}]}}, parameters >> param, value >> admin.agilebank.com

host getting data from guest stdout: true

# if owner label is NOT admin, decision is false
$ go run greet-host.go '{"apiVersion":"v1","kind":"Pod","metadata":{"labels":{"owner":"admin.agilebank.com"},"name":"test-pod1"},"spec":{"containers":[{"image":"tomcat","name":"tomcat"}]}}' param

wasm guest objectToTest >> {"apiVersion":"v1","kind":"Pod","metadata":{"labels":{"owner":"user.agilebank.com"},"name":"test-pod1"},"spec":{"containers":[{"image":"tomcat","name":"tomcat"}]}}, parameters >> param, value >> admin.agilebank.com

host getting data from guest stdout: false
```