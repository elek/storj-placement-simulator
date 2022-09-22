package uploadselection

import (
	"context"
	"crypto/rand"
	"math/big"
)

type NodeSelection interface {
	Select(ctx context.Context, request Request) (_ []*Node, err error)
}
type SpaceSelector struct {
	nodes []*Node
}

func NewSpaceSelector(reputableNodes, newNodes []*Node) *SpaceSelector {
	return &SpaceSelector{
		nodes: append(reputableNodes, newNodes...),
	}
}

func (s *SpaceSelector) Select(ctx context.Context, request Request) (_ []*Node, err error) {
	sum := 0
	for _, n := range s.nodes {
		sum += n.PieceCount
	}

	selected := []*Node{}

	for i := 0; i < request.Count; i++ {
		point, err := randomInt(sum)
		if err != nil {
			return nil, err
		}
		m := 0
		for _, c := range s.nodes {
			if point > m && point <= m+c.PieceCount {
				selected = append(selected, c)
				break
			}
			m += c.PieceCount
		}
	}
	return selected, nil
}

func randomInt(sum int) (int, error) {
	p, err := rand.Int(rand.Reader, big.NewInt(int64(sum)))
	if err != nil {
		return 0, err
	}
	point := int(p.Int64())
	return point, nil
}

type RandomSelector struct {
	reputableNodes []*Node
	newNodes       []*Node
}

func NewRandomSelector(reputableNodes, newNodes []*Node) *RandomSelector {
	return &RandomSelector{
		reputableNodes: reputableNodes,
		newNodes:       newNodes,
	}
}

func (s *RandomSelector) Select(ctx context.Context, request Request) (_ []*Node, err error) {
	newCount := int(float64(request.Count) * request.NewFraction)
	reputableCount := request.Count - newCount

	selected := []*Node{}
	for i := 0; i < newCount; i++ {
		point, err := randomInt(len(s.newNodes))
		if err != nil {
			return nil, err
		}
		selected = append(selected, s.newNodes[point])
	}

	for i := 0; i < reputableCount; i++ {
		point, err := randomInt(len(s.reputableNodes))
		if err != nil {
			return nil, err
		}
		selected = append(selected, s.reputableNodes[point])
	}
	return selected, nil
}
