package singleflight

import "sync"

type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	mu sync.Mutex // protects m
	m  map[string]*call
}

// 加锁的时间与访问数据源相比，可以忽略
// 针对相同的 key，无论 Do 被调用多少次，函数 fn 都只会被调用一次
// 第一次调用时已进入 Do 函数加了互斥锁，在没有释放锁之前，之后的调用都会阻塞
// 第一次调用时 m 是 nil，key 不会存在在 m 中. 当把 call 第一次放到 m 中时并且释放了锁，第二次调用的 Do 会
// 获取到锁，往下执行，走到 g.m[key] 中，此时 key 已经存在了，但是由于第一次执行了 wg.Add(1) 方法，所有第二次会
// 阻塞在 wg.Wait()，直到第一次执行 wg.Done() 才会取消阻塞.
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()

	if g.m == nil {
		g.m = make(map[string]*call)
	}

	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	// 这里可以理解成在并发情况下，要让等待锁的都执行完了，最后再执行这一步
	// 也就是说，第一个 Do 会做收尾工作，释放内存
	g.mu.Lock()
	delete(g.m, key) // update g.m
	g.mu.Unlock()

	return c.val, c.err
}
