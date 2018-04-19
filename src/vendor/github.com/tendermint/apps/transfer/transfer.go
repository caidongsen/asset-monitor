package transfer

import (
	"encoding/hex"
	"errors"
	"log"
	"sync"

	"github.com/tendermint/abci/types"
	"github.com/tendermint/tendermint/apps"
	godb "github.com/tendermint/tendermint/apps/kvdb"
	//"fmt"
)

var errNoFunds = errors.New("errNoFunds")
var errAmount = errors.New("errAmount")
var BTY = int32(1)

type Application struct {
	apps.SupplyBaseApplication
	db   godb.DB
	mu   sync.Mutex
	conf apps.IConfig
}

func NewApplication(db godb.DB, conf apps.IConfig) *Application {
	return &Application{apps.SupplyBaseApplication{}, db, sync.Mutex{}, conf}
}

func (app *Application) LoadAccount(id []byte, coinId int32) *apps.Account {
	acc, err := apps.LoadAccount(app.db, id, coinId)
	if err != nil {
		acc = &apps.Account{}
		acc.Account = []byte(id)
		acc.Frozen = 0
		acc.Active = 0
		acc.CoinId = coinId
		acc.Transfer = 0
	}

	return acc
}

func (app *Application) SaveAccount(a *apps.Account, b *apps.Account) error {
	log.Println("save:", hex.EncodeToString(a.Account), hex.EncodeToString(b.Account))
	return apps.SaveAccounts(app.db, []*apps.Account{a, b})
}

func (app *Application) DeliverTxs(txs [][]byte) []types.Result {
	ret := make([]types.Result, len(txs))
	for i := 0; i < len(ret); i++ {
		ret[i] = app.DeliverTx(txs[i])
	}
	return ret
}

func (app *Application) initBalance() {
	conf, err := app.conf.LoadConfig(int32(apps.App_Transfer), false)
	if err != nil {
		panic(err)
	}
	acc := app.LoadAccount(conf.GetAdmin(), BTY)
	if acc.Active == 0 {
		log.Println("height == 1, init balance of admin")
		acc.Active = 3 * 1e8 * 1e8
	} else {
		return
	}
	err = apps.SaveAccounts(app.db, []*apps.Account{acc})
	if err != nil {
		panic(err)
	}
}

func (app *Application) DeliverTx(tx []byte) types.Result {
	height := app.CurrentHeader.GetHeight()
	if height == 1 || height == 0 { //genisesblock
		app.initBalance()
	}
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
	if actionId == apps.MessageType_MsgTransfer {
		transfer := req.GetTransfer()
		//余额不够，不扣钱, 返回错误，但是如果是手续费扣费，那么扣为0
		log.Println("send to:", hex.EncodeToString(transfer.GetToAddr()))
		err = app.TransferTo(req.GetAccount(), transfer.GetToAddr(), transfer.GetCoinId(), transfer.GetAmount(), false)
		if err != nil {
			return types.NewError(types.CodeType_BaseInsufficientFunds, err.Error())
		}
		return types.OK
	}
	return types.NewError(types.CodeType_BaseInvalidInput, "err actionId")
}

func (app *Application) TransferTo(from, to []byte, coinId int32, amount int64, emptyzero bool) error {
	app.mu.Lock()
	defer app.mu.Unlock()
	balance, _ := app.balance(from, coinId)
	if balance == 0 {
		return errNoFunds
	}
	if amount <= 0 || amount > 1e16 {
		return errAmount
	}
	if balance < amount {
		if emptyzero {
			amount = balance
		} else {
			return errNoFunds
		}
	}
	//opdatabase
	return app.transferTo(from, to, coinId, amount)
}

func (app *Application) Balance(from []byte, coinId int32) (int64, int64) {
	app.mu.Lock()
	defer app.mu.Unlock()
	return app.balance(from, coinId)
}

func (app *Application) balance(from []byte, coinId int32) (int64, int64) {
	account := app.LoadAccount(from, coinId)
	return account.Active, account.Frozen
}

func (app *Application) transferTo(from, to []byte, coinId int32, amount int64) error {
	accountFrom := app.LoadAccount(from, coinId)
	accountTo := app.LoadAccount(to, coinId)
	accountFrom.Active -= amount
	accountTo.Active += amount
	err := app.SaveAccount(accountFrom, accountTo)
	if err != nil {
		panic(err)
	}
	return nil
}

func (app *Application) CheckTx(tx []byte) types.Result {
	return types.NewResultOK(nil, "")
}

func (app *Application) Check(req *apps.WriteRequest) types.Result {
	println("transfer check beg")
	actionId := req.GetMsgType()
	if actionId == apps.MessageType_MsgTransfer {
		transfer := req.GetTransfer()
		if len(transfer.GetToAddr()) != 32 {
			return types.NewError(types.CodeType_BaseInvalidInput, "to addr err")
		}
		if transfer.GetAmount() <= 0 || transfer.GetAmount() > 1e16 {
			return types.NewError(types.CodeType_BaseInvalidInput, "amount err")
		}

		signs := req.GetSigns()
		account := req.GetAccount()
		coinId := transfer.GetCoinId()
		weight, err := apps.LoadSubAccount(app.db, account, account, transfer.GetCoinId())
		if err != nil {
			if err == apps.ErrNotFound {
				if len(signs) != 1 {
					return types.NewError(types.CodeType_BaseInvalidInput, "weight signs err")
				}
				s := signs[0]
				if string(s[:32]) != string(account) {
					return types.NewError(types.CodeType_BaseInvalidInput, "weight signs cmp err")
				}
				println("check ok. no-sub account")
				return types.OK
			} else {
				return types.NewError(types.CodeType_BaseInvalidInput, "weight load account err")
			}
		}

		if weight.TransferWeight == 0 {
			if len(signs) != 1 {
				return types.NewError(types.CodeType_BaseInvalidInput, "weight signs err")
			}
			s := signs[0]
			if string(s[:32]) != string(account) {
				return types.NewError(types.CodeType_BaseInvalidInput, "weight signs cmp err")
			}
		} else {
			var t int32
			for _, s := range signs {
				if string(s[:32]) == string(account) {
					continue
				}
				w, err := apps.LoadSubAccount(app.db, account, s[:32], coinId)
				if err != nil {
					return types.NewError(types.CodeType_BaseInvalidInput, "weight sign not exit")
				}
				t += w.TransferWeight
			}

			if t < weight.TransferWeight {
				return types.NewError(types.CodeType_BaseInvalidInput, "weight too low")
			}
		}
	}
	return types.OK
}

func (app *Application) Commit() types.Result {
	return types.NewResultOK([]byte("nil"), "")
}

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
	if actionId == apps.MessageType_MsgWalletInfo {
		query := req.GetWalletinfo()
		active, frozen := app.Balance(req.Account, query.CoinId)
		resp := &apps.ResponseWalletInfo{req.Account, frozen, active, query.CoinId}
		data, err := apps.Encode(&apps.Response{Value: &apps.Response_WalletInfo{resp}})
		if err != nil {
			panic(err)
		}
		return apps.NewData(types.CodeType_OK, data)
	}
	return apps.NewError(types.CodeType_BaseInvalidInput, "err actionId")
}
