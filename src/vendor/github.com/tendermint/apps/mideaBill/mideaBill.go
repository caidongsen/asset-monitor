package mideaBill

import (
	//"encoding/hex"
	//"log"

	"github.com/tendermint/abci/types"
	"github.com/tendermint/merkleeyes/iavl"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/merkle"
)

type MideaBillApplication struct {
	types.BaseApplication

	state merkle.Tree
}

func NewMideaBillApplication() *MideaBillApplication {
	state := iavl.NewIAVLTree(0, nil)
	return &MideaBillApplication{state: state}
}

func (app *MideaBillApplication) Info() (resInfo types.ResponseInfo) {
	return types.ResponseInfo{Data: cmn.Fmt("{\"size\":%v}", app.state.Size())}
}

// tx is either "key=value" or just arbitrary bytes
func (app *MideaBillApplication) DeliverTx(tx []byte) types.Result {
	result, err := app.Exec(tx)
	if err != nil {
		println("app.DeliverTx", err.Error())
		return types.NewResult(types.CodeType_InternalError, nil, err.Error())
	}
	return types.NewResultOK(result, "")
	/*parts:= strings.Split(string(tx), "=")
	if len(parts) == 2 {
		app.state.Set([]byte(parts[0]), []byte(parts[1]))
	} else {
		app.state.Set(tx, tx)
	}
	return types.OK*/
}

func (app *MideaBillApplication) CheckTx(tx []byte) types.Result {
	err := app.Check(tx)
	if err != nil {
		println("app.CheckTx:", err.Error())
		return types.NewResult(types.CodeType_InternalError, nil, err.Error())
	}
	return types.OK
}

func (app *MideaBillApplication) Commit() types.Result {
	hash := app.state.Hash()
	return types.NewResultOK(hash, "")
}

func (app *MideaBillApplication) Query(reqQuery types.RequestQuery) (resQuery types.ResponseQuery) {
	if reqQuery.Path == "check" {
		err := app.Check(reqQuery.Data)
		if err != nil {
			resQuery.Code = -1
			resQuery.Log = err.Error()
		} else {
			resQuery.Code = 0
		}

		return
	}

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

func (app *MideaBillApplication) Check(tx []byte) error {
	var req Request
	err := UnmarshalMessage(tx, &req)
	if err != nil {
		return err
	}

	err = req.CheckSign()
	if err != nil {
		return err
	}

	err = app.checkInstructionId(req.GetInstructionId())
	if err != nil {
		return err
	}

	err = app.doCheck(&req)
	if err != nil {
		return err
	}
	return nil
}

func (app *MideaBillApplication) Exec(tx []byte) ([]byte, error) {
	var req Request
	var resp *Response
	err := UnmarshalMessage(tx, &req)
	if err != nil {
		return nil, err
	}

	resp, err = app.doRequest(&req)
	if err != nil {
		return nil, err
	} else {
		return MarshalMessage(resp)
	}
}

func (app *MideaBillApplication) doCheck(req *Request) error {
	var err error

	if req.GetActionId() != MessageType_MsgInitPlatform &&
	   req.GetActionId() != MessageType_MsgRegisterUser {
		err = app.checkUserSign(req)
		if err != nil {
			return err
		}
	}

	switch req.GetActionId() {
	case MessageType_MsgInitPlatform:
		err = app.checkInitPlatform(req)
	case MessageType_MsgRegisterUser:
		err = app.checkRegisterUser(req)
	case MessageType_MsgAddUser:
		err = app.checkAddUser(req)
	case MessageType_MsgUserPwdModify:
		err = app.checkUserPwdModify(req)
	case MessageType_MsgUserPwdReset:
		err = app.checkUserPwdReset(req)
	case MessageType_MsgEntIdentifyCheck:
		err = app.checkEntIdentifyCheck(req)
	case MessageType_MsgEntInfoModify:
		err = app.checkEntInfoModify(req)
	case MessageType_MsgApplyBill:
		err = app.checkApplyBill(req)
	case MessageType_MsgApplyBillSign:
		err = app.checkApplyBillSign(req)
	case MessageType_MsgApplyBillSignRefuse:
		err = app.checkApplyBillSignRefuse(req)
	case MessageType_MsgApplyBillSignCancle:
		err = app.checkApplyBillSignCancle(req)
	case MessageType_MsgBillTotalTransfer:
		err = app.checkBillTotalTransfer(req)
	case MessageType_MsgBillPartTransfer:
		err = app.checkBillPartTransfer(req)
	case MessageType_MsgBillTransferSign:
		err = app.checkBillTransferSign(req)
	case MessageType_MsgBillTransferRefuse:
		err = app.checkBillTransferRefuse(req)
	case MessageType_MsgBillTransferCancle:
		err = app.checkBillTransferCancle(req)
	case MessageType_MsgBillTransferForcePay:
		err = app.checkBillTransferForcePay(req)
	case MessageType_MsgBillTotalFinancing:
		err = app.checkBillTotalFinancing(req)
	case MessageType_MsgBillPartFinancing:
		err = app.checkBillPartFinancing(req)
	case MessageType_MsgBillFinancingCheckOk:
		err = app.checkBillFinancingCheckOk(req)
	case MessageType_MsgBillFinancingCheckFail:
		err = app.checkBillFinancingCheckFail(req)
	case MessageType_MsgBillFinancingFail:
		err = app.checkBillFinancingFail(req)
	case MessageType_MsgBillPay:
		err = app.checkBillPay(req)
	default:
		err = ErrWrongMessageType
	}
	return err
}

func (app *MideaBillApplication) doRequest(req *Request) (*Response, error) {
	var resp *Response
	var err error
	switch req.GetActionId() {
	case MessageType_MsgInitPlatform:
		resp, err = app.initPlatform(req)
	case MessageType_MsgRegisterUser:
		resp, err = app.registerUser(req)
	case MessageType_MsgAddUser:
		resp, err = app.addUser(req)
	case MessageType_MsgUserPwdModify:
		resp, err = app.userPwdModify(req)
	case MessageType_MsgUserPwdReset:
		resp, err = app.userPwdReset(req)
	case MessageType_MsgEntIdentifyCheck:
		resp, err = app.entIdentifyCheck(req)
	case MessageType_MsgEntInfoModify:
		resp, err = app.entInfoModify(req)
	case MessageType_MsgApplyBill:
		resp, err = app.applyBill(req)
	case MessageType_MsgApplyBillSign:
		resp, err = app.applyBillSign(req)
	case MessageType_MsgApplyBillSignRefuse:
		resp, err = app.applyBillSignRefuse(req)
	case MessageType_MsgApplyBillSignCancle:
		resp, err = app.applyBillSignCancle(req)
	case MessageType_MsgBillTotalTransfer:
		resp, err = app.billTotalTransfer(req)
	case MessageType_MsgBillPartTransfer:
		resp, err = app.billPartTransfer(req)
	case MessageType_MsgBillTransferSign:
		resp, err = app.billTransferSign(req)
	case MessageType_MsgBillTransferRefuse:
		resp, err = app.billTransferRefuse(req)
	case MessageType_MsgBillTransferCancle:
		resp, err = app.billTransferCancle(req)
	case MessageType_MsgBillTransferForcePay:
		resp, err = app.billTransferForcePay(req)
	case MessageType_MsgBillTotalFinancing:
		resp, err = app.billTotalFinancing(req)
	case MessageType_MsgBillPartFinancing:
		resp, err = app.billPartFinancing(req)
	case MessageType_MsgBillFinancingCheckOk:
		resp, err = app.billFinancingCheckOk(req)
	case MessageType_MsgBillFinancingCheckFail:
		resp, err = app.billFinancingCheckFail(req)
	case MessageType_MsgBillFinancingFail:
		resp, err = app.billFinancingFail(req)
	case MessageType_MsgBillPay:
		resp, err = app.billPay(req)
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
