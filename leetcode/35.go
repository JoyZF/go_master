package main

type Node struct {
	Val    int
	Next   *Node
	Random *Node
}

// 存储节点， 递归时遇到就返回
var cachedNode map[*Node]*Node

func deepCopy(node *Node) *Node {
	// 递归终止条件
	if node == nil {
		return nil
	}
	if n, has := cachedNode[node]; has {
		return n
	}
	newNode := &Node{Val: node.Val}
	cachedNode[node] = newNode
	newNode.Next = deepCopy(node.Next)
	newNode.Random = deepCopy(node.Random)
	return newNode
}

func copyRandomList(head *Node) *Node {
	cachedNode = map[*Node]*Node{}
	return deepCopy(head)
}
