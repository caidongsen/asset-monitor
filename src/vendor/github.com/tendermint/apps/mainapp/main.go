package mainapp

import (
	"github.com/tendermint/abci/types"
	"github.com/tendermint/tendermint/apps"
	"github.com/tendermint/tendermint/apps/account"
	godb "github.com/tendermint/tendermint/apps/kvdb"
	"github.com/tendermint/tendermint/apps/receipt"
	"github.com/tendermint/tendermint/apps/reg"
	"github.com/tendermint/tendermint/apps/transfer"
	//cfg "github.com/tendermint/tendermint/config"
	//"github.com/tendermint/tmlibs/merkle"
)

type MainApplication struct {
	apps.SupplyBaseApplication
	apps      []Application
	signs     []Signer
	db        godb.DB
	unitfee   int
	mempoolTx map[string]bool
	hash      []byte
}

type Application interface {
	types.Application
	Check(req *apps.WriteRequest) types.Result
}

type Signer interface {
	GetLen() int
	GetPubKey(sign []byte) ([]byte, error)
	Verify(data []byte, sign []byte) bool
}

func NewApplication(tmroot string) *MainApplication {
	db := godb.NewDB("app", godb.DBBackendLevelDB, tmroot)
	main := &MainApplication{}
	main.db = db
	//注册app,暂时内部app大小为255
	main.apps = make([]Application, apps.App_Max)
	main.signs = make([]Signer, apps.Sign_MaxSign)

	//初始化reg app
	//这个app 负责负责app的注册和版本的管理
	//这个是最基础的app,第二个参数是加载配置的最后一个appId
	main.apps[apps.App_Reg] = reg.NewApplication(db, int32(apps.App_Last))
	main.mempoolTx = make(map[string]bool)

	main.apps[apps.App_Transfer] = transfer.NewApplication(db, main)
	main.apps[apps.App_Receipt] = receipt.NewApplication(db, main)
	main.apps[apps.App_AccountManage] = account.NewApplication(db, main)
	main.signs[apps.Sign_Ed25519] = newEd25519()

	main.unitfee = int(apps.Coin_PAN)
	return main
}

func (app *MainApplication) LoadConfig(appId int32, update bool) (*apps.SystemAccount, error) {
	return app.apps[apps.App_Reg].(*reg.Application).LoadConfig(appId, false)
}

func (app *MainApplication) loadconfig(appId int32) (*apps.SystemAccount, error) {
	return app.LoadConfig(appId, false)
}

func (app *MainApplication) hasTx(key string) bool {
	return app.db.Get([]byte(key)) != nil
}

func newResultHash(hash []byte) []byte {
	resp := &apps.ResponseHash{hash}
	data, err := apps.Encode(&apps.Response{Value: &apps.Response_Hash{resp}})
	if err != nil {
		panic(err)
	}
	return data
}

func (app *MainApplication) DeliverTx(tx []byte) (ret types.Result) {
	println("DeliverTx begin")
	result := app.DeliverTxs([][]byte{tx})
	if result == nil {
		return types.OK
	}
	return result[0]
}

func (app *MainApplication) DeliverTxs(txs [][]byte) (ret []types.Result) {
	//gen init
	height := app.CurrentHeader.GetHeight()
	if height == 1 {
		app.apps[apps.App_Transfer].DeliverTx(nil)
		//return nil
	}
	if len(txs) == 0 {
		return nil
	}
	transactions := make([]*apps.Transaction, len(txs))
	for i := 0; i < len(txs); i++ {
		tx, err := apps.ParseTx(txs[i])
		if err != nil {
			ret = append(ret, types.NewError(types.CodeType_BaseInvalidInput, err.Error()))
			continue
		}
		transactions[i] = tx
		key := apps.KeyTx(tx.Hash)
		//tx 已经存在了,过滤掉
		if app.hasTx(key) {
			ret = append(ret, types.NewError(types.CodeType_BaseInvalidInput, "same tx"))
			continue
		}
		//用里面的txHash 作为ID,而不是外面的txHash
		app.db.Set([]byte(key), txs[i])
		if err != nil {
			ret = append(ret, types.NewError(types.CodeType_BaseInvalidInput, err.Error()))
			continue
		}
		err = app.processFee(tx, len(txs[i]))
		if err != nil {
			ret = append(ret, types.NewError(types.CodeType_InsufficientFunds, err.Error()))
			continue
		}
		if tx.AppId < int32(apps.App_Max) {
			if app.apps[tx.AppId] == nil {
				ret = append(ret, types.NewError(types.CodeType_BaseInvalidInput, "app not exist"))
				continue
			}
			//检查appId

			result := app.apps[tx.AppId].DeliverTx(tx.Data)
			ret = append(ret, result)
		}
	}
	//每个tx的receipt
	hashes := make([][]byte, len(ret))
	batch := app.db.NewBatch(true)
	for i := 0; i < len(ret); i++ {
		result := ret[i]
		var receipt apps.ResponseReceipt
		receipt.TxHash = transactions[i].Hash
		appId := transactions[i].GetAppId()
		conf, err := app.loadconfig(appId)
		if err != nil {
			panic(err)
		}
		receipt.AppId = conf.GetAppId()
		receipt.AppVersion = conf.GetAppVersion()
		if result.Code == types.CodeType_OK {
			receipt.IsOk = true
			//log.Println("result.data = ", result.Data)
			if result.Data == nil {
				//log.Println("result.data = ", receipt.TxHash)
				ret[i].Data = newResultHash(receipt.TxHash)
			}
			receipt.Result = ret[i].Data
		} else {
			receipt.Result = []byte(result.Log)
		}
		key := apps.KeyReceipt(receipt.AppVersion, receipt.TxHash)
		value, err := apps.Encode(&receipt)
		if err != nil {
			panic(err)
		}
		err = batch.Put([]byte(key), value)
		if err != nil {
			panic(err)
		}
		hashes[i] = apps.HashSha256(value)
	}
	err := batch.Commit()
	if err != nil {
		panic(err)
	}
	app.hash = apps.CalcMerkle(hashes)
	return ret
}

func (app *MainApplication) processFee(txmain *apps.Transaction, txlen int) error {
	signer := app.signs[txmain.SignType]
	pub, err := signer.GetPubKey(txmain.Sign)
	if err != nil {
		return err
	}
	fee := int64(app.unitfee) * int64(txlen)
	conf, err := app.loadconfig(int32(apps.App_Admin))
	if err != nil {
		panic(err)
	}
	adminpub := conf.GetAdmin()
	return app.apps[apps.App_Transfer].(*transfer.Application).TransferTo(pub, adminpub, transfer.BTY, int64(fee), true)
}

func (app *MainApplication) CheckTx(tx []byte) types.Result {
	height := app.CurrentHeader.GetHeight()
	if height == 0 {
		println("check tx ... 0")
		app.apps[apps.App_Transfer].DeliverTx(nil)
	}
	txmain, err := apps.ParseTx(tx)
	if err != nil {
		return types.NewError(types.CodeType_BaseInvalidInput, err.Error())
	}
	if len(tx) > 2*1024*1024 {
		return types.NewError(types.CodeType_BaseInvalidInput, "input too long")
	}
	key := apps.KeyTx(txmain.Hash)
	if _, ok := app.mempoolTx[key]; ok || app.hasTx(key) {
		return types.NewError(types.CodeType_BaseInvalidInput, "tx dup")
	}
	conf, err := app.loadconfig(txmain.GetAppId())
	if err != nil {
		return types.NewError(types.CodeType_BaseInvalidInput, err.Error())
	}
	signer := app.signs[txmain.SignType]
	if signer == nil {
		return types.NewError(types.CodeType_BaseInvalidInput, errSignFormat.Error())
	}
	pub, err := signer.GetPubKey(txmain.Sign)
	if err != nil {
		return types.NewError(types.CodeType_BaseInvalidPubKey, err.Error())
	}
	//非transfer的交易，必须有专门的sender
	if txmain.GetAppId() != int32(apps.App_Transfer) && !conf.IsSender(pub) {
		return types.NewError(types.CodeType_BaseInvalidPubKey, "sender error")
	}

	//签名方法修改
	//if !signer.Verify(txmain.Data, txmain.Sign) {
	if err = apps.CheckSign(txmain.Data, txmain.Sign[:32], txmain.Sign[32:]); err != nil {
		return types.NewError(types.CodeType_BaseInvalidSignature, "sign err")
	}
	//判断账户余额是否可以支付手续费的1000倍
	//这样设计主要是为了防止批量发送交易来攻击系统
	//同时，账户余额太小的账户无法进行交易太大的交易，也是有道理的。这样可以防止用户开很多账户影响系统性能。
	//小额的账户可以通过代理账户把钱汇总起来。
	//目前 1K 的 fee = 0.01BTY， 2M 的fee 是 2BTY 也就是是 账户有 2000 个BTY 就可以实现顶格发送交易，这个数字并不大。
	//大部分交易在 1K 左右，这样的话，只要账户中有 10个BTY，也就是 1块钱，就可以进行交易，就算这部分钱太少不能交易
	//了，用户也不回有很大的损失。
	balance := app.Balance(pub)
	fee := app.unitfee * len(tx)
	if 1000*int64(fee) > balance {
		return types.NewError(types.CodeType_BaseInsufficientFees, "balance too low")
	}
	//check for tx is dup,防止重复发送攻击
	if txmain.AppId < int32(apps.App_Max) {
		//inner tx
		var req apps.WriteRequest
		err = apps.Decode(txmain.Data, &req)
		if err != nil {
			return types.NewError(types.CodeType_BaseInvalidInput, err.Error())
		}
		err = req.CheckSign()
		if err != nil {
			return types.NewError(types.CodeType_BaseInvalidSignature, err.Error())
		}
		//检查某些基本条件
		if txmain.AppId != int32(req.GetAppId()) {
			return types.NewError(types.CodeType_BaseInvalidInput, "inner and out appid not the same")
		}
		result := app.apps[txmain.AppId].Check(&req)
		if result.Code != types.CodeType_OK {
			return result
		}
	}
	app.mempoolTx[key] = true
	return types.OK
}

func (app *MainApplication) Balance(pub []byte) int64 {
	b, _ := app.apps[apps.App_Transfer].(*transfer.Application).Balance(pub, transfer.BTY)
	return b
}

func (app *MainApplication) Commit() types.Result {
	return types.NewResultOK(app.hash, "")
}

func (app *MainApplication) Query(reqQuery types.RequestQuery) (resQuery types.ResponseQuery) {
	//查询暂时只有内部合约可以查询，外部合约的查询计划采用代理机制来提供统一的接口
	//查询权限开放给所有的人,写权限有限制
	var req apps.ReadRequest
	query := reqQuery.GetData()
	err := apps.Decode(query, &req)
	if err != nil {
		return apps.NewError(types.CodeType_BaseInvalidInput, err.Error())
	}
	if req.GetAppId() >= apps.App_Max {
		return apps.NewError(types.CodeType_BaseInvalidInput, "appId error")
	}
	application := app.apps[req.GetAppId()]
	if application == nil {
		return apps.NewError(types.CodeType_BaseInvalidInput, "app no exist")
	}
	return application.Query(reqQuery)
}

func (app *MainApplication) BeginBlock(hash []byte, header *types.Header) {
	app.CurrentHash = hash
	app.CurrentHeader = header
	for i := 0; i < int(apps.App_Last); i++ {
		if app.apps[apps.App(i)] == nil {
			continue
		}
		app.apps[apps.App(i)].BeginBlock(hash, header)
	}
}
