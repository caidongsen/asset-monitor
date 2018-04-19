package reg

import (
	"bytes"
	"sync"

	"github.com/tendermint/abci/types"
	"github.com/tendermint/tendermint/apps"
	godb "github.com/tendermint/tendermint/apps/kvdb"
)

type Application struct {
	types.BaseApplication
	db           godb.DB
	mu           sync.Mutex
	systemconfig map[apps.App]*apps.SystemAccount
}

func NewApplication(db godb.DB, applast int32) *Application {
	main := &Application{types.BaseApplication{}, db, sync.Mutex{}, make(map[apps.App]*apps.SystemAccount)}

	main.systemconfig[apps.App_Transfer] = apps.NewSystemAccount(apps.CreateSystemApp(int32(apps.App_Transfer), 1, "fc19b9384e58c968455f080a3247290d3f061fb4843b1513f5bca26e0a656bea", []string{"999e527b26f59b6ca49d0a3d55ff7983df2afef478900930f60830a6044f70bd", "f18ea7ac4d7e28503ee1c105d5a6d91dede625eeacf65c2c6136209de9d1b8dc"}))

	main.systemconfig[apps.App_Reg] = apps.NewSystemAccount(apps.CreateSystemApp(int32(apps.App_Reg), 1, "b99b96d76f3269e4b299b890b832c601baed1e161b5ddb0867c973569bd5c23a", []string{"ca56d67684a0ae7c5a154301818b97c0fc5969e3dadf6af392cafa20ce63e93d", "85faa6aaa166752ab9e0175d2d6e551596c50ea11f4bc3337748d099dcf94ac0"}))

	main.systemconfig[apps.App_Receipt] = apps.NewSystemAccount(apps.CreateSystemApp(int32(apps.App_Receipt), 1, "801848366ebaeab54e475068a797dc65d92ec70835f45b21e24dc3bdb3e6f283", []string{"abae68f32628cc9348ead30ab3b9b4ba0fde7a60f4d397d7c9d06fc96e4d2a97", "bdf2cdbb7054f41aceafd277dfb52d2ff8454fc057ea8109cccde618ce18b832"}))

	main.systemconfig[apps.App_Admin] = apps.NewSystemAccount(apps.CreateSystemApp(int32(apps.App_Admin), 1, "cafeed9f80b9bd7c1baecc0408fb73e4e66fba2f05b3418d900bbc56554efac2", []string{"498c6d7686d4ee44046700139022d76ebceeea3a0e34998533b8ad06f1d23fe8", "48e2314156ec6524eae155243331029f047ca4e8bf80b3653b2da21e29f2e2d6"}))

	main.systemconfig[apps.App_AccountManage] = apps.NewSystemAccount(apps.CreateSystemApp(int32(apps.App_AccountManage), 1, "fc19b9384e58c968455f080a3247290d3f061fb4843b1513f5bca26e0a656bea", []string{"85faa6aaa166752ab9e0175d2d6e551596c50ea11f4bc3337748d099dcf94ac0", "fc19b9384e58c968455f080a3247290d3f061fb4843b1513f5bca26e0a656bea"}))

	for i := 0; i < int(applast); i++ {
		main.loadconfig(int32(i), true)
	}
	return main
}

func (app *Application) LoadConfig(appId int32, update bool) (*apps.SystemAccount, error) {
	app.mu.Lock()
	defer app.mu.Unlock()
	return app.loadconfig(appId, update)
}

func (app *Application) loadconfig(appId int32, update bool) (*apps.SystemAccount, error) {
	if !update {
		if conf, ok := app.systemconfig[apps.App(appId)]; ok {
			return conf, nil
		}
	}
	key := apps.KeyConf(appId)
	value := app.db.Get([]byte(key))
	if value == nil {
		return nil, apps.ErrNotFound
	}
	var appconf apps.ResponseAppConf
	err := apps.Decode(value, &appconf)
	if err != nil {
		panic(err)
	}
	app.systemconfig[apps.App(appId)] = apps.NewSystemAccount(&appconf)
	return app.systemconfig[apps.App(appId)], nil
}

func (app *Application) create(c *apps.RequestCreateApp) error {
	app.mu.Lock()
	defer app.mu.Unlock()
	_, err := app.loadconfig(c.GetAppId(), false)
	if err == nil {
		return apps.ErrExist
	}
	conf := apps.ResponseAppConf{}
	conf.Admin = c.GetAdmin()
	conf.AppId = c.GetAppId()
	conf.AppVersion = 1
	app.systemconfig[apps.App(conf.AppId)] = apps.NewSystemAccount(&conf)
	app.saveconfig(conf.AppId)
	return nil
}

func (app *Application) setadmin(appId int32, admin []byte) error {
	app.mu.Lock()
	defer app.mu.Unlock()
	conf, err := app.loadconfig(appId, false)
	if err != nil {
		return err
	}
	err = conf.SetAdmin(admin)
	if err != nil {
		return err
	}
	return app.saveconfig(appId)
}

func (app *Application) addsender(appId int32, sender []byte) error {
	app.mu.Lock()
	defer app.mu.Unlock()
	conf, err := app.loadconfig(appId, false)
	if err != nil {
		return err
	}
	err = conf.AddSender(sender)
	if err != nil {
		return err
	}
	return app.saveconfig(appId)
}

func (app *Application) delsender(appId int32, sender []byte) error {
	app.mu.Lock()
	defer app.mu.Unlock()
	conf, err := app.loadconfig(appId, false)
	if err != nil {
		return err
	}
	err = conf.DelSender(sender)
	if err != nil {
		return err
	}
	return app.saveconfig(appId)
}

func (app *Application) setappversion(appId int32, version int32) error {
	app.mu.Lock()
	defer app.mu.Unlock()
	conf, err := app.loadconfig(appId, false)
	if err != nil {
		return err
	}
	err = conf.SetAppVersion(version)
	if err != nil {
		return err
	}
	return app.saveconfig(appId)
}

func (app *Application) saveconfig(appId int32) error {
	conf, err := app.loadconfig(appId, false)
	if err != nil {
		return err
	}
	sapp := conf.GetSystemApp()
	data, err := apps.Encode(sapp)
	if err != nil {
		panic(err)
	}
	app.db.Set([]byte(apps.KeyConf(appId)), data)
	return nil
}

func (app *Application) DeliverTxs(txs [][]byte) []types.Result {
	ret := make([]types.Result, len(txs))
	for i := 0; i < len(ret); i++ {
		ret[i] = types.NewResultOK(nil, "")
	}
	return ret
}

//acction list
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
	if actionId == apps.MessageType_MsgCreateApp {
		create := req.GetCreateApp()
		//余额不够，不扣钱, 返回错误，但是如果是手续费扣费，那么扣为0
		err := app.create(create)
		if err != nil {
			return types.NewError(types.CodeType_BaseInsufficientFunds, err.Error())
		}
		return types.OK
	} else if actionId == apps.MessageType_MsgAddSender {
		sender := req.GetAddSender()
		err := app.addsender(sender.AppId, sender.Sender)
		if err != nil {
			return types.NewError(types.CodeType_BaseInsufficientFunds, err.Error())
		}
		return types.OK
	} else if actionId == apps.MessageType_MsgDelSender {
		sender := req.GetDelSender()
		err := app.delsender(sender.AppId, sender.Sender)
		if err != nil {
			return types.NewError(types.CodeType_BaseInsufficientFunds, err.Error())
		}
	} else if actionId == apps.MessageType_MsgSetAdmin {
		admin := req.GetSetAdmin()
		err := app.setadmin(admin.AppId, admin.Admin)
		if err != nil {
			return types.NewError(types.CodeType_BaseInsufficientFunds, err.Error())
		}
	} else if actionId == apps.MessageType_MsgSetAppVersion {
		appversion := req.GetSetAppVersion()
		err := app.setappversion(appversion.AppId, appversion.AppVersion)
		if err != nil {
			return types.NewError(types.CodeType_BaseInsufficientFunds, err.Error())
		}
	}
	return types.NewError(types.CodeType_BaseInvalidInput, "err actionId")
}

func (app *Application) CheckTx(tx []byte) types.Result {
	return types.NewResultOK(nil, "")
}

func (app *Application) Check(req *apps.WriteRequest) types.Result {
	actionId := req.GetMsgType()
	if actionId == apps.MessageType_MsgCreateApp {
		createapp := req.GetCreateApp()
		/*
		   regapp, _ := app.LoadConfig(int32(apps.App_Reg), false)
		   //sender is ok
		   if !bytes.Equal(req.GetAccount(), regapp.GetAdmin()) {
		       return types.NewError(types.CodeType_BaseInvalidInput, "account must reg.admin")
		   }
		*/
		_, err := app.LoadConfig(createapp.GetAppId(), false)
		if err == nil {
			return types.NewError(types.CodeType_BaseInvalidInput, "appid exist")
		}
		if len(createapp.GetAdmin()) != 32 {
			return types.NewError(types.CodeType_BaseInvalidInput, "admin format error")
		}
	} else if actionId == apps.MessageType_MsgAddSender || actionId == apps.MessageType_MsgDelSender {
		item := req.GetAddSender()
		conf, err := app.LoadConfig(item.GetAppId(), false)
		if err != nil {
			return types.NewError(types.CodeType_BaseInvalidInput, err.Error())
		}
		if !bytes.Equal(req.GetAccount(), conf.GetAdmin()) {
			return types.NewError(types.CodeType_BaseInvalidInput, "account must admin")
		}
		if len(item.GetSender()) != 32 {
			return types.NewError(types.CodeType_BaseInvalidInput, "sender format error")
		}
	} else if actionId == apps.MessageType_MsgSetAdmin {
		item := req.GetSetAdmin()
		conf, err := app.LoadConfig(item.GetAppId(), false)
		if err != nil {
			return types.NewError(types.CodeType_BaseInvalidInput, err.Error())
		}
		if !bytes.Equal(req.GetAccount(), conf.GetAdmin()) {
			return types.NewError(types.CodeType_BaseInvalidInput, "account must admin")
		}
		if len(item.GetAdmin()) != 32 {
			return types.NewError(types.CodeType_BaseInvalidInput, "admin format error")
		}
	} else if actionId == apps.MessageType_MsgSetAppVersion {
		item := req.GetSetAppVersion()
		conf, err := app.LoadConfig(item.GetAppId(), false)
		if err != nil {
			return types.NewError(types.CodeType_BaseInvalidInput, err.Error())
		}
		if !bytes.Equal(req.GetAccount(), conf.GetAdmin()) {
			return types.NewError(types.CodeType_BaseInvalidInput, "account must admin")
		}
		if (conf.GetAppVersion() + 1) != item.GetAppVersion() {
			return types.NewError(types.CodeType_BaseInvalidInput, "appversion error")
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
	if actionId == apps.MessageType_MsgAppConf {
		query := req.GetAppConf()
		conf, err := app.LoadConfig(query.AppId, false)
		if err != nil {
			return apps.NewError(types.CodeType_BaseInvalidInput, err.Error())
		}
		data, err := apps.Encode(&apps.Response{Value: &apps.Response_AppConf{conf.GetSystemApp()}})
		if err != nil {
			panic(err)
		}
		return apps.NewData(types.CodeType_OK, data)
	}
	return apps.NewError(types.CodeType_BaseInvalidInput, "err actionId")
}
