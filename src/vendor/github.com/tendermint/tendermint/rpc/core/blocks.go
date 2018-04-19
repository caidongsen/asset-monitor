package core

import (
	"fmt"

	abci "github.com/tendermint/abci/types"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
	. "github.com/tendermint/tmlibs/common"
)

//-----------------------------------------------------------------------------

// TODO: limit/permission on (max - min)
func BlockchainInfo(minHeight, maxHeight int) (*ctypes.ResultBlockchainInfo, error) {
	if maxHeight == 0 {
		maxHeight = blockStore.Height()
	} else {
		maxHeight = MinInt(blockStore.Height(), maxHeight)
	}
	if minHeight == 0 {
		minHeight = MaxInt(1, maxHeight-20)
	}
	logger.Debug("BlockchainInfoHandler", "maxHeight", maxHeight, "minHeight", minHeight)

	blockMetas := []*types.BlockMeta{}
	for height := maxHeight; height >= minHeight; height-- {
		blockMeta := blockStore.LoadBlockMeta(height)
		blockMetas = append(blockMetas, blockMeta)
	}

	return &ctypes.ResultBlockchainInfo{blockStore.Height(), blockMetas}, nil
}

//-----------------------------------------------------------------------------

func Block(height int) (*ctypes.ResultBlock, error) {
	if height == 0 {
		return nil, fmt.Errorf("Height must be greater than 0")
	}
	if height > blockStore.Height() {
		return nil, fmt.Errorf("Height must be less than the current blockchain height")
	}

	blockMeta := blockStore.LoadBlockMeta(height)
	block := blockStore.LoadBlock(height)
	blockTxsResults, err := BlockTxsResults(block)
	if err != nil {
		return nil, fmt.Errorf("TxsResults error: %v", err.Error())
	}
	return &ctypes.ResultBlock{blockMeta, block, blockTxsResults}, nil
}

//-----------------------------------------------------------------------------

func Commit(height int) (*ctypes.ResultCommit, error) {
	if height == 0 {
		return nil, fmt.Errorf("Height must be greater than 0")
	}
	storeHeight := blockStore.Height()
	if height > storeHeight {
		return nil, fmt.Errorf("Height must be less than or equal to the current blockchain height")
	}

	header := blockStore.LoadBlockMeta(height).Header

	// If the next block has not been committed yet,
	// use a non-canonical commit
	if height == storeHeight {
		commit := blockStore.LoadSeenCommit(height)
		return &ctypes.ResultCommit{header, commit, false}, nil
	}

	// Return the canonical commit (comes from the block at height+1)
	commit := blockStore.LoadBlockCommit(height)
	return &ctypes.ResultCommit{header, commit, true}, nil
}

func BlockTxsResults(block *types.Block) ([]*abci.Result, error) {
	txsResults := []*abci.Result{}
	for i := 0; i < len(block.Data.Txs); i++ {
		hash := block.Data.Txs[i].Hash()
		txResult, err := Tx(hash, false)
		if err != nil {
			return nil, err
		}
		txsResults = append(txsResults, &txResult.TxResult)
	}

	return txsResults, nil
}
