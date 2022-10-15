package consistenthash

import (
	"hash/crc32"
	"sort"
)

// HashFunc 定义用于生成hash的函数
type HashFunc func(data []byte) uint32

// NodeMap 存储节点，实现从NodeMap中选节点
type NodeMap struct {
	hashFunc    HashFunc
	nodeHashs   []int // sorted
	nodehashMap map[int]string
}

// NewNodeMap 创建新的NodeMap
func NewNodeMap(fn HashFunc) *NodeMap {
	m := &NodeMap{
		hashFunc:    fn,
		nodehashMap: make(map[int]string),
	}
	if m.hashFunc == nil {
		m.hashFunc = crc32.ChecksumIEEE
	}
	return m
}

// IsEmpty NodeMap是否为空
func (m *NodeMap) IsEmpty() bool {
	return len(m.nodeHashs) == 0
}

// AddNode 将给定的节点添加到一致的Hash
func (m *NodeMap) AddNode(keys ...string) {
	for _, key := range keys {
		if key == "" {
			continue
		}
		hash := int(m.hashFunc([]byte(key)))
		m.nodeHashs = append(m.nodeHashs, hash)
		m.nodehashMap[hash] = key
	}
	sort.Ints(m.nodeHashs)
}

// PickNode 获取hash中与提供的键最接近的项。
func (m *NodeMap) PickNode(key string) string {
	if m.IsEmpty() {
		return ""
	}

	hash := int(m.hashFunc([]byte(key)))

	// 二进制搜索以查找适当的副本
	idx := sort.Search(len(m.nodeHashs), func(i int) bool {
		return m.nodeHashs[i] >= hash
	})
	if idx == len(m.nodeHashs) {
		idx = 0
	}

	return m.nodehashMap[m.nodeHashs[idx]]
}
