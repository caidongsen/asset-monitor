package smsc

import (
	//"encoding/hex"
	//"log"

	"github.com/tendermint/abci/types"
	"github.com/tendermint/merkleeyes/iavl"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/merkle"
)

type SmscApplication struct {
	types.BaseApplication

	state merkle.Tree
}

func NewSmscApplication() *SmscApplication {
	state := iavl.NewIAVLTree(0, nil)
	return &SmscApplication{state: state}
}

func (app *SmscApplication) Info() (resInfo types.ResponseInfo) {
	return types.ResponseInfo{Data: cmn.Fmt("{\"size\":%v}", app.state.Size())}
}

// tx is either "key=value" or just arbitrary bytes
func (app *SmscApplication) DeliverTx(tx []byte) types.Result {
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

func (app *SmscApplication) CheckTx(tx []byte) types.Result {
	err := app.Check(tx)
	if err != nil {
		println("app.CheckTx:", err.Error())
		return types.NewResult(types.CodeType_InternalError, []byte(err.Error()), "")
	}
	return types.OK
	//return types.OK
}

func (app *SmscApplication) Commit() types.Result {
	hash := app.state.Hash()
	return types.NewResultOK(hash, "")
}

func (app *SmscApplication) Query(reqQuery types.RequestQuery) (resQuery types.ResponseQuery) {
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

func (app *SmscApplication) Check(tx []byte) error {
	var req Request
	err := UnmarshalMessage(tx, &req)
	if err != nil {
		//log.Println("debug check err", err, tx)
		return err
	}
	//log.Println("debug check right", err, tx)
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

func (app *SmscApplication) Exec(tx []byte) ([]byte, error) {
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

func (app *SmscApplication) doCheck(req *Request, msgType MessageType) error {
	var err error
	switch msgType {
	case MessageType_MsgSetAdmin:
		err = app.checkSetAdmin(req.GetPubkey())
	case MessageType_MsgCreateAccount:
		createAccount := req.Value.(*Request_CreateAccount).CreateAccount
		err = app.checkCreateAccount(req.GetPubkey(), createAccount.Role)
	case MessageType_MsgEditAccount:
		editAccount := req.Value.(*Request_EditAccount).EditAccount
		err = app.checkEditAccount(req.GetPubkey(), editAccount.Id, editAccount.Role)
	case MessageType_MsgDeleteAccount:
		deleteAccount := req.Value.(*Request_DeleteAccount).DeleteAccount
		err = app.checkDeleteAccount(req.GetPubkey(), deleteAccount.Role)
	case MessageType_MsgSetSupplier:
		setSupplier := req.Value.(*Request_SetSupplier).SetSupplier
		err = app.checkSetSupplier(req.GetPubkey(), setSupplier.PlannerId, setSupplier.SupplierIdAdd, setSupplier.SupplierIdDel)
	case MessageType_MsgCreateOrder:
		createOrder := req.Value.(*Request_CreateOrder).CreateOrder
		err = app.checkCreateOrder(req.GetUid(), createOrder.Supplier)
	case MessageType_MsgDelivery:
		delivery := req.Value.(*Request_Delivery).Delivery
		err = app.checkDelivery(req.GetUid(), delivery.Carrier, delivery.OrderId)
	case MessageType_MsgCarry:
		carry := req.Value.(*Request_Carry).Carry
		err = app.checkCarry(req.GetUid(), carry.OrderId)
	case MessageType_MsgCheck:
		check := req.Value.(*Request_Check).Check
		err = app.checkCheck(req.GetUid(), check.OrderId)
	default:
		err = ErrWrongMessageType
	}
	return err
}

func (app *SmscApplication) doRequest(req *Request, msgType MessageType) (*Response, error) {
	var resp *Response
	var err error
	switch msgType {
	case MessageType_MsgSetAdmin:
		setAdmin := req.Value.(*Request_SetAdmin).SetAdmin
		resp, err = app.setAdmin(req.GetPubkey(), setAdmin.Pubkey, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgCreateAccount:
		createAccount := req.Value.(*Request_CreateAccount).CreateAccount
		resp, err = app.createAccount(req.GetPubkey(), createAccount.Id, createAccount.Pubkey, createAccount.Account, createAccount.Name, createAccount.Role, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgEditAccount:
		editAccount := req.Value.(*Request_EditAccount).EditAccount
		resp, err = app.editAccount(req.GetPubkey(), editAccount.Id, editAccount.Pubkey, editAccount.Account, editAccount.Name, editAccount.Role, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgDeleteAccount:
		deleteAccount := req.Value.(*Request_DeleteAccount).DeleteAccount
		resp, err = app.deleteAccount(req.GetPubkey(), deleteAccount.Id, deleteAccount.Role, deleteAccount.Account, deleteAccount.Name, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgSetSupplier:
		setSupplier := req.Value.(*Request_SetSupplier).SetSupplier
		resp, err = app.setSupplier(req.GetPubkey(), setSupplier.PlannerId, setSupplier.SupplierIdAdd, setSupplier.SupplierIdDel, setSupplier.SupplierAccountAdd, setSupplier.SupplierAccountDel, setSupplier.SupplierNameAdd, setSupplier.SupplierNameDel, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgCreateOrder:
		createOrder := req.Value.(*Request_CreateOrder).CreateOrder
		resp, err = app.createOrder(req.GetUid(), createOrder.Supplier, createOrder.PartNum, createOrder.OrderId, createOrder.PartId, createOrder.BoxId, createOrder.RequiredDate, createOrder.PlannerAccount, createOrder.PlannerName, createOrder.SupplierAccount, createOrder.SupplierName, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgDelivery:
		delivery := req.Value.(*Request_Delivery).Delivery
		resp, err = app.delivery(req.GetUid(), delivery.Carrier, delivery.PartNum, delivery.OrderId, delivery.PartId, delivery.BoxId, delivery.DeliveryDate, delivery.SupplierAccount, delivery.SupplierName, delivery.CarrierAccount, delivery.CarrierName, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgCarry:
		carry := req.Value.(*Request_Carry).Carry
		resp, err = app.carry(req.GetUid(), carry.OrderId, carry.BoxId, carry.CarId, carry.CarryDate, carry.CarrierAccount, carry.CarrierName, carry.BoxNum, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgCheck:
		check := req.Value.(*Request_Check).Check
		resp, err = app.check(req.GetUid(), check.OrderId, check.CheckDate, check.Op, check.CheckerAccount, check.CheckerName, check.PartId, check.BoxId, check.CarId, check.BoxNum, req.GetInstructionId())
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
