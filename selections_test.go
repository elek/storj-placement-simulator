package uploadselection

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"storj.io/common/storj"
	"storj.io/common/testrand"
	"strconv"
	"testing"
)

func TestSpaceSelector(t *testing.T) {
	var nodes []*Node
	for i := 0; i < 100; i++ {
		wallet := "wallet" + strconv.Itoa(i)
		if i < 3 {
			wallet = "wallet"
		}
		nodes = append(nodes, &Node{
			PieceCount: 1 * (i + 1),
			NodeURL: storj.NodeURL{
				ID: testrand.NodeID(),
			},
			LastNet: "127.0.0." + strconv.Itoa(i),
			Wallet:  wallet,
		})

	}
	selector := NewSpaceSelector(nodes, []*Node{})
	i, err := selector.Select(context.Background(), Request{
		Count: 10,
	})
	require.NoError(t, err)
	for _, n := range i {
		fmt.Println(n.ID)
		fmt.Println(n.PieceCount)
	}
}
