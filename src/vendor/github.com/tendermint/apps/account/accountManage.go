package account

import (
	"fmt"
	"sync"

	"github.com/tendermint/abci/types"
	"github.com/tendermint/tendermint/apps"
	godb "github.com/tendermint/tendermint/apps/kvdb"
)

type Application struct {
	types.BaseApplication
	db   godb.DB
	mu   sync.Mutex
	conf apps.IConfig
}

func NewApplication(db godb.DB, conf apps.IConfig) *Application {
	return &Application{types.BaseApplication{}, db, sync.Mutex{}, conf}
}

func (app *Application) DeliverTxs(txs [][]byte) []types.Result {
	ret := make([]types.Result, len(txs))
	for i := 0; i < len(ret); i++ {
		ret[i] = app.DeliverTx(txs[i])
	}
	return ret
}

func (app *Application) DeliverTx(tx []byte) types.Result {
	var req apps.WriteRequest
	err := apps.Decode(tx, &req)
	if err != nil {
		return types.NewError(types.CodeType_BaseInvalidInput, err.Error())
	}
	r := app.Check(&req)
	if !r.IsOK() {
		return r
	}
	actionId := req.GetMsgType()
	if actionId == apps.MessageType_MsgSetWeight {
		w := req.GetSetWeight()
		account := req.GetAccount()
		subaccount := w.GetSubaccount()
		coinId := w.GetCoinId()
		transferweight := w.GetTransferWeight()
		queryweight := w.GetQueryWeight()
		managerweight := w.GetManagerWeight()
		err = app.SetWeight(account, subaccount, coinId, transferweight, queryweight, managerweight)
		if err != nil {
			return types.NewError(types.CodeType_BaseInsufficientFunds, err.Error())
		}
		return types.OK
	} else if actionId == apps.MessageType_MsgDelWeight {
		w := req.GetDelWeight()
		account := req.GetAccount()
		subaccount := w.GetSubaccount()
		coinId := w.GetCoinId()
		err = app.DelWeight(account, subaccount, coinId)
		if err != nil {
			return types.NewError(types.CodeType_BaseInsufficientFunds, err.Error())
		}

		return types.OK
	} else if actionId == apps.MessageType_MsgSetAccount {
		w := req.GetSetAccount()
		account := w.GetAccount()
		frozen := w.GetFrozen()
		active := w.GetActive()
		coinId := w.GetCoinId()
		transfer := w.GetTransfer()
		fmt.Println("DeliverTx:67 account = ", account)
		err = app.SetAccount(account, frozen, active, coinId, transfer)
		if err != nil {
			return types.NewError(types.CodeType_BaseInsufficientFunds, err.Error())
		}
		return types.OK
	} else {
		return types.NewError(types.CodeType_BaseInvalidInput, "err actionId")
	}
}

//add weight
func (app *Application) SetWeight(account []byte, subaccount []byte, coinId int32, transfer int32, query int32, manager int32) error {
	app.mu.Lock()
	defer app.mu.Unlock()

	s, err := apps.LoadSubAccount(app.db, account, subaccount, coinId)
	if err != nil {
		if err == apps.ErrNotFound {
			s = &apps.SubAccount{
				Account:        account,
				Subaccount:     subaccount,
				CoinId:         coinId,
				TransferWeight: transfer,
				QueryWeight:    query,
				ManagerWeight:  manager,
			}
		} else {
			return err
		}
	} else {
		s.TransferWeight = transfer
		s.QueryWeight = query
		s.ManagerWeight = manager
	}

	err = apps.SaveSubAccount(app.db, s)
	if err != nil {
		panic(err)
	}

	return nil
}

//delete weight
func (app *Application) DelWeight(account []byte, subaccount []byte, coinId int32) error {
	batch := app.db.NewBatch(true)
	key := apps.KeySubAccount(account, subaccount, coinId)

	err := batch.Delete([]byte(key))
	if err != nil {
		return err
	}

	return batch.Commit()
}

func (app *Application) CheckTx(tx []byte) types.Result {
	return types.NewResultOK(nil, "")
}

func (app *Application) Check(req *apps.WriteRequest) types.Result {
	actionId := req.GetMsgType()
	if actionId == apps.MessageType_MsgSetWeight {
		weight := req.GetSetWeight()
		if len(weight.GetSubaccount()) != 32 {
			return types.NewError(types.CodeType_BaseInvalidInput, "set subaccount err")
		}
		if weight.GetTransferWeight() < 0 || weight.GetTransferWeight() > 100 {
			return types.NewError(types.CodeType_BaseInvalidInput, "transferweight round err")
		}
		if weight.GetQueryWeight() < 0 || weight.GetQueryWeight() > 100 {
			return types.NewError(types.CodeType_BaseInvalidInput, "queryweight round err")
		}
		if weight.GetManagerWeight() < 0 || weight.GetManagerWeight() > 100 {
			return types.NewError(types.CodeType_BaseInvalidInput, "managerweight round err")
		}
	} else if actionId == apps.MessageType_MsgDelWeight {
		weight := req.GetDelWeight()
		if len(weight.GetSubaccount()) != 32 {
			return types.NewError(types.CodeType_BaseInvalidInput, "del subaccount err")
		}

		if string(req.GetAccount()) == string(weight.GetSubaccount()) {
			return types.NewError(types.CodeType_BaseInvalidInput, "can not del account")
		}
	} else if actionId == apps.MessageType_MsgSetAccount {
		account := req.GetSetAccount()
		fmt.Println("Check:160 account = ", account)
		if len(account.GetAccount()) != 32 {
			return types.NewError(types.CodeType_BaseInvalidInput, "set account err")
		}
		if account.GetTransfer() < 0 {
			return types.NewError(types.CodeType_BaseInvalidInput, "transfer round err")
		}
	} else {
		return types.NewError(types.CodeType_BaseInvalidInput, "err actionId")
	}

	return types.OK
}

func (app *Application) Commit() types.Result {
	return types.NewResultOK([]byte("nil"), "")
}

//purpose for query
func (app *Application) Query(reqQuery types.RequestQuery) (resQuery types.ResponseQuery) {
	var req apps.ReadRequest
	tx := reqQuery.GetData()
	err := apps.Decode(tx, &req)
	if err != nil {
		return apps.NewError(types.CodeType_BaseInvalidInput, err.Error())
	}

	err = req.CheckSign()
	if err != nil {
		return apps.NewError(types.CodeType_BaseInvalidSignature, err.Error())
	}

	actionId := req.GetMsgType()
	if actionId == apps.MessageType_MsgWeightInfo {
		query := req.GetWeightinfo()
		s, err := apps.LoadSubAccount(app.db, req.Account, query.Subaccount, query.CoinId)
		if err != nil {
			return apps.NewError(types.CodeType_BaseInvalidInput, err.Error())
		}

		resp := &apps.ResponseWeightInfo{s.Account, s.Subaccount, s.CoinId, s.TransferWeight, s.QueryWeight, s.ManagerWeight}
		data, err := apps.Encode(&apps.Response{Value: &apps.Response_WeightInfo{resp}})
		if err != nil {
			panic(err)
		}
		return apps.NewData(types.CodeType_OK, data)
	} else if actionId == apps.MessageType_MsgAccountInfo {
		query := req.GetAccountinfo()
		fmt.Println("Query:206 req.Account = ", req.Account)
		s, err := apps.LoadAccount(app.db, query.GetAccount(), query.GetCoinId())
		if err != nil {
			fmt.Println("Query:209")
			return apps.NewError(types.CodeType_BaseInvalidInput, err.Error())
		}
		fmt.Println("Query:212")
		resp := &apps.ResponseAccountInfo{s.Account, s.Frozen, s.Active, s.CoinId, s.Transfer}
		data, err := apps.Encode(&apps.Response{Value: &apps.Response_AccountInfo{resp}})
		if err != nil {
			panic(err)
		}
		return apps.NewData(types.CodeType_OK, data)
	}
	return apps.NewError(types.CodeType_BaseInvalidInput, "err actionId")
}

//add new account
func (app *Application) SetAccount(account []byte, frozen int64, active int64, coinId int32, transfer int32) error {
	app.mu.Lock()
	defer app.mu.Unlock()

	fmt.Println("SetAccount:226")
	s, err := apps.LoadAccount(app.db, account, coinId)

	//该账户不存在，需要新建一个账户
	if err != nil {
		fmt.Println("SetAccount:231")
		if err == apps.ErrNotFound {
			s = &apps.Account{
				Account:  account,
				Frozen:   frozen,
				Active:   active,
				CoinId:   coinId,
				Transfer: transfer,
			}
			fmt.Println("SetAccount:240")
			err = apps.SaveAccount(app.db, s)
			if err != nil {
				panic(err)
			}
			fmt.Println("SetAccount:245")
			return nil
		} else {
			fmt.Println("SetAccount:248")
			return err
		}
	}
	fmt.Println("SetAccount:252")
	return nil
}