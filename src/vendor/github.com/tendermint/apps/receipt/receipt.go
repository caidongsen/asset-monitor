package receipt

import (
	"github.com/tendermint/abci/types"
	"github.com/tendermint/tendermint/apps"
	godb "github.com/tendermint/tendermint/apps/kvdb"
)

type Application struct {
	types.BaseApplication
	db   godb.DB
	conf apps.IConfig
}

func NewApplication(db godb.DB, conf apps.IConfig) *Application {
	return &Application{types.BaseApplication{}, db, conf}
}

func (app *Application) DeliverTxs(txs [][]byte) []types.Result {
	ret := make([]types.Result, len(txs))
	for i := 0; i < len(ret); i++ {
		ret[i] = app.DeliverTx(txs[i])
	}
	return ret
}

func (app *Application) save(receipt *apps.RequestReceiptCreate) error {
	appconf, err := app.conf.LoadConfig(int32(apps.App_Receipt), false)
	if err != nil {
		panic(err)
	}
	batch := app.db.NewBatch(false)
	for i := 0; i < len(receipt.GetReceipts()); i++ {
		item := receipt.Receipts[i]
		var rec apps.ResponseReceipt
		rec.TxHash = item.GetTxhash()
		rec.AppId = appconf.GetAppId()
		rec.AppVersion = appconf.GetAppVersion()
		rec.IsOk = item.GetIsok()
		rec.Result = item.GetResult()
		key := apps.KeyReceipt(rec.GetAppVersion(), rec.GetTxHash())
		value, err := apps.Encode(&rec)
		if err != nil {
			panic(err)
		}
		err = batch.Put([]byte(key), value)
		if err != nil {
			panic(err)
		}
	}
	err = batch.Commit()
	if err != nil {
		panic(err)
	}
	return nil
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
	if actionId == apps.MessageType_MsgReceiptCreate {
		receipt := req.GetReceiptCreate()
		//余额不够，不扣钱, 返回错误，但是如果是手续费扣费，那么扣为0
		err = app.save(receipt)
		if err != nil {
			return types.NewError(types.CodeType_BaseInvalidInput, err.Error())
		}
		return types.OK
	}
	return types.NewError(types.CodeType_BaseInvalidInput, "err actionId")
}

func (app *Application) Receipt(hash []byte, version int32) (*apps.ResponseReceipt, error) {
	key := apps.KeyReceipt(version, hash)
	data := app.db.Get([]byte(key))
	if data == nil {
		return nil, apps.ErrNotFound
	}
	var receipt apps.ResponseReceipt
	err := apps.Decode(data, &receipt)
	if err != nil {
		panic(err)
	}
	return &receipt, nil
}

func (app *Application) CheckTx(tx []byte) types.Result {
	return types.NewResultOK(nil, "")
}

func (app *Application) Check(req *apps.WriteRequest) types.Result {
	actionId := req.GetMsgType()
	if actionId == apps.MessageType_MsgReceiptCreate {
		receipt := req.GetReceiptCreate()
		for i := 0; i < len(receipt.GetReceipts()); i++ {
			item := receipt.Receipts[i]
			key := apps.KeyTx(item.GetTxhash())
			reftx := app.db.Get([]byte(key))
			if reftx == nil {
				return types.NewError(types.CodeType_BaseInvalidInput, "receipt:notx")
			}
			rtx, err := apps.ParseTx(reftx)
			if err != nil {
				return types.NewError(types.CodeType_BaseInvalidInput, err.Error())
			}
			if rtx.GetAppId() != receipt.AppId {
				return types.NewError(types.CodeType_BaseInvalidInput, "receipt:appid not same")
			}
			conf, err := app.conf.LoadConfig(rtx.GetAppId(), false)
			if err != nil {
				return types.NewError(types.CodeType_BaseInvalidInput, err.Error())
			}
			if conf.GetAppVersion() != receipt.AppVersion {
				return types.NewError(types.CodeType_BaseInvalidInput, "receipt:appversion not same")
			}
			if !conf.IsSender(req.GetAccount()) {
				return types.NewError(types.CodeType_BaseInvalidInput, "receipt:no permission")
			}
			keyreceipt := apps.KeyReceipt(conf.GetAppVersion(), item.GetTxhash())
			if app.db.Get([]byte(keyreceipt)) != nil {
				return types.NewError(types.CodeType_BaseInvalidInput, "receipt:has set")
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
	query := reqQuery.GetData()
	err := apps.Decode(query, &req)
	if err != nil {
		return apps.NewError(types.CodeType_BaseInvalidInput, err.Error())
	}
	err = req.CheckSign()
	if err != nil {
		return apps.NewError(types.CodeType_BaseInvalidSignature, err.Error())
	}
	actionId := req.GetMsgType()
	if actionId == apps.MessageType_MsgReceipt {
		query := req.GetReceipt()
		receipt, err := app.Receipt(query.GetTxhash(), query.GetAppVersion())
		if err != nil {
			return apps.NewError(types.CodeType_BaseInvalidInput, err.Error())
		}
		data, err := apps.Encode(&apps.Response{Value: &apps.Response_Receipt{receipt}})
		if err != nil {
			panic(err)
		}
		return apps.NewData(types.CodeType_OK, data)
	}
	return apps.NewError(types.CodeType_BaseInvalidInput, "err actionId")
}
