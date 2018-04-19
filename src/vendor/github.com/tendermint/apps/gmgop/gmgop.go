package gmgop

import (
	"github.com/tendermint/abci/types"
	"github.com/tendermint/merkleeyes/iavl"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/merkle"
)

type GmgopApplication struct {
	types.BaseApplication

	state merkle.Tree
}

func NewGmgopApplication() *GmgopApplication {
	state := iavl.NewIAVLTree(0, nil)
	return &GmgopApplication{state: state}
}

func (app *GmgopApplication) Info() (resInfo types.ResponseInfo) {
	return types.ResponseInfo{Data: cmn.Fmt("{\"size\":%v}", app.state.Size())}
}

// tx is either "key=value" or just arbitrary bytes
func (app *GmgopApplication) DeliverTx(tx []byte) types.Result {
	result, err := app.Exec(tx)
	if err != nil {
		println("app.DeliverTx", err.Error())
		return types.NewResult(types.CodeType_InternalError, []byte(err.Error()), "")
	}
	return types.NewResultOK(result, "")
	/*parts := strings.Split(string(tx), "=")
	if len(parts) == 2 {
		app.state.Set([]byte(parts[0]), []byte(parts[1]))
	} else {
		app.state.Set(tx, tx)
	}
	return types.OK*/
}

func (app *GmgopApplication) CheckTx(tx []byte) types.Result {
	err := app.Check(tx)
	if err != nil {
		println("app.CheckTx:", err.Error())
		return types.NewResult(types.CodeType_InternalError, []byte(err.Error()), "")
	}
	return types.OK
	//return types.OK
}

func (app *GmgopApplication) Commit() types.Result {
	hash := app.state.Hash()
	return types.NewResultOK(hash, "")
}

func (app *GmgopApplication) Query(reqQuery types.RequestQuery) (resQuery types.ResponseQuery) {
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

func (app *GmgopApplication) Check(tx []byte) error {
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

func (app *GmgopApplication) Exec(tx []byte) ([]byte, error) {
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

func (app *GmgopApplication) doCheck(req *Request, msgType MessageType) error {
	var err error
	switch msgType {
	case MessageType_MsgUserCreate:
		userCreate := req.Value.(*Request_UserCreate).UserCreate
		err = app.checkUserCreate(req.GetPubkey(), userCreate.Uid)
	case MessageType_MsgGopAssertCreate:
		gopAssertCreate := req.Value.(*Request_GopAssertCreate).GopAssertCreate
		err = app.checkGopAssertCreate(req.GetUid(), req.GetPubkey(), gopAssertCreate.AssertId)
	case MessageType_MsgGopAssertDelete:
		gopAssertDelete := req.Value.(*Request_GopAssertDelete).GopAssertDelete
		err = app.checkGopAssertDelete(req.GetUid(), req.GetPubkey(), gopAssertDelete.AssertId)
	case MessageType_MsgBuContractCreate:
		buContractCreate := req.Value.(*Request_BuContractCreate).BuContractCreate
		err = app.checkBuContractCreate(req.GetUid(), req.GetPubkey(), buContractCreate.AssertId, buContractCreate.ContractId, buContractCreate.Price)
	case MessageType_MsgGopContractAgree:
		gopContractAgree := req.Value.(*Request_GopContractAgree).GopContractAgree
		err = app.checkGopContractAgree(req.GetUid(), req.GetPubkey(), gopContractAgree.ContractId)
	case MessageType_MsgGopContractReject:
		gopContractReject := req.Value.(*Request_GopContractReject).GopContractReject
		err = app.checkGopContractReject(req.GetUid(), req.GetPubkey(), gopContractReject.ContractId)
	case MessageType_MsgBuRecallContract:
		buRecallContract := req.Value.(*Request_BuRecallContract).BuRecallContract
		err = app.checkBuRecallContract(req.GetUid(), req.GetPubkey(), buRecallContract.ContractId)
	case MessageType_MsgGopDeliverAssert:
		gopDeliverAssert := req.Value.(*Request_GopDeliverAssert).GopDeliverAssert
		err = app.checkGopDeliverAssert(req.GetUid(), req.GetPubkey(), gopDeliverAssert.ContractId)
	case MessageType_MsgBuDeliverAssertConfirm:
		buDeliverAssertConfirm := req.Value.(*Request_BuDeliverAssertConfirm).BuDeliverAssertConfirm
		err = app.checkBuDeliverAssertConfirm(req.GetUid(), req.GetPubkey(), buDeliverAssertConfirm.ContractId)
	case MessageType_MsgGopFin:
		gopFin := req.Value.(*Request_GopFin).GopFin
		err = app.checkGopFin(req.GetUid(), req.GetPubkey(), gopFin.ContractId)
	case MessageType_MsgChangePubkey:
		changePubkey := req.Value.(*Request_ChangePubkey).ChangePubkey
		err = app.checkChangePubkey(req.GetUid(), req.GetPubkey(), changePubkey.NewPubkey)
	default:
		err = ErrWrongMessageType
	}
	return err
}

func (app *GmgopApplication) doRequest(req *Request, msgType MessageType) (*Response, error) {
	var resp *Response
	var err error
	switch msgType {
	case MessageType_MsgUserCreate:
		userCreate := req.Value.(*Request_UserCreate).UserCreate
		resp, err = app.userCreate(userCreate.Uid, req.GetPubkey(), userCreate.Pubkey, userCreate.Info, userCreate.Role, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgGopAssertCreate:
		gopAssertCreate := req.Value.(*Request_GopAssertCreate).GopAssertCreate
		resp, err = app.gopAssertCreate(req.GetUid(), req.GetPubkey(), gopAssertCreate.AssertId, gopAssertCreate.Price, gopAssertCreate.Info, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgGopAssertDelete:
		gopAssertDelete := req.Value.(*Request_GopAssertDelete).GopAssertDelete
		resp, err = app.gopAssertDelete(req.GetUid(), req.GetPubkey(), gopAssertDelete.AssertId, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgBuContractCreate:
		buContractCreate := req.Value.(*Request_BuContractCreate).BuContractCreate
		resp, err = app.buContractCreate(req.GetUid(), req.GetPubkey(), buContractCreate.AssertId, buContractCreate.ContractId, buContractCreate.Price, buContractCreate.StartT, buContractCreate.EndT, buContractCreate.Info, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgGopContractAgree:
		gopContractAgree := req.Value.(*Request_GopContractAgree).GopContractAgree
		resp, err = app.gopContractAgree(req.GetUid(), req.GetPubkey(), gopContractAgree.ContractId, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgGopContractReject:
		gopContractReject := req.Value.(*Request_GopContractReject).GopContractReject
		resp, err = app.gopContractReject(req.GetUid(), req.GetPubkey(), gopContractReject.ContractId, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgBuRecallContract:
		buRecallContract := req.Value.(*Request_BuRecallContract).BuRecallContract
		resp, err = app.buRecallContract(req.GetUid(), req.GetPubkey(), buRecallContract.ContractId, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgGopDeliverAssert:
		gopDeliverAssert := req.Value.(*Request_GopDeliverAssert).GopDeliverAssert
		resp, err = app.gopDeliverAssert(req.GetUid(), req.GetPubkey(), gopDeliverAssert.ContractId, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgBuDeliverAssertConfirm:
		buDeliverAssertConfirm := req.Value.(*Request_BuDeliverAssertConfirm).BuDeliverAssertConfirm
		resp, err = app.buDeliverAssertConfirm(req.GetUid(), req.GetPubkey(), buDeliverAssertConfirm.ContractId, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgGopFin:
		gopFin := req.Value.(*Request_GopFin).GopFin
		resp, err = app.gopFin(req.GetUid(), req.GetPubkey(), gopFin.ContractId, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgChangePubkey:
		changePubkey := req.Value.(*Request_ChangePubkey).ChangePubkey
		resp, err = app.changePubkey(req.GetUid(), req.GetPubkey(), changePubkey.NewPubkey, req.GetInstructionId())
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
