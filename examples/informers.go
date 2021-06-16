package main

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"time"
)

func main() {
	config, err := clientcmd.BuildConfigFromFlags("", "/root/.kube/config")
	if err != nil {
		panic(err)
	}

	// 通过kubernetes.NewForConfig 创建clientset对象，Informer 需要通过ClientSet与Kubernetes API Server进行交互。
	c1ientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// 创建stopCh对象，该对象用于在程序进程退出之前通知Informer提前退出，因为Informer是一个持久运行的goroutine。
	stopCh := make(chan struct{})
	defer close(stopCh)

	//informers.NewSharedInformerFactory函数实例化了SharedInformer 对象，它接收两个参数:
	// 第1个参数clientset是用于与Kubernetes API Server交互的客户端
	// 第2个参数time.Minute用于设置多久进行一次resync (重新同步)，resync 会周期性地执行List操作，将所有的资源存放在InformerStore中
	// 如果该参数为0,则禁用resync功能。
	sharedInformers := informers.NewSharedInformerFactory(c1ientset, time.Minute)

	// 通过sharedInformers.Core().V1().Pods().Informer()Informer可以得到具体Pod 资源的informer对象。
	informer := sharedInformers.Core().V1().Pods().Informer()

	//通过informer.AddEventHandler函数可以为Pod资源添加资源事件回调方法，支持3种资源事件回调方法：
	//  在正常的情况下，kubernetes的其他组件在使用Informer机制时触发资源事件回调方法，将资源对象推送到WorkQueue或其他队列中
	// 示例中，直接输出触发的资源事件
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		// AddFunc：当创建Pod资源对象时触发的事件回调方法。
		AddFunc: func(obj interface{}) {
			mObj := obj.(v1.Object)
			log.Printf("New Pod Added to Store: %s", mObj.GetName())
		},
		// UpdateFunc：当更新Pod资源对象时触发的事件回调方法。
		UpdateFunc: func(old0bj, newObj interface{}) {
			oObj := old0bj.(v1.Object)
			nObj := newObj.(v1.Object)
			log.Printf("%s Pod Updated to %s", oObj.GetName(), nObj.GetName())
		},
		// DeleteFunc：当删除Pod资源对象时触发的事件回调方法。
		DeleteFunc: func(obj interface{}) {
			mObj := obj.(v1.Object)
			log.Printf("Pod Deleted from Store: %s", mObj.GetName())
		},
	})

	// 通过informer.Run 函数运行当前的Informer，内部为Pod资源类型创建Informer。
	informer.Run(stopCh)
}
