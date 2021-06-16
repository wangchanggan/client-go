/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package workqueue

import (
	"math"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type RateLimiter interface {
	// When gets an item and gets to decide how long that item should wait
	// 获取指定元素应该等待的时间。
	When(item interface{}) time.Duration
	// Forget indicates that an item is finished being retried.  Doesn't matter whether its for perm failing
	// or for success, we'll stop tracking it
	// 释放指定元素，清空该元素的排队数。
	Forget(item interface{})
	// NumRequeues returns back how many failures the item has had
	// 获取指定元素的排队数。
	NumRequeues(item interface{}) int

	// 注意：这里有一个非常重要的概念——限速周期，一个限速周期是指从执行AddRateLimited方法到执行完Forget方法之间的时间。
	// 如果该元素被Forget方法处理完，则清空排队数。
}

// DefaultControllerRateLimiter is a no-arg constructor for a default rate limiter for a workqueue.  It has
// both overall and per-item rate limiting.  The overall is a token bucket and the per-item is exponential
func DefaultControllerRateLimiter() RateLimiter {
	return NewMaxOfRateLimiter(
		NewItemExponentialFailureRateLimiter(5*time.Millisecond, 1000*time.Second),
		// 10 qps, 100 bucket size.  This is only for retry speed and its only the overall factor (not per item)
		// 在实例化rate.NewLimiter后，传入r和b两个参数，其中r参数表示每秒往“桶”里填充的token 数量，b参数表示令牌桶的大小(即令牌桶最多存放的token数量)
		//我们假定r为10, b为100。假设在一个限速周期内插入了1000 个元素，通过r.Limiter.Reserve().Delay函数返回指定元素应该等待的时间
		//那么前b (即100)个元素会被立刻处理，而后面元素的延迟时间分别为item100/100ms、 iteml01/200ms、item102/300ms、item103/400ms, 以此类推。
		&BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(10), 100)},
	)
}

// BucketRateLimiter adapts a standard bucket to the workqueue ratelimiter API
type BucketRateLimiter struct {
	*rate.Limiter
}

var _ RateLimiter = &BucketRateLimiter{}

func (r *BucketRateLimiter) When(item interface{}) time.Duration {
	return r.Limiter.Reserve().Delay()
}

func (r *BucketRateLimiter) NumRequeues(item interface{}) int {
	return 0
}

func (r *BucketRateLimiter) Forget(item interface{}) {
}

// ItemExponentialFailureRateLimiter does a simple baseDelay*2^<num-failures> limit
// dealing with max failures and expiration are up to the caller
type ItemExponentialFailureRateLimiter struct {
	failuresLock sync.Mutex
	// 用于统计元素排队数，每当AddRateLimited方法插入新元素时，会为该字段加1.
	failures     map[interface{}]int
	// baseDelay 字段是最初的限速单位(默认为5ms)
	baseDelay time.Duration
	// maxDelay 字段是最大限速单位(默认为1000s)。
	maxDelay  time.Duration

	// 注意：在同一限速周期内，如果不存在相同元素，那么所有元素的延迟时间为baseDelay;
	// 而在同一限速周期内，如果存在相同元素，那么相同元素的延迟时间呈指数级增长，最长延迟时间不超过maxDelay。
}

var _ RateLimiter = &ItemExponentialFailureRateLimiter{}

func NewItemExponentialFailureRateLimiter(baseDelay time.Duration, maxDelay time.Duration) RateLimiter {
	return &ItemExponentialFailureRateLimiter{
		failures:  map[interface{}]int{},
		baseDelay: baseDelay,
		maxDelay:  maxDelay,
	}
}

func DefaultItemBasedRateLimiter() RateLimiter {
	return NewItemExponentialFailureRateLimiter(time.Millisecond, 1000*time.Second)
}

func (r *ItemExponentialFailureRateLimiter) When(item interface{}) time.Duration {
	r.failuresLock.Lock()
	defer r.failuresLock.Unlock()

	/*假定baseDelay是1 * time.Millisecond，maxDelay 是1000 * time.Second。
	假设在一个限速周期内通过AddRateLimited方法插入10个相同元素，
	那么第1个元素会通过延迟队列的AddAfter方法插入并设置延迟时间为1ms (即baseDelay)，
	第2个相同元素的延迟时间为2ms，第3个相同元素的延迟时间为4ms，第4个相同元素的延迟时间为8ms，第5个相同元素的延迟时间为16ms.....
	第10个相同元素的延迟时间为512ms，最长延迟时间不超过1000s (即maxDelay)。*/

	exp := r.failures[item]
	r.failures[item] = r.failures[item] + 1

	// The backoff is capped such that 'calculated' value never overflows.
	backoff := float64(r.baseDelay.Nanoseconds()) * math.Pow(2, float64(exp))
	if backoff > math.MaxInt64 {
		return r.maxDelay
	}

	calculated := time.Duration(backoff)
	if calculated > r.maxDelay {
		return r.maxDelay
	}

	return calculated
}

func (r *ItemExponentialFailureRateLimiter) NumRequeues(item interface{}) int {
	r.failuresLock.Lock()
	defer r.failuresLock.Unlock()

	return r.failures[item]
}

func (r *ItemExponentialFailureRateLimiter) Forget(item interface{}) {
	r.failuresLock.Lock()
	defer r.failuresLock.Unlock()

	delete(r.failures, item)
}

// ItemFastSlowRateLimiter does a quick retry for a certain number of attempts, then a slow retry after that
type ItemFastSlowRateLimiter struct {
	failuresLock sync.Mutex
	//masteptsp字段。计数器算法核心实现的
	failures     map[interface{}]int
    // 用于控制从fast速率转换到slow速率
	maxFastAttempts int
	// 用于定义fast速率
	fastDelay       time.Duration
	// 用于定义slow速率
	slowDelay       time.Duration
}

var _ RateLimiter = &ItemFastSlowRateLimiter{}

func NewItemFastSlowRateLimiter(fastDelay, slowDelay time.Duration, maxFastAttempts int) RateLimiter {
	return &ItemFastSlowRateLimiter{
		failures:        map[interface{}]int{},
		fastDelay:       fastDelay,
		slowDelay:       slowDelay,
		maxFastAttempts: maxFastAttempts,
	}
}

func (r *ItemFastSlowRateLimiter) When(item interface{}) time.Duration {
	r.failuresLock.Lock()
	defer r.failuresLock.Unlock()

	/*假设fastDelay 是5 * time.Millisecond，slowDelay 是10 * time.Second，maxFastAttempts是3。
	在一个限速周期内通过AddRateLimited方法插入4个相同的元素，那么前3个元素使用fastDelay定义的fast速率，
	当触发maxFastAttempts字段时，第4个元素使用slowDelay定义的slow速率。*/
	r.failures[item] = r.failures[item] + 1

	if r.failures[item] <= r.maxFastAttempts {
		return r.fastDelay
	}

	return r.slowDelay
}

func (r *ItemFastSlowRateLimiter) NumRequeues(item interface{}) int {
	r.failuresLock.Lock()
	defer r.failuresLock.Unlock()

	return r.failures[item]
}

func (r *ItemFastSlowRateLimiter) Forget(item interface{}) {
	r.failuresLock.Lock()
	defer r.failuresLock.Unlock()

	delete(r.failures, item)
}

// MaxOfRateLimiter calls every RateLimiter and returns the worst case response
// When used with a token bucket limiter, the burst could be apparently exceeded in cases where particular items
// were separately delayed a longer time.
type MaxOfRateLimiter struct {
	limiters []RateLimiter
}

func (r *MaxOfRateLimiter) When(item interface{}) time.Duration {
	ret := time.Duration(0)
	for _, limiter := range r.limiters {
		curr := limiter.When(item)
		if curr > ret {
			ret = curr
		}
	}

	return ret
}

func NewMaxOfRateLimiter(limiters ...RateLimiter) RateLimiter {
	return &MaxOfRateLimiter{limiters: limiters}
}

func (r *MaxOfRateLimiter) NumRequeues(item interface{}) int {
	ret := 0
	for _, limiter := range r.limiters {
		curr := limiter.NumRequeues(item)
		if curr > ret {
			ret = curr
		}
	}

	return ret
}

func (r *MaxOfRateLimiter) Forget(item interface{}) {
	for _, limiter := range r.limiters {
		limiter.Forget(item)
	}
}
