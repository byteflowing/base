package balancer

import (
	"math"
	"sync"

	"github.com/byteflowing/base/ecode"
)

type Node struct {
	ID            any
	Weight        int64
	CurrentWeight int64
}

type EqualFn func(a, b any) bool

type SWRRBalancer struct {
	mu                 sync.RWMutex
	nodes              []*Node
	totalWeight        int64
	normalizeThreshold int64
	equalFn            EqualFn
}

func NewSWRRBalancer(equal EqualFn) *SWRRBalancer {
	if equal == nil {
		panic("equal function is required")
	}
	return &SWRRBalancer{
		nodes:              make([]*Node, 0),
		equalFn:            equal,
		normalizeThreshold: 1 << 50,
	}
}

func (s *SWRRBalancer) Add(id any, weight int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, n := range s.nodes {
		if s.equalFn(n.ID, id) {
			n.Weight = weight
			n.CurrentWeight = 0
			s.recalculateTotalWeight()
			return
		}
	}
	s.nodes = append(s.nodes, &Node{ID: id, Weight: weight})
	s.recalculateTotalWeight()
}

func (s *SWRRBalancer) Remove(id any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	idx := -1
	for i, n := range s.nodes {
		if s.equalFn(n.ID, id) {
			idx = i
			break
		}
	}
	if idx == -1 {
		return
	}
	s.nodes = append(s.nodes[:idx], s.nodes[idx+1:]...)
	s.resetAllCurrentWeight()
	s.recalculateTotalWeight()
}

func (s *SWRRBalancer) BatchRemove(ids []any) {
	if len(ids) == 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	toRemove := make(map[any]struct{}, len(ids))
	for _, id := range ids {
		toRemove[id] = struct{}{}
	}
	newNodes := make([]*Node, 0, len(s.nodes))
	for _, n := range s.nodes {
		if _, ok := toRemove[n.ID]; !ok {
			newNodes = append(newNodes, n)
		}
	}
	s.nodes = newNodes
	s.resetAllCurrentWeight()
	s.recalculateTotalWeight()
}

func (s *SWRRBalancer) UpdateWeight(id any, weight int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, n := range s.nodes {
		if s.equalFn(n.ID, id) {
			n.Weight = weight
			n.CurrentWeight = 0
			s.recalculateTotalWeight()
			return
		}
	}
	s.nodes = append(s.nodes, &Node{ID: id, Weight: weight, CurrentWeight: 0})
	s.recalculateTotalWeight()
}

func (s *SWRRBalancer) Next() (*Node, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.nodes) == 0 {
		return nil, ecode.ErrNoResource
	}
	if len(s.nodes) == 1 {
		n := s.nodes[0]
		if n.Weight <= 0 {
			return nil, ecode.ErrNoResource
		}
		n.CurrentWeight += n.Weight
		n.CurrentWeight -= s.totalWeight
		s.maybeNormalizeLocked()
		return n, nil
	}
	var best *Node
	for _, n := range s.nodes {
		if n.Weight <= 0 {
			continue
		}
		n.CurrentWeight += n.Weight
		if best == nil || n.CurrentWeight > best.CurrentWeight {
			best = n
		}
	}
	if best == nil {
		return nil, ecode.ErrNoResource
	}
	best.CurrentWeight -= s.totalWeight
	s.maybeNormalizeLocked()
	return best, nil
}

func (s *SWRRBalancer) Nodes() []Node {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]Node, len(s.nodes))
	for i, n := range s.nodes {
		res[i] = Node{ID: n.ID, Weight: n.Weight, CurrentWeight: n.CurrentWeight}
	}
	return res
}

func (s *SWRRBalancer) recalculateTotalWeight() {
	total := int64(0)
	for _, node := range s.nodes {
		if node.Weight > 0 {
			total += node.Weight
		}
	}
	s.totalWeight = total
}

func (s *SWRRBalancer) resetAllCurrentWeight() {
	for _, node := range s.nodes {
		node.CurrentWeight = 0
	}
}

func (s *SWRRBalancer) maybeNormalizeLocked() {
	if s.normalizeThreshold <= 0 {
		return
	}
	// find min and max cw
	var maxCW int64 = math.MinInt64
	var minCW int64 = math.MaxInt64
	for _, n := range s.nodes {
		if n.CurrentWeight > maxCW {
			maxCW = n.CurrentWeight
		}
		if n.CurrentWeight < minCW {
			minCW = n.CurrentWeight
		}
	}
	// if values too large in magnitude, subtract minCW (so min becomes 0)
	if maxCW > s.normalizeThreshold || minCW < -s.normalizeThreshold {
		// normalize by subtracting minCW (keep differences)
		for _, n := range s.nodes {
			n.CurrentWeight -= minCW
		}
	}
}
