package basetool

import "testing"

// 定义一个实现 INode 接口的结构体，用于测试
type testNode struct {
	id       int
	pid      int
	children []INode
}

// 实现 INode 接口的方法
func (n *testNode) GetId() int {
	return n.id
}

func (n *testNode) GetPid() int {
	return n.pid
}

func (n *testNode) IsRoot() bool {
	return n.pid == 0
}

func (n *testNode) SetChildren(children interface{}) {
	n.children = children.([]INode)
}

// TestGenerateTree tests the GenerateTree function
func TestGenerateTree(t *testing.T) {
	// 定义一个测试用的节点切片
	nodes := []INode{
		&testNode{id: 1, pid: 0},
		&testNode{id: 2, pid: 1},
		&testNode{id: 3, pid: 1},
		&testNode{id: 4, pid: 2},
		&testNode{id: 5, pid: 2},
		&testNode{id: 6, pid: 3},
		&testNode{id: 7, pid: 0},
		&testNode{id: 8, pid: 7},
		&testNode{id: 9, pid: 7},
		&testNode{id: 10, pid: 8},
	}

	// 调用 GenerateTree 函数，生成树结构
	trees := GenerateTree(nodes)

	// 检查生成的树是否符合预期
	if len(trees) != 2 {
		// 如果根节点的个数不是 2，报错
		t.Errorf("Expected 2 root nodes, but got %d", len(trees))
	}

	// 检查第一个根节点的子节点个数是否为 2
	if len(trees[0].(*testNode).children) != 2 {
		t.Errorf("Expected 2 children for node 1, but got %d", len(trees[0].(*testNode).children))
	}

	// 检查第二个根节点的子节点个数是否为 2
	if len(trees[1].(*testNode).children) != 2 {
		t.Errorf("Expected 2 children for node 7, but got %d", len(trees[1].(*testNode).children))
	}

	// 检查第一个根节点的第一个子节点的子节点个数是否为 2
	if len(trees[0].(*testNode).children[0].(*testNode).children) != 2 {
		t.Errorf("Expected 2 children for node 2, but got %d", len(trees[0].(*testNode).children[0].(*testNode).children))
	}

	// 检查第二个根节点的第一个子节点的子节点个数是否为 1
	if len(trees[1].(*testNode).children[0].(*testNode).children) != 1 {
		t.Errorf("Expected 1 child for node 8, but got %d", len(trees[1].(*testNode).children[0].(*testNode).children))
	}
}
