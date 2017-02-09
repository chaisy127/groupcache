/*
Copyright 2013 Google Inc.

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

// Package lru implements an LRU cache.
package lru

import "container/list"

// Cache is an LRU cache. It is not safe for concurrent access.
// Cache结构不是线程安全的
type Cache struct {
	// MaxEntries is the maximum number of cache entries before
	// an item is evicted. Zero means no limit.
	MaxEntries int

	// OnEvicted optionally specificies a callback function to be
	// executed when an entry is purged from the cache.
	// 在删除元素的时候回调用该函数，主要用于清理内存
	OnEvicted func(key Key, value interface{})

	// golang官方库，本质上是一个双向链表
	// 该链表仅为了对已缓存的元素进行排序
	ll *list.List
	// 该map结构用于通过key值快速的获取到value
	// map结构中的value为指针类型，与ll变量中的每个元素共用同一块内存
	// list.Element类型见https://golang.org/pkg/container/list/#Element
	cache map[interface{}]*list.Element
}

// A Key may be any value that is comparable. See http://golang.org/ref/spec#Comparison_operators
type Key interface{}

// 该类型的数据会存储到ll变量和cache变量的*list.Element中
type entry struct {
	key   Key
	value interface{}
}

// New creates a new Cache.
// If maxEntries is zero, the cache has no limit and it's assumed
// that eviction is done by the caller.
// 创建Cache结构体指针，该函数在groupcache项目中未调用
// groupcache项目中创建Cache结构体指针是通过下面的Add函数自动创建的
// groupcache项目中的所有Cache结构体指针中的MaxEntries属性的值默认都为0
// 如需要执行MaxEntries的值，需要开发者手动调用该函数，并指定MaxEntries的值
func New(maxEntries int) *Cache {
	return &Cache{
		MaxEntries: maxEntries,
		ll:         list.New(),
		cache:      make(map[interface{}]*list.Element),
	}
}

// Add adds a value to the cache.
// 如果key对应的value未缓存，则将key和value设置到cache成员变量中，并将value添加到ll链表的最前面
// 如果key对应的value已缓存，则分别更新其在ll和cache成员变量中的值，并将value移动到ll链表的最前面
// 将value移动到ll链表最前面的原因详见lru算法
// 添加完value值之后，判断当前元素个数是否已经达到MaxEntries指定的最大限制，
// 如果c.ll.Len() > c.MaxEntries，则删除ll链表中最后一个元素的值，同时将该值从cache成员变量中删除
// 由于MaxEntries变量在groupcache项目中的默认值为0，所以在Add函数中不会主动调用RemoveOldest函数
func (c *Cache) Add(key Key, value interface{}) {
	if c.cache == nil {
		c.cache = make(map[interface{}]*list.Element)
		c.ll = list.New()
	}
	if ee, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ee)
		ee.Value.(*entry).value = value
		return
	}
	ele := c.ll.PushFront(&entry{key, value})
	c.cache[key] = ele
	if c.MaxEntries != 0 && c.ll.Len() > c.MaxEntries {
		c.RemoveOldest()
	}
}

// Get looks up a key's value from the cache.
// 如果cache成员变量为nil或缓存为命中，则返回的value为nil，ok为false，具体原因可查看golang变量默认值相关说明
// 如果缓存命中，则返回命中的value，并将value移动到ll链表的最前面
func (c *Cache) Get(key Key) (value interface{}, ok bool) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.ll.MoveToFront(ele)
		return ele.Value.(*entry).value, true
	}
	return
}

// Remove removes the provided key from the cache.
// 从ll和cache成员变量中删除key对应的value
// 该函数在groupcache项目中未调用
func (c *Cache) Remove(key Key) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.removeElement(ele)
	}
}

// RemoveOldest removes the oldest item from the cache.
func (c *Cache) RemoveOldest() {
	if c.cache == nil {
		return
	}
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

// 从成员变量ll和cache中删除指定element，并执行OnEvicted清理函数
func (c *Cache) removeElement(e *list.Element) {
	c.ll.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}

// Len returns the number of items in the cache.
func (c *Cache) Len() int {
	if c.cache == nil {
		return 0
	}
	return c.ll.Len()
}
