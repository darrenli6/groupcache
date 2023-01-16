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
// lru用于实现LRU cache
package lru

import "container/list"

// 内存管理的一种页面置换算法
// lru 最近最少使用

// Cache is an LRU cache. It is not safe for concurrent access.
// cache结构用于实现LRU cache 算法,并发访问不安全
type Cache struct {
	// MaxEntries is the maximum number of cache entries before
	// an item is evicted. Zero means no limit.
	// 最大入口数 也就是缓冲最多存几条数据 超过了就会触发淘汰 0表示没有限制
	MaxEntries int

	// OnEvicted optionally specifies a callback function to be
	// executed when an entry is purged from the cache.
	// 销毁前的回调
	OnEvicted func(key Key, value interface{})

	// 链表
	ll *list.List
	// key为任意类型 值为指向链表一个节点的指针
	cache map[interface{}]*list.Element
}

// A Key may be any value that is comparable. See http://golang.org/ref/spec#Comparison_operators
// 任意类型
type Key interface{}

type entry struct {
	key   Key
	value interface{}
}

// New creates a new Cache.
// If maxEntries is zero, the cache has no limit and it's assumed
// that eviction is done by the caller.
// 初始化一个cache类型的实例
func New(maxEntries int) *Cache {
	return &Cache{
		MaxEntries: maxEntries,
		ll:         list.New(),
		cache:      make(map[interface{}]*list.Element),
	}
}

// Add adds a value to the cache.
// 往缓冲中增加一个值
func (c *Cache) Add(key Key, value interface{}) {
	// 如果cache还没有初始化，先初始化cache和 ll
	if c.cache == nil {
		c.cache = make(map[interface{}]*list.Element)
		c.ll = list.New()
	}
	// 如果key存在 记得前移动到头部 然后设置value
	if ee, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ee)
		ee.Value.(*entry).value = value
		return
	}
	// key不存在的时候 创建一条记录 插入到链表的头部 ele是这个Element的指针
	// 这里的Element是一个*entry类型 ele是*list.Element的类型
	ele := c.ll.PushFront(&entry{key, value})
	// cache这个map设置key为key类型的key value 为*list.Element的ele
	c.cache[key] = ele
	// 链表长度超过最大值入口值的时候触发清除操作
	if c.MaxEntries != 0 && c.ll.Len() > c.MaxEntries {
		c.RemoveOldest()
	}
}

// Get looks up a key's value from the cache.
// 根据key查找到value
func (c *Cache) Get(key Key) (value interface{}, ok bool) {
	if c.cache == nil {
		return
	}
	// 如果存在
	if ele, hit := c.cache[key]; hit {

		c.ll.MoveToFront(ele)
		// 将这个Element移动到链表头部
		// 返回entry的值
		return ele.Value.(*entry).value, true
	}
	return
}

// Remove removes the provided key from the cache.

func (c *Cache) Remove(key Key) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.removeElement(ele)
	}
}

// RemoveOldest removes the oldest item from the cache.
// 删除最久的
func (c *Cache) RemoveOldest() {
	if c.cache == nil {
		return
	}
	// 找到链表尾部的节点
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

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

// Clear purges all stored items from the cache.
func (c *Cache) Clear() {
	if c.OnEvicted != nil {
		for _, e := range c.cache {
			kv := e.Value.(*entry)
			c.OnEvicted(kv.key, kv.value)
		}
	}
	c.ll = nil
	c.cache = nil
}
