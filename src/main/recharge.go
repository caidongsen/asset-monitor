package main

import (
	//	"encoding/json"
	"dev.33.cn/33/btrade/msq"
	cmn "dev.33.cn/33/common"
	rpc "dev.33.cn/33/trade_tools/jsonrpc"
	"fmt"
	"strconv"
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

func (r *Recharge) RechargeInbank(key *[64]byte, currency int32, symbol string, uid string, to *[32]byte, amount_ int, node string) (*RechargeResult, error) { // (*RechargeResult, error) {
	tRechargeResult := &RechargeResult{}
	timeStr := time.Now().Format("2006-01-02 15:04:05")
	amount := strconv.Itoa(amount_)
	req := &msq.RequestTransfer{}
	req.InstructionId = cmn.RandInstructionId()
	req.Amount = int64(amount_)
	req.SymbolId = currency
	req.ActionId = msq.MessageType_MsgTransfer
	req.Uid = key[32:]
	req.ToAddr = to[:]
	data, err := msq.MarshalMessage(req)
	if err != nil {
		tRechargeResult.Code = "1"
		tRechargeResult.Msg = " transfer fail"
		return nil, err

	}
	request := &msq.WriteRequest{}
	request.Value = &msq.WriteRequest_Transfer{req}
	request.Sign = cmn.Signdata(key, data) //1
	request.CreateTime = time.Now().Unix() //2

	resp, err := rpc.Send(request, node)
	if err != nil {
		tRechargeResult.Code = "1"
		tRechargeResult.Msg = " transfer fail"
		return nil, err
	}
	fmt.Println("data = ", resp)
	tRechargeResult.Code = "0"
	tRechargeResult.Msg = "transfer success"
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
