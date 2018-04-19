package apps

import "encoding/hex"

//这是最简单的一个版本管理系统
//目前修改版本必须重放所有的交易
type SystemAccount struct {
	admin      []byte
	sender     map[string]bool
	appVersion int32
	appId      int32
}

func CreateSystemApp(appId int32, appVersion int32, admin string, senders []string) *ResponseAppConf {
	app := &ResponseAppConf{}
	app.AppId = appId
	app.AppVersion = appVersion
	badmin, err := hex.DecodeString(admin)
	if err != nil {
		panic(err)
	}
	app.Admin = badmin
	for i := 0; i < len(senders); i++ {
		bsender, err := hex.DecodeString(senders[i])
		if err != nil {
			panic(err)
		}
		app.Senders = append(app.Senders, bsender)
	}
	return app
}

func NewSystemAccount(app *ResponseAppConf) *SystemAccount {
	acc := &SystemAccount{}
	acc.appVersion = app.AppVersion
	acc.appId = app.AppId
	acc.admin = app.GetAdmin()
	acc.sender = make(map[string]bool)
	for i := 0; i < len(app.Senders); i++ {
		sender := app.Senders[i]
		acc.sender[string(sender)] = true
	}
	return acc
}

func (acc *SystemAccount) GetAppId() int32 {
	return acc.appId
}

func (acc *SystemAccount) GetAppVersion() int32 {
	return acc.appVersion
}

func (acc *SystemAccount) GetAdmin() []byte {
	return acc.admin
}

func (acc *SystemAccount) SetAdmin(admin []byte) error {
	if len(admin) != 32 {
		return ErrPubKeyFormat
	}
	acc.admin = admin
	return nil
}

func (acc *SystemAccount) IsSender(pub []byte) bool {
	_, ok := acc.sender[string(pub)]
	return ok
}

func (acc *SystemAccount) AddSender(pub []byte) error {
	if len(pub) != 32 {
		return ErrPubKeyFormat
	}
	if len(acc.sender) > 10 {
		return ErrTooManySender
	}
	acc.sender[string(pub)] = true
	return nil
}

func (acc *SystemAccount) DelSender(pub []byte) error {
	if len(pub) != 32 {
		return ErrPubKeyFormat
	}
	if !acc.IsSender(pub) {
		return ErrNotSender
	}
	delete(acc.sender, string(pub))
	return nil
}

func (acc *SystemAccount) GetSystemApp() *ResponseAppConf {
	app := &ResponseAppConf{}
	app.AppId = acc.appId
	app.AppVersion = acc.appVersion
	app.Admin = acc.admin
	for key, _ := range acc.sender {
		app.Senders = append(app.Senders, []byte(key))
	}
	return app
}

func (acc *SystemAccount) SetAppVersion(version int32) error {
	if (acc.appVersion + 1) != version {
		return ErrVersion
	}
	acc.appVersion = version
	return nil
}

type IConfig interface {
	LoadConfig(appId int32, update bool) (*SystemAccount, error)
}
