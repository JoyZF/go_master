package main

import (
	"container/list"
	"fmt"
)

type keyLru struct {
	limit    int                      //缓存数量
	evicts   *list.List               //双向链表用于淘汰数据
	elements map[string]*list.Element //记录缓存数据
	onEvict  func(string)
}

func NewKeyLru(limit int, onEvict func(key string)) *keyLru {
	return &keyLru{
		limit:    limit,
		evicts:   list.New(),
		elements: make(map[string]*list.Element),
		onEvict:  onEvict,
	}
}

func (k *keyLru) Add(key string) {
	if elem, ok := k.elements[key]; ok {
		k.evicts.MoveToFront(elem)
		return
	}
	// 添加新节点
	elem := k.evicts.PushFront(key)
	k.elements[key] = elem

	// 检查是否超出容量，如果超出就淘汰末尾节点数据
	if k.evicts.Len() > k.limit {
		k.removeOldest()
	}
}

func (klru *keyLru) removeOldest() {
	elem := klru.evicts.Back() //获取链表末尾节点
	if elem != nil {
		klru.removeElement(elem)
	}
}

//删除节点操作
func (klru *keyLru) removeElement(e *list.Element) {
	klru.evicts.Remove(e)
	key := e.Value.(string)
	delete(klru.elements, key)
	klru.onEvict(key)
}

func main() {
	lru := NewKeyLru(3, func(key string) {
		fmt.Println(fmt.Sprintf("key %s has removed", key))
	})
	lru.Add("1")
	lru.Add("2")
	lru.Add("3")
	lru.Add("1")
	lru.Add("4")
	lru.Add("2")
}
