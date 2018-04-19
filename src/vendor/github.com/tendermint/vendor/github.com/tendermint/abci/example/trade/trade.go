// +build trade

package trade

import (
	"dev.33.cn/33/btrade/tradeserver"
	"github.com/tendermint/abci/types"
)

type TradeApplication struct {
	state *tradeserver.TradeServer
}

func NewTradeApplication(dbDir string) *TradeApplication {
	state, err := tradeserver.Init(dbDir, "trade")
	if err != nil {
		panic(err)
	}
	return &TradeApplication{state: state}
}

func (app *TradeApplication) Info() (resInfo types.ResponseInfo) {
	return app.state.Info()
}

func (app *TradeApplication) SetOption(key string, value string) (log string) {
	return app.state.SetOption(key, value)
}

func (app *TradeApplication) DeliverTx(tx []byte) types.Result {
	result, err := app.state.Exec(tx)
	if err != nil {
		return types.NewResult(types.CodeType_InternalError, []byte(err.Error()), "")
	}
	return types.NewResultOK(result, "")
}

func (app *TradeApplication) CheckTx(tx []byte) types.Result {
	err := app.state.Check(tx)
	if err != nil {
		return types.NewResult(types.CodeType_InternalError, []byte(err.Error()), "")
	}
	return types.OK
}

func (app *TradeApplication) Commit() types.Result {
	return app.state.Commit()
}

func (app *TradeApplication) Query(reqQuery types.RequestQuery) (resQuery types.ResponseQuery) {
	value, err := app.state.Query(reqQuery.Data)
	if err != nil {
		resQuery.Code = types.CodeType_InternalError
		resQuery.Log = err.Error()
	}

	resQuery.Key = reqQuery.Data
	resQuery.Value = value
	return
}

// Save the validators in the merkle tree
func (app *TradeApplication) InitChain(validators []*types.Validator) {

}

// Track the block hash and header information
func (app *TradeApplication) BeginBlock(hash []byte, header *types.Header) {
	app.state.BeginBlock(hash, header)
}

// Update the validator set
func (app *TradeApplication) EndBlock(height uint64) (resEndBlock types.ResponseEndBlock) {
	return app.state.EndBlock(height)
}
