package main

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/tools/clientcmd"
)

// 列出Kubernetes API Server所支持的资源组、资源版本、资源信息。
func main() {
	// 加载kubeconfig 配置信息
	config, err := clientcmd.BuildConfigFromFlags("", "/root/.kube/config")
	if err != nil {
		panic(err)
	}

	// discovery.NewDiscoveryClientForConfig通过kubeconfig配置信息实例化discoveryClient对象，
	// 该对象是用于发现Kubernetes API Server所支持的资源组、资源版本、资源信息的客户端。
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		panic(err)
	}

	// discoveryClient.ServerGroupsAndResources函数会返回Kubernetes API Server 所支持的资源组、资源版本、资源信息(即APIResourceList)
	_, APIResourceList, err := discoveryClient.ServerGroupsAndResources()
	if err != nil {
		panic(err)
	}

	// 通过遍历APIResourceList输出信息。
	for _, list := range APIResourceList {
		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			panic(err)
		}
		for _, resource := range list.APIResources {
			fmt.Printf("name: %v，group: %v, version: %v\n", resource.Name, gv.Group, gv.Version)
		}
	}
}
