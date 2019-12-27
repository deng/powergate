package deals

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/filecoin-project/lotus/chain/address"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/textileio/filecoin/client"
	"github.com/textileio/filecoin/tests"
)

var (
	authToken = ""
)

func TestMain(m *testing.M) {
	var err error
	authToken, err = tests.GetLotusToken()
	if err != nil {
		fmt.Println("couldn't get/generate lotus authtoken")
		os.Exit(-1)
	}
	os.Exit(m.Run())
}

func TestAskCache(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test since we're on short mode")
	}
	c, cls, err := client.New(tests.DaemonAddr, authToken)
	checkErr(t, err)
	defer cls()

	queryAskRateLim = 2000
	asks, err := takeFreshAskSnapshot(c)
	checkErr(t, err)
	if len(asks) == 0 {
		t.Fatalf("current asks can't be empty")
	}
	fmt.Printf("ask cache: %v\n", asks)
}

func TestQueryAsk(t *testing.T) {
	t.Parallel()
	dm := DealModule{}
	dm.askCache = []*types.StorageAsk{
		{Price: types.NewInt(20), MinPieceSize: 128, Miner: newaddr("t01")},
		{Price: types.NewInt(30), MinPieceSize: 64, Miner: newaddr("t02")},
		{Price: types.NewInt(40), MinPieceSize: 256, Miner: newaddr("t03")},
		{Price: types.NewInt(50), MinPieceSize: 16, Miner: newaddr("t04")},
	}

	facr := []StorageAsk{
		{Price: 20, MinPieceSize: 128, Miner: "t01"},
		{Price: 30, MinPieceSize: 64, Miner: "t02"},
		{Price: 40, MinPieceSize: 256, Miner: "t03"},
		{Price: 50, MinPieceSize: 16, Miner: "t04"},
	}

	tests := []struct {
		name   string
		q      Query
		expect []StorageAsk
	}{
		{name: "All", q: Query{}, expect: facr},
		{name: "LeqPrice35", q: Query{MaxPrice: 35}, expect: []StorageAsk{facr[0], facr[1]}},
		{name: "LeqPrice50", q: Query{MaxPrice: 50}, expect: facr},
		{name: "LeqPrice40Piece96", q: Query{MaxPrice: 35, PieceSize: 96}, expect: []StorageAsk{facr[1]}},
		{name: "AllLimit2Offset1", q: Query{Limit: 2, Offset: 1}, expect: []StorageAsk{facr[1], facr[2]}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := dm.AvailableAsks(tt.q)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(tt.expect, got) {
				t.Fatalf("expected %v, got %v", tt.expect, got)
			}
		})
	}
}

func newaddr(s string) address.Address {
	a, err := address.NewFromString(s)
	if err != nil {
		panic("invalid address")
	}
	return a
}

func checkErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}