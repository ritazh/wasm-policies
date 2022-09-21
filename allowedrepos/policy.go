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

	// objectToTest := `{"apiVersion":"v1","kind":"Pod","metadata":{"annotations":{"kubectl.kubernetes.io/last-applied-configuration":"{\"apiVersion\":\"v1\",\"kind\":\"Pod\",\"metadata\":{\"annotations\":{},\"name\":\"test-pod1\",\"namespace\":\"default\"},\"spec\":{\"containers\":[{\"image\":\"tomcat\",\"name\":\"tomcat\"}]}}\n"},"creationTimestamp":"2022-09-21T00:20:34Z","name":"test-pod1","namespace":"default","resourceVersion":"40743","uid":"5c0a9a18-97fe-4093-b40a-8c7cad9a2a17"},"spec":{"containers":[{"image":"tomcat","imagePullPolicy":"Always","name":"tomcat","resources":{},"terminationMessagePath":"/dev/termination-log","terminationMessagePolicy":"File","volumeMounts":[{"mountPath":"/var/run/secrets/kubernetes.io/serviceaccount","name":"kube-api-access-q68kb","readOnly":true}]}],"dnsPolicy":"ClusterFirst","enableServiceLinks":true,"nodeName":"gk-control-plane","preemptionPolicy":"PreemptLowerPriority","priority":0,"restartPolicy":"Always","schedulerName":"default-scheduler","securityContext":{},"serviceAccount":"default","serviceAccountName":"default","terminationGracePeriodSeconds":30,"tolerations":[{"effect":"NoExecute","key":"node.kubernetes.io/not-ready","operator":"Exists","tolerationSeconds":300},{"effect":"NoExecute","key":"node.kubernetes.io/unreachable","operator":"Exists","tolerationSeconds":300}],"volumes":[{"name":"kube-api-access-q68kb","projected":{"defaultMode":420,"sources":[{"serviceAccountToken":{"expirationSeconds":3607,"path":"token"}},{"configMap":{"items":[{"key":"ca.crt","path":"ca.crt"}],"name":"kube-root-ca.crt"}},{"downwardAPI":{"items":[{"fieldRef":{"apiVersion":"v1","fieldPath":"metadata.namespace"},"path":"namespace"}]}}]}}]},"status":{"conditions":[{"lastProbeTime":null,"lastTransitionTime":"2022-09-21T00:20:34Z","status":"True","type":"Initialized"},{"lastProbeTime":null,"lastTransitionTime":"2022-09-21T00:20:43Z","status":"True","type":"Ready"},{"lastProbeTime":null,"lastTransitionTime":"2022-09-21T00:20:43Z","status":"True","type":"ContainersReady"},{"lastProbeTime":null,"lastTransitionTime":"2022-09-21T00:20:34Z","status":"True","type":"PodScheduled"}],"containerStatuses":[{"containerID":"containerd://ee2dd34e593f8c991a38493044b150d855608d0d131d8266d633a7ea4edf7443","image":"docker.io/library/tomcat:latest","imageID":"docker.io/library/tomcat@sha256:bb81645575fef90e48e6f9fff50e06d5b78d4ac9d2683845401164ba1ddfe199","lastState":{},"name":"tomcat","ready":true,"restartCount":0,"started":true,"state":{"running":{"startedAt":"2022-09-21T00:20:43Z"}}}],"hostIP":"172.18.0.3","phase":"Running","podIP":"10.244.0.5","podIPs":[{"ip":"10.244.0.5"}],"qosClass":"BestEffort","startTime":"2022-09-21T00:20:34Z"}}`
	// parameters := `{"imagePrefix":["tom"]}`

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
