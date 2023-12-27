package fesgo

import "strings"

type treeNode struct {
	name     string
	children []*treeNode
}

func (t *treeNode) Put(path string) {
	root := t
	strs := strings.Split(path, "/")
	for index, name := range strs {
		if index == 0 {
			continue
		}
		children := root.children
		isMatch := false
		for _, node := range children {
			// 找到结点就结束，继续找下一个 name
			if node.name == name {
				isMatch = true
				children = node.children
				root = node
				break
			}
		}
		// 没找到节点就创建节点
		if !isMatch {
			node := &treeNode{
				name:     name,
				children: make([]*treeNode, 0),
			}
			root.children = append(root.children, node)
			root = node
		}
	}
}

func (t *treeNode) Get(path string) *treeNode {
	strs := strings.Split(path, "/")
	for index, name := range strs {
		if index == 0 {
			continue
		}
		children := t.children
		isMatch := false
		for _, node := range children {
			// 找到结点就结束，继续找下一个 name
			if node.name == name || node.name == "*" || strings.Contains(node.name, ":") {
				isMatch = true
				children = node.children
				t = node
				// 最尾部的节点
				if index == len(strs)-1 {
					return node
				}
				break
			}
		}
		if !isMatch {
			for _, node := range children {
				if node.name == "**" {
					return node
				}
			}
		}
	}
	return nil
}
