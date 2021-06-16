package main

import (
	"context"
	"fmt"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// 列出default命名空间下的所有Pod资源对象的相关信息。
func main() {
	// 加载kubeconfig配置信息
	config, err := clientcmd.BuildConfigFromFlags("", "/root/.kube/config")
	if err != nil {
		panic(err)
	}

	// kubernetes.NewForConfig通过kubeconfig配置信息实例化 clientset对象，该对象用于管理所有Resource的客户端。
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	var ctx context.Context
	// clientset.CoreV1().Pods函数表示请求core核心资源组的v1资源版本下的Pod资源对象，
	// 其内部设置了APIPath请求的HTTP路径，GroupVersion 请求的资源组、资源版本，NegotiatedSerializer 数据的编解码器。
	podClient := clientset.CoreV1().Pods(apiv1.NamespaceDefault)
	// 其中，Pods函数是一个资源接口对象，用于Pod资源对象的管理，例如，对Pod 资源执行Create、Update、 Delete、 Get、 List、 Watch、 Patch 等操作，
	// 这些操作实际上是对RESTClient进行了封装，可以设置选项(如Limit、 TimeoutSeconds 等)
	list, err := podClient.List(ctx, metav1.ListOptions{Limit: 500})
	if err != nil {
		panic(err)
	}

	for _, d := range list.Items {
		fmt.Printf("NAMESPACE:%v \t NAME:%v \t STATU:%+v\n", d.Namespace, d.Name, d.Status.Phase)
	}
}
