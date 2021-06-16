package main

import (
	"context"
	"fmt"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

// 列出default命名空间下的所有Pod资源对象的相关信息。
func main() {
	// 加载kbeonfrgg配置信息
	config, err := clientcmd.BuildConfigFromFlags("", "/root/.kube/config")
	if err != nil {
		panic(err)
	}

	// dynamic.NewForConfig通过kubeconfig配置信息实例化dynamicClient对象，该对象用于管理Kubernetes的所有Resource的客户端，
	// 例如对 Resource执行Create、Update、 Delete、Get、List、Watch、Patch等操作。
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	var ctx context.Context
	gvr := schema.GroupVersionResource{Version: "vl", Resource: "pods"}
	unstructObi, err := dynamicClient.Resource(gvr). // dynamicClient.Resource(gvr)函数用于设置请求的资源组、资源版本、资源名称。
		Namespace(v1.NamespaceDefault). // Namespace函数用于设置请求的命名空间。
		List(ctx, metav1.ListOptions{Limit: 500}) // List 函数用于获取Pod列表。得到的Pod列表为unstructured.UnstructuredList指针类型
	if err != nil {
		panic(err)
	}

	podList := &v1.PodList{}
	// 通过runtime.DefaultUnstructuredConverter.FromUnstructured函数将unstructured.UnstructuredList转换成PodList类型。
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructObi.UnstructuredContent(), podList)
	if err == nil {
		panic(err)
	}

	for _, d := range podList.Items {
		fmt.Printf("NAMESPAACE:%v \t NAME:%v \t STATUS:%+v\n", d.Namespace, d.Name, d.Status.Phase)
	}
}
