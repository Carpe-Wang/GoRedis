package sortedset

import (
	"math/rand"
	"time"
)

const (
	maxLevel = 16
)

// Element is a key-score pair
type Element struct {
	Member string
	Score  float64
}

// Level is a forward list with span
type Level struct {
	forward *node
	span    int64
}

// node is the data structure for a skip list node
type node struct {
	Element
	backward *node
	level    []*Level
}

// SortedSet is a sorted set implemented by skip list
type SortedSet struct {
	header *node
	tail   *node
	length int64
	level  int
}

// Make creates a new sorted set
func Make() *SortedSet {
	// init random seed
	rand.Seed(time.Now().UnixNano())

	sortedSet := &SortedSet{
		level: 1,
	}

	// init header node
	sortedSet.header = &node{
		level: make([]*Level, maxLevel),
	}
	for i := 0; i < maxLevel; i++ {
		sortedSet.header.level[i] = &Level{}
	}

	return sortedSet
}

// RandomLevel returns a random level for the new node
func randomLevel() int {
	level := 1
	for float32(rand.Int31()&0xFFFF) < (0.25 * 0xFFFF) {
		level++
		if level >= maxLevel {
			return maxLevel
		}
	}
	return level
}

// insert inserts a new node to the skip list
func (sortedSet *SortedSet) insert(member string, score float64) *node {
	update := make([]*node, maxLevel)
	rank := make([]int64, maxLevel)

	// find position to insert
	currNode := sortedSet.header
	for i := sortedSet.level - 1; i >= 0; i-- {
		if i == sortedSet.level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}

		for currNode.level[i].forward != nil &&
			(currNode.level[i].forward.Score < score ||
				(currNode.level[i].forward.Score == score && // score is the same, compare member
					currNode.level[i].forward.Member < member)) {
			rank[i] += currNode.level[i].span
			currNode = currNode.level[i].forward
		}
		update[i] = currNode
	}

	level := randomLevel()
	// extend sorted set level
	if level > sortedSet.level {
		for i := sortedSet.level; i < level; i++ {
			rank[i] = 0
			update[i] = sortedSet.header
			update[i].level[i].span = sortedSet.length
		}
		sortedSet.level = level
	}

	// create node
	newNode := &node{
		Element: Element{
			Member: member,
			Score:  score,
		},
		level: make([]*Level, level),
	}

	for i := 0; i < level; i++ {
		newNode.level[i] = &Level{}
		newNode.level[i].forward = update[i].level[i].forward
		update[i].level[i].forward = newNode

		// update span covered by update[i] as newNode is inserted here
		newNode.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		update[i].level[i].span = (rank[0] - rank[i]) + 1
	}

	// increment span for untouched levels
	for i := level; i < sortedSet.level; i++ {
		update[i].level[i].span++
	}

	// update backward node
	if update[0] == sortedSet.header {
		newNode.backward = nil
	} else {
		newNode.backward = update[0]
	}

	if newNode.level[0].forward != nil {
		newNode.level[0].forward.backward = newNode
	} else {
		sortedSet.tail = newNode
	}

	sortedSet.length++
	return newNode
}

// deleteNode removes a node from sorted set
func (sortedSet *SortedSet) deleteNode(node *node, update []*node) {
	for i := 0; i < sortedSet.level; i++ {
		if update[i].level[i].forward == node {
			update[i].level[i].span += node.level[i].span - 1
			update[i].level[i].forward = node.level[i].forward
		} else {
			update[i].level[i].span--
		}
	}

	if node.level[0].forward != nil {
		node.level[0].forward.backward = node.backward
	} else {
		sortedSet.tail = node.backward
	}

	// update the level of sorted set
	for sortedSet.level > 1 && sortedSet.header.level[sortedSet.level-1].forward == nil {
		sortedSet.level--
	}

	sortedSet.length--
}

// Remove deletes a member from the sorted set
func (sortedSet *SortedSet) Remove(member string) bool {
	update := make([]*node, maxLevel)
	node := sortedSet.header

	for i := sortedSet.level - 1; i >= 0; i-- {
		for node.level[i].forward != nil &&
			(node.level[i].forward.Member < member) {
			node = node.level[i].forward
		}
		update[i] = node
	}

	node = node.level[0].forward
	if node != nil && node.Member == member {
		sortedSet.deleteNode(node, update)
		return true
	}

	return false
}

// Exists checks if a member exists in the sorted set
func (sortedSet *SortedSet) Exists(member string) bool {
	node := sortedSet.getByMember(member)
	return node != nil
}

// Add adds or updates a member in the sorted set
func (sortedSet *SortedSet) Add(member string, score float64) bool {
	// find if member exists
	existed := false
	update := make([]*node, maxLevel)
	rank := make([]int64, maxLevel)

	node := sortedSet.header
	for i := sortedSet.level - 1; i >= 0; i-- {
		if i == sortedSet.level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}

		for node.level[i].forward != nil &&
			(node.level[i].forward.Score < score ||
				(node.level[i].forward.Score == score &&
					node.level[i].forward.Member < member)) {
			rank[i] += node.level[i].span
			node = node.level[i].forward
		}
		update[i] = node
	}

	/* If the node is already in the skip list, remove it and re-insert it. */
	node = node.level[0].forward
	if node != nil && node.Member == member {
		sortedSet.deleteNode(node, update)
		existed = true
	}

	sortedSet.insert(member, score)
	return existed
}

// GetRank returns the rank of a member
func (sortedSet *SortedSet) GetRank(member string, reverse bool) (int64, bool) {
	var rank int64 = 0
	node := sortedSet.header

	for i := sortedSet.level - 1; i >= 0; i-- {
		for node.level[i].forward != nil &&
			(node.level[i].forward.Member < member) {
			rank += node.level[i].span
			node = node.level[i].forward
		}
	}

	node = node.level[0].forward
	if node != nil && node.Member == member {
		if reverse {
			return sortedSet.length - rank - 1, true
		}
		return rank, true
	}

	return 0, false
}

// GetScore returns the score of a member
func (sortedSet *SortedSet) GetScore(member string) (float64, bool) {
	node := sortedSet.getByMember(member)
	if node != nil {
		return node.Score, true
	}
	return 0, false
}

// GetByRank returns a member at the given rank
func (sortedSet *SortedSet) GetByRank(rank int64, reverse bool) (*Element, bool) {
	// handle reverse
	if reverse {
		rank = sortedSet.length - rank - 1
	}

	// validate rank
	if rank < 0 || rank >= sortedSet.length {
		return nil, false
	}

	var i int64 = 0
	n := sortedSet.header

	// scan forward from header
	for i = 0; i < rank; {
		if n.level[0].forward == nil {
			// should not happen
			return nil, false
		}

		i += n.level[0].span
		n = n.level[0].forward
	}

	return &Element{
		Member: n.Member,
		Score:  n.Score,
	}, true
}

// GetByScoreRange returns members with score in the given range
func (sortedSet *SortedSet) GetByScoreRange(min, max float64, offset, limit int64, reverse bool) []*Element {
	if reverse {
		return sortedSet.getByScoreRangeReverse(min, max, offset, limit)
	}
	return sortedSet.getByScoreRange(min, max, offset, limit)
}

func (sortedSet *SortedSet) getByScoreRange(min, max float64, offset, limit int64) []*Element {
	// find start node
	//var i int64 = 0 // used for offset
	n := sortedSet.header

	// skip to the first node with score >= min
	for i := sortedSet.level - 1; i >= 0; i-- {
		for n.level[i].forward != nil && n.level[i].forward.Score < min {
			n = n.level[i].forward
		}
	}

	// move to the next node (first node with score >= min)
	n = n.level[0].forward

	// skip offset nodes
	for n != nil && offset > 0 {
		offset--
		n = n.level[0].forward
	}

	var result []*Element
	// get all nodes with score <= max
	for n != nil && n.Score <= max && (limit < 0 || limit > 0) {
		result = append(result, &Element{
			Member: n.Member,
			Score:  n.Score,
		})

		if limit > 0 {
			limit--
		}

		n = n.level[0].forward
	}

	return result
}

func (sortedSet *SortedSet) getByScoreRangeReverse(min, max float64, offset, limit int64) []*Element {
	var result []*Element

	// get the last node
	n := sortedSet.tail

	// skip nodes with score > max
	for n != nil && n.Score > max {
		n = n.backward
	}

	// skip offset nodes
	for n != nil && offset > 0 {
		offset--
		n = n.backward
	}

	// get all nodes with score >= min
	for n != nil && n.Score >= min && (limit < 0 || limit > 0) {
		result = append(result, &Element{
			Member: n.Member,
			Score:  n.Score,
		})

		if limit > 0 {
			limit--
		}

		n = n.backward
	}

	return result
}

// GetByLexRange returns members with member in the given range
func (sortedSet *SortedSet) GetByLexRange(min, max string, offset, limit int64, reverse bool) []*Element {
	if reverse {
		return sortedSet.getByLexRangeReverse(min, max, offset, limit)
	}
	return sortedSet.getByLexRange(min, max, offset, limit)
}

func (sortedSet *SortedSet) getByLexRange(min, max string, offset, limit int64) []*Element {
	// find start node
	//var i int64 = 0 // used for offset
	n := sortedSet.header

	// skip to the first node with member >= min
	for i := sortedSet.level - 1; i >= 0; i-- {
		for n.level[i].forward != nil && n.level[i].forward.Member < min {
			n = n.level[i].forward
		}
	}

	// move to the next node (first node with member >= min)
	n = n.level[0].forward

	// skip offset nodes
	for n != nil && offset > 0 {
		offset--
		n = n.level[0].forward
	}

	var result []*Element
	// get all nodes with member <= max
	for n != nil && n.Member <= max && (limit < 0 || limit > 0) {
		result = append(result, &Element{
			Member: n.Member,
			Score:  n.Score,
		})

		if limit > 0 {
			limit--
		}

		n = n.level[0].forward
	}

	return result
}

func (sortedSet *SortedSet) getByLexRangeReverse(min, max string, offset, limit int64) []*Element {
	var result []*Element

	// get the last node
	n := sortedSet.tail

	// skip nodes with member > max
	for n != nil && n.Member > max {
		n = n.backward
	}

	// skip offset nodes
	for n != nil && offset > 0 {
		offset--
		n = n.backward
	}

	// get all nodes with member >= min
	for n != nil && n.Member >= min && (limit < 0 || limit > 0) {
		result = append(result, &Element{
			Member: n.Member,
			Score:  n.Score,
		})

		if limit > 0 {
			limit--
		}

		n = n.backward
	}

	return result
}

// Count returns the number of elements with score between min and max
func (sortedSet *SortedSet) Count(min, max float64) int64 {
	return int64(len(sortedSet.GetByScoreRange(min, max, 0, -1, false)))
}

// RangeCount returns the number of elements with member between min and max
func (sortedSet *SortedSet) RangeCount(min, max string) int64 {
	return int64(len(sortedSet.GetByLexRange(min, max, 0, -1, false)))
}

// Len returns the number of elements in the sorted set
func (sortedSet *SortedSet) Len() int64 {
	return sortedSet.length
}

func (sortedSet *SortedSet) getByMember(member string) *node {
	n := sortedSet.header

	for i := sortedSet.level - 1; i >= 0; i-- {
		for n.level[i].forward != nil && n.level[i].forward.Member < member {
			n = n.level[i].forward
		}
	}

	n = n.level[0].forward
	if n != nil && n.Member == member {
		return n
	}

	return nil
}

// ForEach traverses the sorted set and executes the given function on each element
func (sortedSet *SortedSet) ForEach(fn func(element *Element) bool) {
	n := sortedSet.header.level[0].forward

	for n != nil {
		if !fn(&Element{
			Member: n.Member,
			Score:  n.Score,
		}) {
			break
		}
		n = n.level[0].forward
	}
}

// Range traverses the sorted set in given range
func (sortedSet *SortedSet) Range(start, stop int64, reverse bool, fn func(element *Element) bool) {
	if reverse {
		// handle negative indexes
		if start < 0 {
			start = sortedSet.length + start
			if start < 0 {
				start = 0
			}
		}
		if stop < 0 {
			stop = sortedSet.length + stop
			if stop < 0 {
				stop = 0
			}
		}

		// swap if start is greater than stop
		if start > stop {
			start, stop = stop, start
		}

		n := sortedSet.tail
		i := sortedSet.length - 1

		// skip elements before start
		for n != nil && i > start {
			n = n.backward
			i--
		}

		// traverse elements in range
		for n != nil && i >= start && i <= stop {
			if !fn(&Element{
				Member: n.Member,
				Score:  n.Score,
			}) {
				break
			}
			n = n.backward
			i--
		}
	} else {
		// handle negative indexes
		if start < 0 {
			start = sortedSet.length + start
			if start < 0 {
				start = 0
			}
		}
		if stop < 0 {
			stop = sortedSet.length + stop
			if stop < 0 {
				stop = 0
			}
		}

		// swap if start is greater than stop
		if start > stop {
			start, stop = stop, start
		}

		n := sortedSet.header.level[0].forward
		i := int64(0)

		// skip elements before start
		for n != nil && i < start {
			n = n.level[0].forward
			i++
		}

		// traverse elements in range
		for n != nil && i >= start && i <= stop {
			if !fn(&Element{
				Member: n.Member,
				Score:  n.Score,
			}) {
				break
			}
			n = n.level[0].forward
			i++
		}
	}
}
