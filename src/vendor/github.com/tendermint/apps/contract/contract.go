package contract

import (
	"github.com/tendermint/abci/types"
	"github.com/tendermint/merkleeyes/iavl"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/merkle"
)

type ContractApplication struct {
	types.BaseApplication

	state merkle.Tree
}

func NewContractApplication() *ContractApplication {
	state := iavl.NewIAVLTree(0, nil)
	return &ContractApplication{state: state}
}

func (app *ContractApplication) Info() (resInfo types.ResponseInfo) {
	return types.ResponseInfo{Data: cmn.Fmt("{\"size\":%v}", app.state.Size())}
}

// tx is either "key=value" or just arbitrary bytes
func (app *ContractApplication) DeliverTx(tx []byte) types.Result {
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

func (app *ContractApplication) CheckTx(tx []byte) types.Result {
	err := app.Check(tx)
	if err != nil {
		println("app.CheckTx:", err.Error())
		return types.NewResult(types.CodeType_InternalError, []byte(err.Error()), "")
	}
	return types.OK
	//return types.OK
}

func (app *ContractApplication) Commit() types.Result {
	hash := app.state.Hash()
	return types.NewResultOK(hash, "")
}

func (app *ContractApplication) Query(reqQuery types.RequestQuery) (resQuery types.ResponseQuery) {
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

func (app *ContractApplication) Check(tx []byte) error {
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

func (app *ContractApplication) Exec(tx []byte) ([]byte, error) {
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

func (app *ContractApplication) doCheck(req *Request, msgType MessageType) error {
	var err error
	switch msgType {
	case MessageType_MsgUserCreate:
		userCreate := req.Value.(*Request_UserCreate).UserCreate
		err = app.checkUserCreate(req.GetPubkey(), userCreate.Uid)
	case MessageType_MsgChangePubkey:
		changePubkey := req.Value.(*Request_ChangePubkey).ChangePubkey
		err = app.checkChangePubkey(req.GetUid(), req.GetPubkey(), changePubkey.NewPubkey)
	case MessageType_MsgCreateContract:
		createContract := req.Value.(*Request_CreateContract).CreateContract
		err = app.checkCreateContract(req.GetUid(), req.GetPubkey(), createContract.Id)
	case MessageType_MsgLaunchSign:
		launchSign := req.Value.(*Request_LaunchSign).LaunchSign
		err = app.checkLaunchSign(req.GetUid(), req.GetPubkey(), launchSign.ContractId, launchSign.Signers, launchSign.EndT, req.GetOpT())
	case MessageType_MsgCreatorSign:
		creatorSign := req.Value.(*Request_CreatorSign).CreatorSign
		err = app.checkCreatorSign(req.GetUid(), req.GetPubkey(), creatorSign.ContractId, req.GetOpT())
	case MessageType_MsgEditContract:
		editContract := req.Value.(*Request_EditContract).EditContract
		err = app.checkEditContract(req.GetUid(), req.GetPubkey(), editContract.ContractId, req.GetOpT())
	case MessageType_MsgRejectContract:
		rejectContract := req.Value.(*Request_RejectContract).RejectContract
		err = app.checkRejectContract(req.GetUid(), req.GetPubkey(), rejectContract.ContractId, req.GetOpT())
	case MessageType_MsgAcceptContract:
		acceptContract := req.Value.(*Request_AcceptContract).AcceptContract
		err = app.checkAcceptContract(req.GetUid(), req.GetPubkey(), acceptContract.ContractId, req.GetOpT())
	case MessageType_MsgRepairContract:
		repairContract := req.Value.(*Request_RepairContract).RepairContract
		err = app.checkRepairContract(req.GetUid(), req.GetPubkey(), repairContract.ContractId, repairContract.EndT, req.GetOpT())
	case MessageType_MsgDisuseContract:
		disuseContract := req.Value.(*Request_DisuseContract).DisuseContract
		err = app.checkDisuseContract(req.GetUid(), req.GetPubkey(), disuseContract.ContractId, req.GetOpT())
	default:
		err = ErrWrongMessageType
	}
	return err
}

func (app *ContractApplication) doRequest(req *Request, msgType MessageType) (*Response, error) {
	var resp *Response
	var err error
	switch msgType {
	case MessageType_MsgUserCreate:
		userCreate := req.Value.(*Request_UserCreate).UserCreate
		resp, err = app.userCreate(userCreate.Uid, req.GetPubkey(), userCreate.Pubkey, userCreate.Info, req.GetInstructionId())
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
	case MessageType_MsgCreateContract:
		createContract := req.Value.(*Request_CreateContract).CreateContract
		resp, err = app.createContract(req.GetUid(), req.GetPubkey(), req.GetInstructionId(), createContract.Id, createContract.Info)
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgLaunchSign:
		launchSign := req.Value.(*Request_LaunchSign).LaunchSign
		resp, err = app.launchSign(req.GetUid(), req.GetPubkey(), req.GetInstructionId(), launchSign.ContractId, launchSign.EndT, req.GetOpT(), launchSign.Signers)
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgCreatorSign:
		creatorSign := req.Value.(*Request_CreatorSign).CreatorSign
		resp, err = app.creatorSign(req.GetUid(), req.GetPubkey(), req.GetInstructionId(), creatorSign.ContractId, req.GetOpT())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgEditContract:
		editContract := req.Value.(*Request_EditContract).EditContract
		resp, err = app.editContract(req.GetUid(), req.GetPubkey(), req.GetInstructionId(), editContract.ContractId, req.GetOpT(), editContract.Info)
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgRejectContract:
		rejectContract := req.Value.(*Request_RejectContract).RejectContract
		resp, err = app.rejectContract(req.GetUid(), req.GetPubkey(), req.GetInstructionId(), rejectContract.ContractId, req.GetOpT())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgAcceptContract:
		acceptContract := req.Value.(*Request_AcceptContract).AcceptContract
		resp, err = app.acceptContract(req.GetUid(), req.GetPubkey(), req.GetInstructionId(), acceptContract.ContractId, req.GetOpT())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgRepairContract:
		repairContract := req.Value.(*Request_RepairContract).RepairContract
		resp, err = app.repairContract(req.GetUid(), req.GetPubkey(), req.GetInstructionId(), repairContract.ContractId, repairContract.EndT, req.GetOpT(), repairContract.Info)
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgDisuseContract:
		disuseContract := req.Value.(*Request_DisuseContract).DisuseContract
		resp, err = app.disuseContract(req.GetUid(), req.GetPubkey(), req.GetInstructionId(), disuseContract.ContractId, req.GetOpT())
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
