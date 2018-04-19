package gfcollection

import (
	"github.com/tendermint/abci/types"
	"github.com/tendermint/merkleeyes/iavl"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/merkle"
)

type GfcollectionApplication struct {
	types.BaseApplication

	state merkle.Tree
}

var g_packId int64

func NewGfcollectionApplication() *GfcollectionApplication {
	state := iavl.NewIAVLTree(0, nil)
	return &GfcollectionApplication{state: state}
}

func (app *GfcollectionApplication) Info() (resInfo types.ResponseInfo) {
	return types.ResponseInfo{Data: cmn.Fmt("{\"size\":%v}", app.state.Size())}
}

// tx is either "key=value" or just arbitrary bytes
func (app *GfcollectionApplication) DeliverTx(tx []byte) types.Result {
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

func (app *GfcollectionApplication) CheckTx(tx []byte) types.Result {
	err := app.Check(tx)
	if err != nil {
		println("app.CheckTx:", err.Error())
		return types.NewResult(types.CodeType_InternalError, []byte(err.Error()), "")
	}
	return types.OK
	//return types.OK
}

func (app *GfcollectionApplication) Commit() types.Result {
	hash := app.state.Hash()
	return types.NewResultOK(hash, "")
}

func (app *GfcollectionApplication) Query(reqQuery types.RequestQuery) (resQuery types.ResponseQuery) {
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

func (app *GfcollectionApplication) Check(tx []byte) error {
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

func (app *GfcollectionApplication) Exec(tx []byte) ([]byte, error) {
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

func (app *GfcollectionApplication) doRequest(req *Request, msgType MessageType) (*Response, error) {
	var resp *Response
	var err error
	switch msgType {
	case MessageType_MsgInitPlatform:
		initPlatform := req.Value.(*Request_InitPlatform).InitPlatform
		resp, err = app.initPlatform(req.GetPubkey(), initPlatform.PlatformKey, initPlatform.Info, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgSetBank:
		setBank := req.Value.(*Request_SetBank).SetBank
		resp, err = app.setBank(req.GetPubkey(), setBank.GetBankId(), setBank.GetBankName(), setBank.GetPubkey(), req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgSetCompany:
		setCompany := req.Value.(*Request_SetCompany).SetCompany
		resp, err = app.setCompany(req.GetPubkey(), setCompany.GetCompanyId(), setCompany.GetCompanyName(), setCompany.GetCompanyPubkey(), req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgAddCompany:
		addCompany := req.Value.(*Request_AddCompany).AddCompany
		resp, err = app.addCompany(req.GetPubkey(), req.GetUid(), addCompany.GetCompanyId(), addCompany.GetCompanyName(), addCompany.GetCompanyArea(), addCompany.GetCompanyPubkey(), addCompany.GetWeight(), req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgAddCaseConf:
		addCaseConf := req.Value.(*Request_AddCaseConf).AddCaseConf
		resp, err = app.addCaseConf(req.GetPubkey(), req.GetUid(), addCaseConf.GetCaseConfId(), addCaseConf.GetCaseMinAmount(), addCaseConf.GetCaseMaxAmount(), addCaseConf.GetExpireDays(), addCaseConf.GetOverdueDays(), addCaseConf.GetRate(), req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgEditCaseConf:
		editCaseConf := req.Value.(*Request_EditCaseConf).EditCaseConf
		resp, err = app.editCaseConf(req.GetPubkey(), req.GetUid(), editCaseConf.GetCaseConfId(), editCaseConf.GetCaseMinAmount(), editCaseConf.GetCaseMaxAmount(), editCaseConf.GetExpireDays(), editCaseConf.GetOverdueDays(), editCaseConf.GetRate(), req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgDelCaseConf:
		delCaseConf := req.Value.(*Request_DelCaseConf).DelCaseConf
		resp, err = app.delCaseConf(req.GetPubkey(), req.GetUid(), delCaseConf.GetCaseConfId(), req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgImportCase:
		importCase := req.Value.(*Request_ImportCase).ImportCase
		resp, err = app.importCase(req.GetPubkey(), req.GetUid(), importCase.GetBankCard(), importCase.GetCaseArea(), importCase.GetCaseId(), importCase.GetCaseIdCard(), importCase.GetCaseOwner(), importCase.GetContract(), importCase.GetDebtAmount(), importCase.GetFees(), importCase.GetOriginalAmount(), importCase.GetOverdueDays(), req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgImportCaseList:
		importCaseList := req.Value.(*Request_ImportCaseList).ImportCaseList
		resp, err = app.importCaseList(req.GetPubkey(), req.GetUid(), importCaseList.GetCaseList(), req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgApplyCaseConf:
		applyCaseConf := req.Value.(*Request_ApplyCaseConf).ApplyCaseConf
		resp, err = app.applyCaseConf(req.GetPubkey(), req.GetUid(), applyCaseConf.GetCaseIds(), applyCaseConf.GetIsApply(), req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgAddCompanyConf:
		addCompanyConf := req.Value.(*Request_AddCompanyConf).AddCompanyConf
		resp, err = app.addCompanyConf(req.GetPubkey(), req.GetUid(), addCompanyConf.GetCompanyConfId(), addCompanyConf.GetCompanyConfName(), addCompanyConf.GetIsAutoAdd(), addCompanyConf.GetMaxAmount(), addCompanyConf.GetMaxReceive(), addCompanyConf.GetMinAmount(), addCompanyConf.GetOverdueDays(), addCompanyConf.GetRate(), req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgEditCompanyConf:
		editCompanyConf := req.Value.(*Request_EditCompanyConf).EditCompanyConf
		resp, err = app.editCompanyConf(req.GetPubkey(), req.GetUid(), editCompanyConf.GetCompanyConfId(), editCompanyConf.GetCompanyConfName(), editCompanyConf.GetIsAutoAdd(), editCompanyConf.GetMaxAmount(), editCompanyConf.GetMaxReceive(), editCompanyConf.GetMinAmount(), editCompanyConf.GetOverdueDays(), editCompanyConf.GetRate(), req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgDelCompanyConf:
		delCompanyConf := req.Value.(*Request_DelCompanyConf).DelCompanyConf
		resp, err = app.delCompanyConf(req.GetPubkey(), req.GetUid(), delCompanyConf.GetCompanyConfId(), req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgDelayCaseList:
		delayCaseList := req.Value.(*Request_DelayCaseList).DelayCaseList
		resp, err = app.delayCaseList(req.GetPubkey(), req.GetUid(), delayCaseList.GetCaseDelay(), req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgCancelCaseList:
		cancelCaseList := req.Value.(*Request_CancelCaseList).CancelCaseList
		resp, err = app.cancelCaseList(req.GetPubkey(), req.GetUid(), cancelCaseList.GetCaseId(), req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgFinishCaseList:
		finishCaseList := req.Value.(*Request_FinishCaseList).FinishCaseList
		resp, err = app.finishCaseList(req.GetPubkey(), req.GetUid(), finishCaseList.GetCaseId(), req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgCollectCaseList:
		collectCaseList := req.Value.(*Request_CollectCaseList).CollectCaseList
		resp, err = app.collectCaseList(req.GetPubkey(), req.GetUid(), collectCaseList.GetCaseId(), req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgSwitchCase:
		switchCase := req.Value.(*Request_SwitchCase).SwitchCase
		resp, err = app.switchCase(req.GetPubkey(), req.GetUid(), switchCase.GetCaseId(), switchCase.GetCompanyId(), req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgUpdateWeight:
		updateWeight := req.Value.(*Request_UpdateWeight).UpdateWeight
		resp, err = app.updateWeight(req.GetPubkey(), req.GetUid(), updateWeight.GetWeightList(), req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgDeliverCaseList:
		deliverCaseList := req.Value.(*Request_DeliverCaseList).DeliverCaseList
		resp, err = app.deliverCaseList(req.GetPubkey(), req.GetUid(), deliverCaseList.GetCaseIds(), req.GetInstructionId())
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

func (app *GfcollectionApplication) doCheck(req *Request, msgType MessageType) error {
	var err error
	switch msgType {
	case MessageType_MsgInitPlatform:
		initPlatform := req.Value.(*Request_InitPlatform).InitPlatform
		err = app.checkInitPlatform(req.GetPubkey(), initPlatform.PlatformKey)
	case MessageType_MsgSetBank:
		setBank := req.Value.(*Request_SetBank).SetBank
		err = app.checkSetBank(req.GetPubkey(), setBank.GetBankId(), setBank.GetPubkey())
	case MessageType_MsgSetCompany:
		setCompany := req.Value.(*Request_SetCompany).SetCompany
		err = app.checkSetCompany(req.GetPubkey(), setCompany.GetCompanyId(), setCompany.GetCompanyName(), setCompany.GetCompanyPubkey())
	case MessageType_MsgAddCompany:
		addCompany := req.Value.(*Request_AddCompany).AddCompany
		err = app.checkAddCompany(req.GetPubkey(), req.GetUid(), addCompany.GetCompanyId(), addCompany.GetCompanyPubkey(), addCompany.GetWeight())
	case MessageType_MsgAddCaseConf:
		addCaseConf := req.Value.(*Request_AddCaseConf).AddCaseConf
		err = app.checkAddCaseConf(req.GetPubkey(), req.GetUid(), addCaseConf.GetCaseConfId())
	case MessageType_MsgEditCaseConf:
		editCaseConf := req.Value.(*Request_EditCaseConf).EditCaseConf
		err = app.checkEditCaseConf(req.GetPubkey(), req.GetUid(), editCaseConf.GetCaseConfId())
	case MessageType_MsgDelCaseConf:
		delCaseConf := req.Value.(*Request_DelCaseConf).DelCaseConf
		err = app.checkDelCaseConf(req.GetPubkey(), req.GetUid(), delCaseConf.GetCaseConfId())
	case MessageType_MsgApplyCaseConf:
		applyCaseConf := req.Value.(*Request_ApplyCaseConf).ApplyCaseConf
		err = app.checkApplyCaseConf(req.GetPubkey(), req.GetUid(), applyCaseConf.GetCaseIds())
	case MessageType_MsgImportCase:
		importCase := req.Value.(*Request_ImportCase).ImportCase
		err = app.checkImportCase(req.GetPubkey(), req.GetUid(), importCase.GetCaseId())
	case MessageType_MsgImportCaseList:
		importCaseList := req.Value.(*Request_ImportCaseList).ImportCaseList
		err = app.checkImportCaseList(req.GetPubkey(), req.GetUid(), importCaseList.GetCaseList())
	case MessageType_MsgAddCompanyConf:
		addCompanyConf := req.Value.(*Request_AddCompanyConf).AddCompanyConf
		err = app.checkAddCompanyConf(req.GetPubkey(), req.GetUid(), addCompanyConf.GetCompanyConfId())
	case MessageType_MsgEditCompanyConf:
		editCompanyConf := req.Value.(*Request_EditCompanyConf).EditCompanyConf
		err = app.checkEditCompanyConf(req.GetPubkey(), req.GetUid(), editCompanyConf.GetCompanyConfId())
	case MessageType_MsgDelCompanyConf:
		delCompanyConf := req.Value.(*Request_DelCompanyConf).DelCompanyConf
		err = app.checkDelCompanyConf(req.GetPubkey(), req.GetUid(), delCompanyConf.GetCompanyConfId())
	case MessageType_MsgDelayCaseList:
		delayCaseList := req.Value.(*Request_DelayCaseList).DelayCaseList
		err = app.checkDelayCaseList(req.GetPubkey(), req.GetUid(), delayCaseList.GetCaseDelay())
	case MessageType_MsgCancelCaseList:
		cancelCaseList := req.Value.(*Request_CancelCaseList).CancelCaseList
		err = app.checkCancelCaseList(req.GetPubkey(), req.GetUid(), cancelCaseList.GetCaseId())
	case MessageType_MsgFinishCaseList:
		finishCaseList := req.Value.(*Request_FinishCaseList).FinishCaseList
		err = app.checkFinishCaseList(req.GetPubkey(), req.GetUid(), finishCaseList.GetCaseId())
	case MessageType_MsgCollectCaseList:
		collectCaseList := req.Value.(*Request_CollectCaseList).CollectCaseList
		err = app.checkCollectCaseList(req.GetPubkey(), req.GetUid(), collectCaseList.GetCaseId())
	case MessageType_MsgSwitchCase:
		switchCase := req.Value.(*Request_SwitchCase).SwitchCase
		err = app.checkSwitchCase(req.GetPubkey(), req.GetUid(), switchCase.GetCaseId(), switchCase.GetCompanyId())
	case MessageType_MsgUpdateWeight:
		updateWeight := req.Value.(*Request_UpdateWeight).UpdateWeight
		err = app.checkUpdateWeight(req.GetPubkey(), req.GetUid(), updateWeight.GetWeightList())
	case MessageType_MsgDeliverCaseList:
		deliverCaseList := req.Value.(*Request_DeliverCaseList).DeliverCaseList
		err = app.checkDeliverCaseList(req.GetPubkey(), req.GetUid(), deliverCaseList.GetCaseIds())
	default:
		err = ErrWrongMessageType
	}
	return err
}
