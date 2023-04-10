package main

import "fmt"

type Node struct {
	Key   int
	Value int
	pre   *Node
	next  *Node
}

type LRUCache struct {
	limit   int
	HashMap map[int]*Node
	head    *Node
	end     *Node
}

func Constructor(capacity int) LRUCache {
	lruCache := LRUCache{limit: capacity}
	lruCache.HashMap = make(map[int]*Node, capacity)
	return lruCache
}

func (l *LRUCache) Get(key int) int {
	if v, ok := l.HashMap[key]; ok {
		l.refreshNode(v)
		return v.Value
	} else {
		return -1
	}
}

func (l *LRUCache) Put(key int, value int) {
	if v, ok := l.HashMap[key]; !ok {
		if len(l.HashMap) >= l.limit {
			oldKey := l.removeNode(l.head)
			delete(l.HashMap, oldKey)
		}
		node := Node{Key: key, Value: value}
		l.addNode(&node)
		l.HashMap[key] = &node
	} else {
		v.Value = value
		l.refreshNode(v)
	}
}

func (l *LRUCache) refreshNode(node *Node) {
	if node == l.end {
		return
	}
	l.removeNode(node)
	l.addNode(node)
}

func (l *LRUCache) removeNode(node *Node) int {
	if node == l.end {
		l.end = l.end.pre
		l.end.next = nil
	} else if node == l.head {
		l.head = l.head.next
		l.head.pre = nil
	} else {
		node.pre.next = node.next
		node.next.pre = node.pre
	}
	return node.Key
}

func (l *LRUCache) addNode(node *Node) {
	if l.end != nil {
		l.end.next = node
		node.pre = l.end
		node.next = nil
	}
	l.end = node
	if l.head == nil {
		l.head = node
	}
}

func main() {
	lruCache := Constructor(3)
	lruCache.Put(1, 1)
	lruCache.Put(2, 2)
	lruCache.Put(3, 3)
	fmt.Println(lruCache.Get(1))
	lruCache.Put(4, 4)

	fmt.Println(lruCache.Get(1))
	fmt.Println(lruCache.Get(2))
}
