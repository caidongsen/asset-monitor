package provinvest

import (
	"github.com/tendermint/abci/types"
	"github.com/tendermint/merkleeyes/iavl"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/merkle"
)

type ProvinvestApplication struct {
	types.BaseApplication

	state merkle.Tree
}

func NewProvinvestApplication() *ProvinvestApplication {
	state := iavl.NewIAVLTree(0, nil)
	return &ProvinvestApplication{state: state}
}

func (app *ProvinvestApplication) Info() (resInfo types.ResponseInfo) {
	return types.ResponseInfo{Data: cmn.Fmt("{\"size\":%v}", app.state.Size())}
}

// tx is either "key=value" or just arbitrary bytes
func (app *ProvinvestApplication) DeliverTx(tx []byte) types.Result {
	result, err := app.Exec(tx)
	if err != nil {
		println("app.DeliverTx", err.Error())
		return types.NewResult(types.CodeType_InternalError, []byte(err.Error()), "")
	}
	return types.NewResultOK(result, "")
}

func (app *ProvinvestApplication) CheckTx(tx []byte) types.Result {
	err := app.Check(tx)
	if err != nil {
		println("app.CheckTx:", err.Error())
		return types.NewResult(types.CodeType_InternalError, []byte(err.Error()), "")
	}
	return types.OK
	//return types.OK
}

func (app *ProvinvestApplication) Commit() types.Result {
	hash := app.state.Hash()
	return types.NewResultOK(hash, "")
}

func (app *ProvinvestApplication) Query(reqQuery types.RequestQuery) (resQuery types.ResponseQuery) {
	if reqQuery.Prove {
		value, proof, exists := app.state.Proof(reqQuery.Data)
		resQuery.Index = -1 // TODO make Proof return index
		resQuery.Key = reqQuery.Data
		resQuery.Value = value
		resQuery.Proof = proof
		if exists {
			resQuery.Log = "exists"
		} else {
			resQuery.Log = "does not exist"
		}
		return
	} else {
		index, value, exists := app.state.Get(reqQuery.Data)
		resQuery.Index = int64(index)
		resQuery.Value = value
		if exists {
			resQuery.Log = "exists"
		} else {
			resQuery.Log = "does not exist"
		}
		return
	}
}

func (app *ProvinvestApplication) Check(tx []byte) error {
	var req Request
	err := UnmarshalMessage(tx, &req)
	if err != nil {
		return err
	}
	msgType := req.GetActionId()
	err = req.CheckSign()
	if err != nil {
		return err
	}
	var instructionId = req.GetInstructionId()
	if err := app.checkInstructionId(instructionId); err != nil {
		return err
	}
	err = app.doCheck(&req, msgType)
	if err != nil {
		return err
	}
	return nil
}

func (app *ProvinvestApplication) Exec(tx []byte) ([]byte, error) {
	var req Request
	var resp *Response
	err := UnmarshalMessage(tx, &req)
	if err != nil {
		return nil, err
	}
	msgType := req.GetActionId()
	resp, err = app.doRequest(&req, msgType)
	if err != nil {
		return nil, err
	} else {
		return MarshalMessage(resp)
	}
	panic("never happen")
}

func (app *ProvinvestApplication) doCheck(req *Request, msgType MessageType) error {
	var err error
	switch msgType {
	case MessageType_MsgOperationRecord:
		operationRecord := req.Value.(*Request_OperationRecord).OperationRecord
		err = app.checkOperationRecord(req.GetPubkey(), operationRecord.RecordVersion)
	default:
		err = ErrWrongMessageType
	}
	return err
}

func (app *ProvinvestApplication) doRequest(req *Request, msgType MessageType) (*Response, error) {
	var resp *Response
	var err error
	switch msgType {
	case MessageType_MsgOperationRecord:
		operationRecord := req.Value.(*Request_OperationRecord).OperationRecord
		resp, err = app.operationRecord(req.GetPubkey(), operationRecord.RecordVersion, operationRecord.Record, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	default:
		err = ErrWrongMessageType
	}
	return resp, err
}
