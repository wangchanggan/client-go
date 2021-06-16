package main

import (
	"fmt"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"strings"
)

// 定义一个索引器函数UsersIndexFunc，在该函数中，定义查询出所有Pod资源下Annotations字段的key为users的Pod
func UsersIndexFunc(obj interface{}) ([]string, error) {
	pod := obj.(*v1.Pod)
	usersString := pod.Annotations[ "users"]
	return strings.Split(usersString, ","), nil
}

func main() {
	// cache.NewIndexer函数实例化了Indexer 对象，该函数接收两个参数:
	// 第1个参数是KeyFunc,它用于计算资源对象的key，计算默认使用cache.MetaNamespaceKeyFunc函数
	// 第2个参数是cache.Indexers，用于定义索引器，其中key为索引器的名称(即byUser)，value 为索引器。
	index := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{"byUser": UsersIndexFunc})
	pod1 := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "one", Annotations: map[string]string{"users": "ernie,bert"}}}
	pod2 := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "two", Annotations: map[string]string{"users": "bert, oscar"}}}
	pod3 := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "tre", Annotations: map[string]string{"users": "ernie, elmo "}}}

	// 通过index.Add函数添加3个Pod 资源对象。
	index.Add(pod1)
	index.Add(pod2)
	index.Add(pod3)

	// 通过index.ByIndex函数查询byUser索引器下匹配ernie字段的Pod列表。
	erniePods, err := index.ByIndex("byUser", "ernie")
	if err != nil {
		panic(err)
	}

	for _, erniePod := range erniePods {
		fmt.Println(erniePod.(*v1.Pod).Name)
	}
}
