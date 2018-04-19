package gwpoints

import (
	//"encoding/hex"
	//"log"

	"github.com/tendermint/abci/types"
	"github.com/tendermint/merkleeyes/iavl"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/merkle"
)

type GwpointsApplication struct {
	types.BaseApplication

	state merkle.Tree
}

func NewGwpointsApplication() *GwpointsApplication {
	state := iavl.NewIAVLTree(0, nil)
	return &GwpointsApplication{state: state}
}

func (app *GwpointsApplication) Info() (resInfo types.ResponseInfo) {
	return types.ResponseInfo{Data: cmn.Fmt("{\"size\":%v}", app.state.Size())}
}

// tx is either "key=value" or just arbitrary bytes
func (app *GwpointsApplication) DeliverTx(tx []byte) types.Result {
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

func (app *GwpointsApplication) CheckTx(tx []byte) types.Result {
	err := app.Check(tx)
	if err != nil {
		println("app.CheckTx:", err.Error())
		return types.NewResult(types.CodeType_InternalError, []byte(err.Error()), "")
	}
	return types.OK
	//return types.OK
}

func (app *GwpointsApplication) Commit() types.Result {
	hash := app.state.Hash()
	return types.NewResultOK(hash, "")
}

func (app *GwpointsApplication) Query(reqQuery types.RequestQuery) (resQuery types.ResponseQuery) {
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

func (app *GwpointsApplication) Check(tx []byte) error {
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

func (app *GwpointsApplication) Exec(tx []byte) ([]byte, error) {
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

func (app *GwpointsApplication) doCheck(req *Request, msgType MessageType) error {
	var err error
	switch msgType {
	case MessageType_MsgInitPlatform:
		err = app.checkInitPlatform(req.GetPubkey())
	case MessageType_MsgUserCreate:
		//log.Println("debug check UserCreate", req.GetUid(), req.GetInstructionId(), req.GetActionId())
		//log.Println("debug check UserCreate", hex.EncodeToString(req.GetPubkey()), hex.EncodeToString(req.GetSign()))
		userCreate := req.Value.(*Request_UserCreate).UserCreate
		err = app.checkUserCreate(req.GetPubkey(), userCreate.UserUid)
		//log.Println("debug check UserCreate", userCreate.UserUid, hex.EncodeToString(userCreate.UserPubkey), hex.EncodeToString(userCreate.Info), err)
	case MessageType_MsgChangePubkey:
		changePubkey := req.Value.(*Request_ChangePubkey).ChangePubkey
		err = app.checkChangePubkey(req.GetUid(), req.GetPubkey(), changePubkey.NewPubkey)
	case MessageType_MsgDistributeGwpoints:
		distributeGwpoints := req.Value.(*Request_DistributeGwpoints).DistributeGwpoints
		err = app.checkDistributeGwpoints(req.GetPubkey(), distributeGwpoints.GwPoints)
	case MessageType_MsgSetCompanyExchangeRate:
		err = app.checkSetCompanyExchangeRate(req.GetPubkey())
	case MessageType_MsgBuyGwpoints:
		//log.Println("debug check BuyGwpoints", req.GetUid(), req.GetInstructionId(), req.GetActionId())
		//log.Println("debug check BuyGwpoints", hex.EncodeToString(req.GetPubkey()), hex.EncodeToString(req.GetSign()))
		err = app.checkBuyGwpoints(req.GetPubkey(), req.GetUid())
	case MessageType_MsgSellGwPoints:
		err = app.checkSellGwpoints(req.GetPubkey(), req.GetUid())
	case MessageType_MsgBuyGoods:
		err = app.checkBuyGoods(req.GetPubkey(), req.GetUid())
	case MessageType_MsgClear:
		err = app.checkClear(req.GetPubkey())
	case MessageType_MsgSyncPoints:
		syncPoints := req.Value.(*Request_SyncPoints).SyncPoints
		err = app.checkSyncPoints(req.GetPubkey(), syncPoints.GwPoints)
	default:
		err = ErrWrongMessageType
	}
	return err
}

func (app *GwpointsApplication) doRequest(req *Request, msgType MessageType) (*Response, error) {
	var resp *Response
	var err error
	switch msgType {
	case MessageType_MsgInitPlatform:
		initPlatform := req.Value.(*Request_InitPlatform).InitPlatform
		resp, err = app.initPlatform(req.GetPubkey(), initPlatform.Pubkey, initPlatform.Info, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgUserCreate:
		//log.Println("debug UserCreate", req.GetUid(), req.GetInstructionId(), req.GetActionId())
		//log.Println("debug UserCreate", hex.EncodeToString(req.GetPubkey()), hex.EncodeToString(req.GetSign()))
		userCreate := req.Value.(*Request_UserCreate).UserCreate
		resp, err = app.userCreate(req.GetPubkey(), userCreate.UserUid, userCreate.UserPubkey, userCreate.Info, req.GetInstructionId())
		//log.Println("debug UserCreate", userCreate.UserUid, hex.EncodeToString(userCreate.UserPubkey), hex.EncodeToString(userCreate.Info), err)
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
	case MessageType_MsgDistributeGwpoints:
		distributeGwpoints := req.Value.(*Request_DistributeGwpoints).DistributeGwpoints
		resp, err = app.distributeGwpoints(req.GetPubkey(), distributeGwpoints.UserUid, distributeGwpoints.UserPubkey, distributeGwpoints.GwPoints, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgSetCompanyExchangeRate:
		setCompanyExchangeRate := req.Value.(*Request_SetCompanyExchangeRate).SetCompanyExchangeRate
		resp, err = app.setCompanyExchangeRate(req.GetPubkey(), setCompanyExchangeRate.Id, setCompanyExchangeRate.CompanyNum, setCompanyExchangeRate.GwNum, setCompanyExchangeRate.Info, req.GetInstructionId())
		//log.Println("debug SetCompanyExchangeRate", req.GetUid(), req.GetInstructionId(), req.GetActionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgBuyGwpoints:
		//log.Println("debug BuyGwpoints", req.GetUid(), req.GetInstructionId(), req.GetActionId())
		//log.Println("debug BuyGwpoints", hex.EncodeToString(req.GetPubkey()), hex.EncodeToString(req.GetSign()))
		buyGwpoints := req.Value.(*Request_BuyGwpoints).BuyGwpoints
		resp, err = app.buyGwpoints(req.GetPubkey(), req.GetUid(), buyGwpoints.CompanyId, buyGwpoints.CompanyPoints, req.GetInstructionId())
		//log.Println("debug BuyGwpoints", buyGwpoints.CompanyId, buyGwpoints.CompanyPoints, err)
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgSellGwPoints:
		sellGwpoints := req.Value.(*Request_SellGwPoints).SellGwPoints
		resp, err = app.sellGwpoints(req.GetPubkey(), req.GetUid(), sellGwpoints.CompanyId, sellGwpoints.CompanyPoints, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgBuyGoods:
		buyGoods := req.Value.(*Request_BuyGoods).BuyGoods
		resp, err = app.buyGoods(req.GetPubkey(), req.GetUid(), buyGoods.GwPoints, buyGoods.OrderId, buyGoods.ProductName, buyGoods.ProductNum, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgClear:
		clear := req.Value.(*Request_Clear).Clear
		resp, err = app.clear(req.GetPubkey(), clear.CompanyId, req.GetInstructionId())
		receipt := &Receipt{}
		if err != nil {
			receipt.Err = []byte(err.Error())
		}
		receipt.IsOk = (err == nil)
		app.saveReceipt(req.GetInstructionId(), receipt)
	case MessageType_MsgSyncPoints:
		syncPoints := req.Value.(*Request_SyncPoints).SyncPoints
		resp, err = app.syncPoints(req.GetPubkey(), syncPoints.UserUid, syncPoints.UserPubkey, syncPoints.GwPoints, syncPoints.Added, req.GetInstructionId())
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
