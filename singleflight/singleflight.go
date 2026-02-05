package singleflight

import (
	"sync"
)

// 代表正在进行或已结束的请求
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// Group manages all kinds of calls
type Group struct {
	m sync.Map // 使用sync.Map来优化并发性能
}

// Do 针对相同的key，保证多次调用Do()，都只会调用一次fn
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	// 1. 创建一个 call 实例
	c := &call{}
	c.wg.Add(1)

	// 2. 原子操作：查找或保存
	// actual: 实际存储在 map 里的值（可能是旧的，也可能是刚才存进去的新的）
	// loaded: 如果是 true，说明 map 里原本就有（我是搭便车的）
	//         如果是 false，说明 map 里原本没有（我是刚才存进去的，我是领头羊）
	actual, loaded := g.m.LoadOrStore(key, c)

	// 3. 情况一：我是搭便车的
	if loaded {
		c.wg.Done() // 我创建的 c 没用上，废弃掉，别忘了要把计数器减回去（或者一开始不Add，但这会增加逻辑复杂度） 。
		// 这里有个小浪费：每次都 new 了一个 call，但为了并发安全是值得的。

		existingCall := actual.(*call)
		existingCall.wg.Wait() // 等那个真正的领头羊
		return existingCall.val, existingCall.err
	}

	// 4. 情况二：我是领头羊 (loaded == false)
	// 此时 c 已经在 map 里了
	c.val, c.err = fn()
	c.wg.Done()

	g.m.Delete(key)
	return c.val, c.err
}
