package uploadselection

import (
	"context"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"github.com/zeebo/errs"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/aclements/go-moremath/stats"
	"github.com/stretchr/testify/require"

	"storj.io/common/storj"
)

func TestPlacement(t *testing.T) {
	ctx := context.Background()
	r, n, err := readNodes("nodes.csv")
	require.NoError(t, err)
	fmt.Printf("Loaded %d reputable and %d new nodes\n", len(r), len(n))

	selections := []NodeSelection{
		NewState(r, n),
		NewRandomSelector(r, n),
		NewSpaceSelector(r, n),
	}

	for _, s := range selections {
		newPiecesByNode := map[string]int{}
		newPiecesByNet := map[string]int{}
		newPiecesByWallet := map[string]int{}
		maxInfluenceWallet := 0
		maxInfluenceNet := 0
		maxInfluenceNode := 0

		//how many times --> controlled this amount of piece by one wallet (in one selection)
		influenceWallet := make(map[int]int)
		influenceNet := make(map[int]int)
		influenceNode := make(map[int]int)
		samples := 1000000
		for i := 0; i < samples; i++ {
			nodes, err := s.Select(ctx, Request{
				Count:       80,
				NewFraction: 0.05,
				Distinct:    true,
			})
			require.NoError(t, err)
			byNet := map[string]int{}
			byWallet := map[string]int{}
			byNode := map[string]int{}
			for _, n := range nodes {
				newPiecesByNode[n.ID.String()]++
				newPiecesByWallet[n.Wallet]++
				newPiecesByNet[n.LastNet]++

				byNet[n.LastNet]++
				byWallet[n.Wallet]++
				byNode[n.LastIPPort]++
			}
			for w, x := range byWallet {
				influenceWallet[x]++
				if x > maxInfluenceWallet {
					maxInfluenceWallet = x
					fmt.Printf("%s %d\n", w, x)
				}
			}
			for _, x := range byNet {
				influenceNet[x]++
				if x > maxInfluenceNet {
					maxInfluenceNet = x
				}
			}
			for _, x := range byNode {
				influenceNode[x]++
				if x > maxInfluenceNode {
					maxInfluenceNode = x
				}
			}
		}

		testName := strings.SplitN(fmt.Sprintf("%T", s), ".", 2)[1]
		fmt.Printf("-----------%s------------\n", testName)
		fmt.Println("node influence")
		fmt.Printf("max: %d\n", maxInfluenceNode)
		for k, v := range influenceNode {
			fmt.Printf("%d %d\n", k, v)
		}
		fmt.Println()
		fmt.Println("net influence")
		fmt.Printf("max: %d\n", maxInfluenceNet)
		for k, v := range influenceNet {
			fmt.Printf("%d %d\n", k, v)
		}
		fmt.Println()
		fmt.Println("wallet influence")
		fmt.Printf("max: %d\n", maxInfluenceWallet)
		for k, v := range influenceWallet {
			fmt.Printf("%d %d\n", k, v)
		}
		fmt.Println()
		showStat(fmt.Sprintf("%s by node", testName), newPiecesByNode)
		err := saveCsv(fmt.Sprintf("%s-nodes-with-pieces.csv", testName), hist(newPiecesByNet))
		require.NoError(t, err)

		showStat(fmt.Sprintf("%s by wallet", testName), newPiecesByWallet)
		err = saveCsv(fmt.Sprintf("%s-wallets-with-pieces.csv", testName), hist(newPiecesByWallet))
		require.NoError(t, err)

		showStat(fmt.Sprintf("%s by net", testName), newPiecesByNet)
		err = saveCsv(fmt.Sprintf("%s-nets-with-pieces.csv", testName), hist(newPiecesByNet))
		require.NoError(t, err)

	}
}

func hist(net map[string]int) map[int]int {
	res := make(map[int]int)
	for _, v := range net {
		res[v]++
	}
	return res
}

func saveCsv(file string, values map[int]int) error {
	f, err := os.Create(file)
	if err != nil {
		return errs.Wrap(err)
	}
	defer f.Close()
	for k, v := range values {
		_, err := f.WriteString(fmt.Sprintf("%d,%d\n", k, v))
		if err != nil {
			return err
		}
	}
	return nil
}

func showStat(name string, newPieces map[string]int) {
	values := []float64{}
	for _, v := range newPieces {
		values = append(values, float64(v))
	}
	fmt.Printf("%s\n", name)
	fmt.Println(len(newPieces))
	min, max := stats.Bounds(values)
	fmt.Printf("StdDev: %f\n", stats.StdDev(values))
	fmt.Printf("Min: %f\n", min)
	fmt.Printf("Max: %f\n", max)
	fmt.Printf("Mean: %f\n", stats.Mean(values))
	fmt.Println()
}

func readNodes(csvFile string) (reputableNodes []*Node, newNodes []*Node, err error) {
	reputableNodes = []*Node{}
	newNodes = []*Node{}

	file, err := os.Open(csvFile)
	if err != nil {
		return reputableNodes, newNodes, err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return reputableNodes, newNodes, err
	}

	columnIndex := map[string]int{}
	for ix, r := range records {
		if ix == 0 {
			for i, column := range r {
				columnIndex[column] = i
			}
			continue
		}

		idHex, err := hex.DecodeString(r[columnIndex["id"]])
		if err != nil {
			return reputableNodes, newNodes, err
		}
		id, err := storj.NodeIDFromBytes(idHex)
		if err != nil {
			return reputableNodes, newNodes, err
		}
		pieceCount, err := strconv.Atoi(r[columnIndex["piece_count"]])
		if err != nil {
			return reputableNodes, newNodes, err
		}

		node := Node{
			NodeURL: storj.NodeURL{
				ID:      id,
				Address: r[columnIndex["address"]],
			},
			LastIPPort: r[columnIndex["last_ip_port"]],
			LastNet:    r[columnIndex["last_net"]],
			PieceCount: pieceCount,
			Wallet:     r[columnIndex["wallet"]],
		}
		if r[columnIndex["vetted_at"]] != "" {
			reputableNodes = append(reputableNodes, &node)
		} else {
			newNodes = append(newNodes, &node)
		}
	}
	return reputableNodes, newNodes, err
}
