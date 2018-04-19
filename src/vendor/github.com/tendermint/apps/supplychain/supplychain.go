package supplychain

import (
	"github.com/tendermint/abci/types"
	"github.com/tendermint/merkleeyes/iavl"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/merkle"
)

type SupplychainApplication struct {
	types.BaseApplication

	state merkle.Tree
}

func NewSupplychainApplication() *SupplychainApplication {
	state := iavl.NewIAVLTree(0, nil)
	return &SupplychainApplication{state: state}
}

func (app *SupplychainApplication) Info() (resInfo types.ResponseInfo) {
	return types.ResponseInfo{Data: cmn.Fmt("{\"size\":%v}", app.state.Size())}
}

// tx is either "key=value" or just arbitrary bytes
func (app *SupplychainApplication) DeliverTx(tx []byte) types.Result {
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

func (app *SupplychainApplication) CheckTx(tx []byte) types.Result {
	err := app.Check(tx)
	if err != nil {
		println("app.CheckTx:", err.Error())
		return types.NewResult(types.CodeType_InternalError, []byte(err.Error()), "")
	}
	return types.OK
	//return types.OK
}

func (app *SupplychainApplication) Commit() types.Result {
	hash := app.state.Hash()
	return types.NewResultOK(hash, "")
}

func (app *SupplychainApplication) Query(reqQuery types.RequestQuery) (resQuery types.ResponseQuery) {
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

func (app *SupplychainApplication) Check(tx []byte) error {
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

func (app *SupplychainApplication) Exec(tx []byte) ([]byte, error) {
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

func (app *SupplychainApplication) doCheck(req *Request, msgType MessageType) error {
	var err error
	switch msgType {
	case MessageType_MsgUserCreate:
		userCreate := req.Value.(*Request_UserCreate).UserCreate
		err = app.checkUserCreate(req.GetPubkey(), userCreate.Uid)
	case MessageType_MsgUserPaid:
		userPaid := req.Value.(*Request_UserPaid).UserPaid
		err = app.checkUserPaid(req.GetPubkey(), userPaid.Uid, userPaid.Pubkey)
	case MessageType_MsgChangePubkey:
		changePubkey := req.Value.(*Request_ChangePubkey).ChangePubkey
		err = app.checkChangePubkey(req.GetPubkey(), req.GetUid(), changePubkey.NewPubkey)
	case MessageType_MsgChangeCredit:
		changeCredit := req.Value.(*Request_ChangeCredit).ChangeCredit
		err = app.checkChangeCredit(req.GetPubkey(), changeCredit.Uid, changeCredit.Credit)
	case MessageType_MsgCreateLoan:
		createLoan := req.Value.(*Request_CreateLoan).CreateLoan
		err = app.checkCreateLoan(req.GetPubkey(), req.GetUid(), createLoan.LoanId, createLoan.Amount)
	case MessageType_MsgApplyCredit:
		applyCredit := req.Value.(*Request_ApplyCredit).ApplyCredit
		err = app.checkApplyCredit(req.GetPubkey(), req.GetUid(), applyCredit.LoanId)
	case MessageType_MsgApplyGuarantee:
		applyGuarantee := req.Value.(*Request_ApplyGuarantee).ApplyGuarantee
		err = app.checkApplyGuarantee(req.GetPubkey(), req.GetUid(), applyGuarantee.GuaranteeUid, applyGuarantee.LoanId)
	case MessageType_MsgGuaranteeFeedback:
		guaranteeFeedback := req.Value.(*Request_GuaranteeFeedback).GuaranteeFeedback
		err = app.checkGuaranteeFeedback(req.GetPubkey(), req.GetUid(), guaranteeFeedback.LoanId)
	case MessageType_MsgCreditFeedback:
		creditFeedback := req.Value.(*Request_CreditFeedback).CreditFeedback
		err = app.checkCreditFeedback(req.GetPubkey(), creditFeedback.LoanId)
	case MessageType_MsgEditCredit:
		editCredit := req.Value.(*Request_EditCredit).EditCredit
		err = app.checkEditCredit(req.GetPubkey(), req.GetUid(), editCredit.LoanId)
	case MessageType_MsgIssueLoan:
		issueLoan := req.Value.(*Request_IssueLoan).IssueLoan
		err = app.checkIssueLoan(req.GetPubkey(), req.GetUid(), issueLoan.LoanId)
	case MessageType_MsgCancelLoan:
		cancelLoan := req.Value.(*Request_CancelLoan).CancelLoan
		err = app.checkCancelLoan(req.GetPubkey(), req.GetUid(), cancelLoan.LoanId)
	case MessageType_MsgPrepareBuy:
		prepareBuy := req.Value.(*Request_PrepareBuy).PrepareBuy
		err = app.checkPrepareBuy(req.GetPubkey(), req.GetUid(), prepareBuy.LoanId)
	case MessageType_MsgPay:
		pay := req.Value.(*Request_Pay).Pay
		err = app.checkPay(req.GetPubkey(), req.GetUid(), pay.LoanId)
	case MessageType_MsgConfirmReceive:
		confirmReceive := req.Value.(*Request_ConfirmReceive).ConfirmReceive
		err = app.checkConfirmReceive(req.GetPubkey(), req.GetUid(), confirmReceive.LoanId)
	case MessageType_MsgRepayAdvance:
		repayAdvance := req.Value.(*Request_RepayAdvance).RepayAdvance
		err = app.checkRepayAdvance(req.GetPubkey(), req.GetUid(), repayAdvance.LoanId, repayAdvance.Amount)
	case MessageType_MsgRepay:
		repay := req.Value.(*Request_Repay).Repay
		err = app.checkRepay(req.GetPubkey(), req.GetUid(), repay.LoanId, repay.Amount)
	case MessageType_MsgConfirmRepay:
		confirmRepay := req.Value.(*Request_ConfirmRepay).ConfirmRepay
		err = app.checkConfirmRepay(req.GetPubkey(), req.GetUid(), confirmRepay.LoanId)
	case MessageType_MsgCreateAdmin:
		createAdmin := req.Value.(*Request_CreateAdmin).CreateAdmin
		err = app.checkCreateAdmin(req.GetPubkey(), createAdmin.Pubkey)
	case MessageType_MsgSetGuarantee:
		setGuarantee := req.Value.(*Request_SetGuarantee).SetGuarantee
		err = app.checkSetGuarantee(req.GetPubkey(), setGuarantee.Uid, setGuarantee.Guarantee)
	case MessageType_MsgDeposit:
		deposit := req.Value.(*Request_Deposit).Deposit
		err = app.checkDeposit(req.GetPubkey(), deposit.UserUid, deposit.UserPubkey, deposit.Cash)
	case MessageType_MsgWithdraw:
		withdraw := req.Value.(*Request_Withdraw).Withdraw
		err = app.checkWithdraw(req.GetPubkey(), req.GetUid(), withdraw.Bank, withdraw.Cash)
	case MessageType_MsgIncreaseBankRmb:
		increaseBankRmb := req.Value.(*Request_IncreaseBankRmb).IncreaseBankRmb
		err = app.checkIncreaseBankRmb(req.GetPubkey(), increaseBankRmb.Bank, increaseBankRmb.Cash)
	default:
		err = ErrWrongMessageType
	}
	return err
}

func (app *SupplychainApplication) doRequest(req *Request, msgType MessageType) (*Response, error) {
	var resp *Response
	var err error
	switch msgType {
	case MessageType_MsgUserCreate:
		userCreate := req.Value.(*Request_UserCreate).UserCreate
		resp, err = app.userCreate(req.GetPubkey(), userCreate.Uid, userCreate.Pubkey, userCreate.Type, userCreate.Credit, userCreate.Info, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgUserPaid:
		userPaid := req.Value.(*Request_UserPaid).UserPaid
		resp, err = app.userPaid(req.GetPubkey(), userPaid.Uid, userPaid.Pubkey, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgChangePubkey:
		changePubkey := req.Value.(*Request_ChangePubkey).ChangePubkey
		resp, err = app.changePubkey(req.GetPubkey(), req.GetUid(), changePubkey.NewPubkey, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgChangeCredit:
		changeCredit := req.Value.(*Request_ChangeCredit).ChangeCredit
		resp, err = app.changeCredit(req.GetPubkey(), changeCredit.Uid, changeCredit.Credit, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgCreateLoan:
		createLoan := req.Value.(*Request_CreateLoan).CreateLoan
		resp, err = app.createLoan(req.GetPubkey(), req.GetUid(), createLoan.LoanId, createLoan.Amount, createLoan.Rate, createLoan.Expiration, createLoan.Info, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgApplyCredit:
		applyCredit := req.Value.(*Request_ApplyCredit).ApplyCredit
		resp, err = app.applyCredit(req.GetPubkey(), req.GetUid(), applyCredit.LoanId, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgApplyGuarantee:
		applyGuarantee := req.Value.(*Request_ApplyGuarantee).ApplyGuarantee
		resp, err = app.applyGuarantee(req.GetPubkey(), req.GetUid(), applyGuarantee.GuaranteeUid, applyGuarantee.LoanId, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgGuaranteeFeedback:
		guaranteeFeedback := req.Value.(*Request_GuaranteeFeedback).GuaranteeFeedback
		resp, err = app.guaranteeFeedback(req.GetPubkey(), req.GetUid(), guaranteeFeedback.LoanId, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgCreditFeedback:
		creditFeedback := req.Value.(*Request_CreditFeedback).CreditFeedback
		resp, err = app.creditFeedback(req.GetPubkey(), creditFeedback.LoanId, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgEditCredit:
		editCredit := req.Value.(*Request_EditCredit).EditCredit
		resp, err = app.editCredit(req.GetPubkey(), req.GetUid(), editCredit.LoanId, editCredit.Info, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgIssueLoan:
		issueLoan := req.Value.(*Request_IssueLoan).IssueLoan
		resp, err = app.issueLoan(req.GetPubkey(), req.GetUid(), issueLoan.LoanId, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgCancelLoan:
		cancelLoan := req.Value.(*Request_CancelLoan).CancelLoan
		resp, err = app.cancelLoan(req.GetPubkey(), req.GetUid(), cancelLoan.LoanId, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgPrepareBuy:
		prepareBuy := req.Value.(*Request_PrepareBuy).PrepareBuy
		resp, err = app.prepareBuy(req.GetPubkey(), req.GetUid(), prepareBuy.LoanId, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgPay:
		pay := req.Value.(*Request_Pay).Pay
		resp, err = app.pay(req.GetPubkey(), req.GetUid(), pay.LoanId, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgConfirmReceive:
		confirmReceive := req.Value.(*Request_ConfirmReceive).ConfirmReceive
		resp, err = app.confirmReceive(req.GetPubkey(), req.GetUid(), confirmReceive.LoanId, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgRepayAdvance:
		repayAdvance := req.Value.(*Request_RepayAdvance).RepayAdvance
		resp, err = app.repayAdvance(req.GetPubkey(), req.GetUid(), repayAdvance.LoanId, repayAdvance.Amount, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgRepay:
		repay := req.Value.(*Request_Repay).Repay
		resp, err = app.repay(req.GetPubkey(), req.GetUid(), repay.LoanId, repay.Amount, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgConfirmRepay:
		confirmRepay := req.Value.(*Request_ConfirmRepay).ConfirmRepay
		resp, err = app.confirmRepay(req.GetPubkey(), req.GetUid(), confirmRepay.LoanId, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgCreateAdmin:
		createAdmin := req.Value.(*Request_CreateAdmin).CreateAdmin
		resp, err = app.createAdmin(req.GetPubkey(), createAdmin.Pubkey, createAdmin.Type, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgSetGuarantee:
		setGuarantee := req.Value.(*Request_SetGuarantee).SetGuarantee
		resp, err = app.setGuarantee(req.GetPubkey(), setGuarantee.Uid, setGuarantee.Guarantee, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgDeposit:
		deposit := req.Value.(*Request_Deposit).Deposit
		resp, err = app.deposit(req.GetPubkey(), deposit.UserUid, deposit.UserPubkey, deposit.Cash, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgWithdraw:
		withdraw := req.Value.(*Request_Withdraw).Withdraw
		resp, err = app.withdraw(req.GetPubkey(), req.GetUid(), withdraw.Bank, withdraw.Cash, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgIncreaseBankRmb:
		increaseBankRmb := req.Value.(*Request_IncreaseBankRmb).IncreaseBankRmb
		resp, err = app.increaseBankRmb(req.GetPubkey(), increaseBankRmb.Bank, increaseBankRmb.Cash, req.GetInstructionId())
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
