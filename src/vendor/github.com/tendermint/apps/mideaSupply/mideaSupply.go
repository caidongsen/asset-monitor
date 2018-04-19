package mideaSupply

import (
	"fmt"
	"github.com/tendermint/abci/types"
	"github.com/tendermint/merkleeyes/iavl"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/merkle"
)

type MideaSupplyApplication struct {
	types.BaseApplication

	state merkle.Tree
}

func NewMideaSupplyApplication() *MideaSupplyApplication {
	state := iavl.NewIAVLTree(0, nil)
	return &MideaSupplyApplication{state: state}
}

func (app *MideaSupplyApplication) Info() (resInfo types.ResponseInfo) {
	return types.ResponseInfo{Data: cmn.Fmt("{\"size\":%v}", app.state.Size())}
}

// tx is either "key=value" or just arbitrary bytes
func (app *MideaSupplyApplication) DeliverTx(tx []byte) types.Result {
	result, err := app.Exec(tx)
	if err != nil {
		println("app.DeliverTx", err.Error())
		return types.NewResult(types.CodeType_InternalError, []byte(err.Error()), "")
	}
	return types.NewResultOK(result, "")
}

func (app *MideaSupplyApplication) CheckTx(tx []byte) types.Result {
	err := app.Check(tx)
	if err != nil {
		println("app.CheckTx:", err.Error())
		return types.NewResult(types.CodeType_InternalError, []byte(err.Error()), "")
	}
	return types.OK
	//return types.OK
}

func (app *MideaSupplyApplication) Commit() types.Result {
	hash := app.state.Hash()
	return types.NewResultOK(hash, "")
}

func (app *MideaSupplyApplication) Query(reqQuery types.RequestQuery) (resQuery types.ResponseQuery) {
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

func (app *MideaSupplyApplication) Check(tx []byte) error {
	var req Request
	err := UnmarshalMessage(tx, &req)
	if err != nil {
		return err
	}
	msgType := req.GetActionId()
	fmt.Println(msgType)
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

func (app *MideaSupplyApplication) Exec(tx []byte) ([]byte, error) {
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

func (app *MideaSupplyApplication) doCheck(req *Request, msgType MessageType) error {
	var err error
	switch msgType {
	case MessageType_MsgInitPlatform:
		err = app.checkInitPlatform(req)
	case MessageType_MsgRegisterSupplier:
		err = app.checkRegisterSupplier(req)
	case MessageType_MsgWarehouseEntry:
		err = app.checkWarehouseEntry(req)
	case MessageType_MsgOpenInvoice:
		err = app.checkOpenInvoice(req)
	case MessageType_MsgCheckInvoice:
		err = app.checkCheckInvoice(req)
	case MessageType_MsgChangePubkey:
		err = app.checkChangePubkey(req)
	case MessageType_MsgResetPubkey:
		err = app.checkResetPubkey(req)
	case MessageType_MsgRegisterSupplierList:
		err = app.checkRegisterSupplierList(req)
	case MessageType_MsgWarehouseEntryList:
		err = app.checkWarehouseEntryList(req)
	default:
		err = ErrWrongMessageType
	}
	return err
}

func (app *MideaSupplyApplication) doRequest(req *Request, msgType MessageType) (*Response, error) {
	var resp *Response
	var err error
	switch msgType {
	case MessageType_MsgInitPlatform:
		resp, err = app.initPlatform(req)
	case MessageType_MsgRegisterSupplier:
		resp, err = app.registerSupplier(req)
	case MessageType_MsgWarehouseEntry:
		resp, err = app.warehouseEntry(req)
	case MessageType_MsgOpenInvoice:
		resp, err = app.openInvoice(req)
	case MessageType_MsgCheckInvoice:
		resp, err = app.checkInvoice(req)
	case MessageType_MsgChangePubkey:
		resp, err = app.changePubkey(req)
	case MessageType_MsgResetPubkey:
		resp, err = app.resetPubkey(req)
	case MessageType_MsgRegisterSupplierList:
		resp, err = app.registerSupplierList(req)
	case MessageType_MsgWarehouseEntryList:
		resp, err = app.warehouseEntryList(req)
	default:
		err = ErrWrongMessageType
	}

	receipt := &Receipt{}
	if err != nil {
		receipt.Err = []byte(err.Error())
	}
	receipt.IsOk = (err == nil)
	app.saveReceipt(req.GetInstructionId(), receipt)

	return resp, err
}
