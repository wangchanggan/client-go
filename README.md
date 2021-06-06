# client-go源码分析

Source Code From https://github.com/kubernetes/client-go/releases/tag/kubernetes-1.21.0

参考Kubernetes源码分析（基于Kubernetes 1.14版本）（郑东旭/著）

## client-go源码结构
| 源码目录 | 说明 |
| :----: | :---- |
| applyconfigurations |  |
| discovery | 提供DiscoveryClient发现客户端 |
| dynamic | 提供DynamicClient动态客户端 |
| examples | 涵盖了client-go各种用例和功能的例子  |
| informers | 每种Kubernetes资源的Informer实现 |
| kubernetes | 提供ClientSet客户端 |
| kubernetes_test |  |
| listers | 为每一个Kubernetes资源提供Lister功能，该功能对Get和 List请求提供只读的缓存数据 |
| metadata |  |
| pkg |  |
| plugin | 提供OpenStack、 GCP和Azure等云服务商授权插件 |
| rest | 提供RESTClient客户端，对Kubernetes API Server执行RESTful操作|
| restmapper | 基于restmapper的客户端 |
| scale | 提供ScaleClient客户端，用于扩容或缩容Deployment、ReplicaSet、Replication Controller等资源对象 |
| testing |  |
| third_party |  |
| tools | 提供常用工具，例如SharedInformer、Reflector、DealtFIFO及Indexers。提供 Client查询和缓存机制，以减少向kube-apiserver发起的请求数等 |
| transport | 提供安全的TCP连接，支持 Http Stream，某些操作需要在客户端和容器之间传输二进制流，例如exec、attach等操作。该功能由内部的spdy包提供支持 |
| util | 提供常用方法，例如 WorkQueue工作队列、Certificate证书管理等 |



## kubeconfig配置管理
client-go读取kubeconfig配置信息并生成config对象，用于与kube-apiserver通信vendor/k8s.io/client-go/tools/clientcmd/client_config.go:616

1.加载kubeconfig配置信息
vendor/k8s.io/client-go/tools/clientcmd/loader.go:174

vendor/k8s.io/client-go/tools/clientcmd/loader.go:406

2.合并多个kubeconfig配置信息
vendor/k8s.io/client-go/tools/clientcmd/loader.go:246