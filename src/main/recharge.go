package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type RechargeMessage struct {
	Symbol string `json:"symbol"` //json格式的币种
	Amount string `json:"amount"` //数量
}

type RechargeResult struct {
	Code string `json:"code"` //"0"成功 "1"失败
	Msg  string `json:"msg"`
	Data string `json:"data"`
}

type Recharge struct {
}

func NewRecharge() *Recharge {
	r := &Recharge{}
	return r
}

func (r *Recharge) RechargeInbank(uid string, symbol string) (*RechargeResult, error) {
	timeStr := time.Now().Format("2006-01-02 15:04:05")
	amount := fmt.Sprintf("%d", Conf.Charges[symbol].RechargeAmount)
	reqUrl := Conf.Api.Transfer

	//将json消息需要的信息写入结构体
	tRechargeMessage := &RechargeMessage{}
	tRechargeMessage.Symbol = symbol
	tRechargeMessage.Symbol = amount

	//将数据类型转换为json格式
	tRechargeMessageBytes, err := json.Marshal(tRechargeMessage)
	if nil != err {
		return nil, err
	}

	//调用发送json消息的工具函数
	body, err := HttpPostJsonReq(reqUrl, string(tRechargeMessageBytes[:]))
	if err != nil {
		return nil, err
	}

	//将收到的[]byte转成json格式
	tRechargeResult := &RechargeResult{}
	err = json.Unmarshal(body, tRechargeResult)
	if err != nil {
		return nil, err
	}

	//添加读写锁
	RWMutex.Lock()

	//判断充值是否成功
	var suc string
	if tRechargeResult.Code == "0" {
		suc = "true"
	} else {
		suc = "false"
	}

	//调用传短信和传邮件程序
	ChMail <- []string{uid, symbol, amount, suc, tRechargeResult.Msg}
	ChSms <- []string{uid, symbol, amount, suc, tRechargeResult.Msg}
	go SmsSend()
	go MailSend()

	//写入数据库 以用户名_币种_时间_充值数量
	var dbKey string
	dbKey = fmt.Sprintf("%s_%s_%s_%s", uid, symbol, timeStr, amount)
	err = Db.Put([]byte(dbKey), []byte(suc /*value*/), nil)
	if err != nil {
		return nil, err
	}

	//解锁
	RWMutex.Unlock()

	//返回消息，调用接口处用*RechargeResult.Code来判断充值是否成功
	return tRechargeResult, nil
}
