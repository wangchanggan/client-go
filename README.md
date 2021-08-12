# client-go源码分析

Source Code From https://github.com/kubernetes/client-go/releases/tag/kubernetes-1.21.0

参考Kubernetes源码分析（基于Kubernetes 1.14版本）（郑东旭/著）

结合client-go编程式交互.doc阅读

## 目录
-   [client-go源码分析](#client-go源码分析)
    -   [client-go源码结构](#client-go源码结构)
    -   [Client客户端对象](#client客户端对象)
        -   [kubeconfig配置管理](#kubeconfig配置管理)
        -   [RESTClient客户端](#restclient客户端)
        -   [ClientSet客户端](#clientset客户端)
        -   [DynamicClient客户端](#dynamicclient客户端)
        -   [DiscoveryClient客户端](#discoveryclient客户端)
            -   [获取Kubernetes API
                Server所支持的资源组、资源版本、资源信息](#获取kubernetes-api-server所支持的资源组资源版本资源信息)
            -   [本地缓存的DiscoveryClient](#本地缓存的discoveryclient)
    -   [Informer机制](#informer机制)
        -   [Informer机制架构设计](#informer机制架构设计)
        -   [Reflector](#reflector)
        -   [DeltaFIFO](#deltafifo)
        -   [Indexer](#indexer)
    -   [WorkQueue](#workqueue)
        -   [FIFO队列](#fifo队列)
        -   [延迟队列](#延迟队列)
        -   [限速队列](#限速队列)
            -   [令牌桶算法](#令牌桶算法)
            -   [排队指数算法](#排队指数算法)
            -   [计数器算法](#计数器算法)
            -   [混合模式](#混合模式)
    -   [EventBroadcaster事件管理器](#eventbroadcaster事件管理器)
        -   [EventRecorder](#eventrecorder)
        -   [EventBroadcaster](#eventbroadcaster)
        -   [broadcasterWatcher](#broadcasterwatcher)
    -   [代码生成器（需结合kubernetes代码）](#代码生成器需结合kubernetes代码)
        -   [client-gen代码生成器](#client-gen代码生成器)
        -   [lister-gen代码生成器](#lister-gen代码生成器)
        -   [informer-gen代码生成器](#informer-gen代码生成器)

## client-go源码结构
| 源码目录 | 说明 |  备注 |
| :---: | :---- | :---- |
| applyconfigurations |  |  |
| discovery | 提供DiscoveryClient发现客户端 |  |
| dynamic | 提供DynamicClient动态客户端 |  |
| examples | 涵盖了client-go各种用例和功能的例子  | 包括源码解析过程中添加的示例 |
| informers | 每种Kubernetes资源的Informer实现 |  |
| kubernetes | 提供ClientSet客户端 |  |
| kubernetes_test |  |  |
| listers | 为每一个Kubernetes资源提供Lister功能，该功能对Get和 List请求提供只读的缓存数据 |  |
| metadata |  |  |
| pkg |  |  |
| plugin | 提供OpenStack、 GCP和Azure等云服务商授权插件 |  |
| rest | 提供RESTClient客户端，对Kubernetes API Server执行RESTful操作|  |
| restmapper | 基于restmapper的客户端 |  |
| scale | 提供ScaleClient客户端，用于扩容或缩容Deployment、ReplicaSet、Replication Controller等资源对象 |  |
| testing |  |  |
| third_party |  |  |
| tools | 提供常用工具，例如SharedInformer、Reflector、DealtFIFO及Indexers。提供 Client查询和缓存机制，以减少向kube-apiserver发起的请求数等 |  |
| transport | 提供安全的TCP连接，支持 Http Stream，某些操作需要在客户端和容器之间传输二进制流，例如exec、attach等操作。该功能由内部的spdy包提供支持 |  |
| util | 提供常用方法，例如 WorkQueue工作队列、Certificate证书管理等 |  |
| vendor | 存放项目依赖的库代码 | 源码中无此目录，为了方便后期编译运行添加，通过go mod也可以 |



## Client客户端对象
### kubeconfig配置管理
client-go读取kubeconfig配置信息并生成config对象，用于与kube-apiserver通信tools/clientcmd/client_config.go:616

1.加载kubeconfig配置信息
tools/clientcmd/loader.go:174

tools/clientcmd/loader.go:406

2.合并多个kubeconfig配置信息
tools/clientcmd/loader.go:246


### RESTClient客户端
见examples/REST_client.go

RESTClient发送请求的过程对Go语言标准库nethttp 进行了封装，由Do-request函数实现rest/request.go:978


### ClientSet客户端
见examples/client_set.go

podCient.List函数通过RESTClient获得Pod列表kubernetes/typed/core/v1/pod.go:89


### DynamicClient客户端
见examples/dynamic_client.go


### DiscoveryClient客户端
见examples/discovery_client.go

#### 获取Kubernetes API Server所支持的资源组、资源版本、资源信息
discovery/discovery_client.go:156

#### 本地缓存的DiscoveryClient
discovery/cached/disk/cached_discovery.go:64



## Informer机制
### Informer机制架构设计
1.资源Informer

每一个Kubernetes资源上都实现了Informer机制。每一个Informer上都会实现Informer和Lister方法，例如PodInformer,代码示例informers/core/v1/pod.go:37

2.Shared Informer共享机制

informers/factory.go:66,168,132


### Reflector
1.获取资源列表数据tools/cache/reflector.go:254

2.监控资源对象tools/cache/reflector.go:429,444


### DeltaFIFO
tools/cache/delta_fifo.go:95

1.生产者方法tools/cache/delta_fifo.go:415

2.消费者方法

tools/cache/delta_fifo.go:528

tools/cache/shared_informer.go:528

3.Resync机制

tools/cache/delta_fifo.go:669

tools/cache/delta_fifo.go:686


### Indexer
1.ThreadSafeMap并发安全存储tools/cache/thread_safe_store.go:65

2.Indexer索引器

examples/indexer.go

tools/cache/index.go

3.Indexer索引器核心实现tools/cache/thread_safe_store.go:186



## WorkQueue
### FIFO队列
util/workqueue/queue.go:26,77


### 延迟队列
util/workqueue/delaying_queue.go:30,75


### 限速队列
util/workqueue/default_rate_limiters.go:27

#### 令牌桶算法
util/workqueue/default_rate_limiters.go:52

#### 排队指数算法
util/workqueue/default_rate_limiters.go:76,103

#### 计数器算法
util/workqueue/default_rate_limiters.go:145,168

#### 混合模式
util/workqueue/default_rate_limiters.go:45



## EventBroadcaster事件管理器
vendor/k8s.io/api/core/v1/types.go:5463


### EventRecorder
tools/record/event.go:88

以Event方法为例，记录当前发生的事件，Event→recorder.generateEvent→recorder.ActionOrDrop 

vendor/k8s.io/apimachinery/pkg/watch/mux.go:220


### EventBroadcaster
tools/record/event.go:159

vendor/k8s.io/apimachinery/pkg/watch/mux.go:260


### broadcasterWatcher
tools/record/event.go:113,297



## 代码生成器（需结合kubernetes代码）
| 代码生成器 | 说明 |
| :---: | :---- |
| client-gen | 一种为资源生成 ClientSet客户端的工具 |
| lister-gen | 一种为资源生成Lister的工具(即get和 list方法) |
| informer-gen | 一种为资源生成 Informer的工具 |


### client-gen代码生成器
生成规则（以Pod资源对象为例）
$GOPATH/src/k8s.io/kubernetes/vendor/k8s.io/api/core/v1/types.go:3688

$GOPATH/src/k8s.io/kubernetes/vendor/k8s.io/code-generator/cmd/client-gen/generators/generator_for_type.go:80


### lister-gen代码生成器
$GOPATH/src/k8s.io/kubernetes/vendor/k8s.io/code-generator/cmd/lister-gen/generators/lister.go:64


### informer-gen代码生成器
$GOPATH/src/k8s.io/kubernetes/vendor/k8s.io/code-generator/cmd/informer-gen/generators/packages.go:94