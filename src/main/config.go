package main

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Accounts []string
	Rpc      string
	Api      Api
	Receipts Receipts
	Charges  map[string]CoinCharge
}

type Api struct {
	Sms        string
	Mail       string
	WalletInfo string
	Transfer   string
	node       string
	fromkey    string
}

type Receipts struct {
	Sms  []string
	Mail []string
}

type CoinCharge struct {
	CoinId           int //币种代号
	MinActiveAllowed int //最低阈值
	RechargeAmount   int //每次充值数量
	MaxRechargeTimes int //每天最多充值次数
}

func InitConfig() *Config {
	var conf Config
	if _, err := toml.DecodeFile("../../etc/config.toml", &conf); err != nil {
		panic(err)
	}
	return &conf
}
