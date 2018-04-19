package apps

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/tendermint/abci/types"
	"github.com/tendermint/tmlibs/merkle"
)

type SupplyBaseApplication struct {
	types.BaseApplication
	CurrentHeader *types.Header
	CurrentHash   []byte
}

//func (app *BaseApplication) DeliverTxs(tx [][]byte) []Result {
//	panic("DeliverTxs not support")
//}

func (app *SupplyBaseApplication) BeginBlock(hash []byte, header *types.Header) {
	app.CurrentHash = hash
	app.CurrentHeader = header
}

func CalcMerkle(mtr [][]byte) (res []byte) {
	var j, i2 int
	for siz := len(mtr); siz > 1; siz = (siz + 1) / 2 {
		for i := 0; i < siz; i += 2 {
			if i+1 < siz-1 {
				i2 = i + 1
			} else {
				i2 = siz - 1
			}
			s := sha256.New()
			s.Write(mtr[j+i])
			s.Write(mtr[j+i2])
			tmp := s.Sum(nil)
			s.Reset()
			s.Write(tmp)
			mtr = append(mtr, s.Sum(nil))
		}
		j += siz
	}
	res = mtr[len(mtr)-1]
	return
}

func GetMerkle(txs []merkle.Hashable) (res []byte) {
	mtr := make([][]byte, len(txs))
	for i := range txs {
		mtr[i] = txs[i].Hash()
	}
	res = CalcMerkle(mtr)
	return
}

func HexToPrivkey(priv string) *[64]byte {
	adminKey, err := hex.DecodeString(priv)
	if err != nil {
		panic(err)
	}
	if len(adminKey) != 64 {
		panic("priv key len is not 64")
	}
	var data [64]byte
	copy(data[:], adminKey)
	return &data
}
