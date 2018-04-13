package main

import (
	"encoding/json"
	l "github.com/inconshreveable/log15"
	"strconv"
	"time"
)

var log_monitor = l.New("module", "main/monitor")

type MonitorData struct {
	M_monitorCoinMsg map[string]MonitorCoinMsg `json:"data"`
}

type MonitorCoinMsg struct {
	Active   float64 `json:"active"`   //余额
	Frozen   float64 `json:"frozen"`   //冻结资产
	Poundage float64 `json:"poundage"` //手续费
}

//存放币种代号和币种的信息  map[int]string
func InitMapCoinIdCharge() map[int]string {
	var coinIdCharge = make(map[int]string)
	for coinCharge, coindata := range Conf.Charges {
		coinIdCharge[coindata.CoinId] = coinCharge //[币种代号]币种
	}
	return coinIdCharge
}

func monitor() {
	for {
		for _, account := range Conf.Accounts {
			reqmsg := `{"uid":` + `"` + account + `"` + "}"
			<-ChRechargeOK
			body, err := HttpPostJsonReq(Conf.Api.WalletInfo, reqmsg)
			if err != nil {
				log_monitor.Warn("HttpPostJsonReq err", "err", err)
				// log.Println(err)
				continue
			}
			ParsingRecharge(account, body)
		}
		time.Sleep(3 * time.Second)
	}
}

func ParsingRecharge(account string, body []byte) error { //解析  返回一个map给充值判断接口
	strRecharge := make(map[string][]string)
	var m MonitorData
	err := json.Unmarshal([]byte(body), &m)
	if err != nil {
		log_monitor.Warn("Unmarshal err", "err ", err)
		// log.Println(err)
		ChRechargeOK <- true
		return err
	}
	for CoinId, CoinName := range CoinIdCharge {
		Recharge := m.M_monitorCoinMsg[strconv.Itoa(CoinId)].Active / 1e8 //账户余额
		minRecharge := float64(Conf.Charges[CoinName].MinActiveAllowed)   //最低余额
		if Recharge < minRecharge {
			strRecharge[account] = append(strRecharge[account], CoinName)
		}

	}
	ChData <- strRecharge
	return nil
}

func ParsingWeb(body []byte) (map[string]float64, error) {
	strWeb := make(map[string]float64)
	var m MonitorData
	err := json.Unmarshal([]byte(body), &m)
	if err != nil {
		return nil, err
	}
	for CoinId, CoinName := range CoinIdCharge {
		strWeb[CoinName] = m.M_monitorCoinMsg[strconv.Itoa(CoinId)].Active / 1e8 //账户余额
	}
	return strWeb, nil

}
