package main

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// 列出default命名空间下的所有Pod资源对象的相关信息。
func main() {
	// 加载kubeconfig配置信息
	config, err := clientcmd.BuildConfigFromFlags("", "/root/.kube/config")
	if err != nil {
		panic(err)
	}

	// 设置config.APIPath请求的HTTP路径。
	config.APIPath = "api"
	// 设置config.GroupVersion请求的资源组/资源版本。
	config.GroupVersion = &corev1.SchemeGroupVersion
	// 设置config.NegotiatedSerializer数据的编解码器。
	config.NegotiatedSerializer = scheme.Codecs

	// rest.RESTClientFor函数通过kubeconfig 配置信息实例化RESTClient对象，
	restClient, err := rest.RESTClientFor(config)
	if err != nil {
		panic(err)
	}

	result := &corev1.PodList{}
	var ctx context.Context
	err = restClient.Get(). // RESTClient对象构建HTTP请求参数，例如Get函数设置请求方法为get操作，它还支持Post、Put、Delete、Patch等请求方法。
		Namespace("default"). // Namespace函数设置请求的命名空间。
		Resource("pods"). // Resource函数设置请求的资源名称。
		VersionedParams(&metav1.ListOptions{Limit: 500}, scheme.ParameterCodec). // VersionedParams函数将一些查询选项(如limit、TimeoutSeconds等)添加到请求参数中。
		Do(ctx). // 通过Do函数执行该请求
		Into(result) // 将kube-apiserver 返回的结果(Result 对象)解析到corev1.PodList对象中
	if err != nil {
		panic(err)
	}

	for _, d := range result.Items {
		// 最终格式化输出结果
		fmt.Printf("NAMESPACE:%v \t NAME:%v \t STATU:%+v\n", d.Namespace, d.Name, d.Status.Phase)
	}
}
