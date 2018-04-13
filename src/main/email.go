package main

import (
	"encoding/json"
	"fmt"
	l "github.com/inconshreveable/log15"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var log_email = l.New("module", "main/email")

func MailSend() {
	//从信道读取数据
	mailInfo := <-ChMail
	account := mailInfo[0]
	currency := mailInfo[1]
	amount := mailInfo[2]
	rea := strings.EqualFold(mailInfo[3], "true")
	errMsg := mailInfo[4]
	emails := Conf.Receipts.Mail

	//循环每个邮箱，进行发邮件操作
	for i := 0; i < len(emails); i++ {
		mailSend(emails[i], account, currency, amount, rea, errMsg)
	}
}

func mailSend(email string, account string, currency string, amount string, rea bool, errMsg string) {
	//当前时间的字符串
	timeStr := time.Now().Format("2006-01-02 15:04:05")
	var sendStr string
	if rea == true {
		sendStr = fmt.Sprintf("充值成功: On %s,  asset monitor tool had successfully "+
			"recharged to account %s with %s %s。", timeStr, account, amount, currency)
	} else {
		sendStr = fmt.Sprintf("充值失败: On %s,  asset monitor tool "+
			"failed to recharged to  account %s with %s %s， reason：[%s]。", timeStr, account, amount, currency, errMsg)
	}

	//从全局配置变量获取数据
	mailUrl := Conf.Api.Mail

	//以表单形式发送消息给email的url
	resp, err := http.PostForm(mailUrl, url.Values{"email": {email}, "codetype": {"notice"}, "vparam": {sendStr}})
	if nil != resp {
		defer resp.Body.Close()
	}
	if err != nil {
		log_email.Warn("PostForm err", "err", err)
		// log.Println(err)
	}

	//读取json消息
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log_email.Warn("ReadAll err", "err", err)
		// log.Println(err)
	}

	//将返回的不确定类型的json消息解析出来
	var tresult interface{}
	err = json.Unmarshal(body, &tresult)
	if err != nil {
		log_email.Warn("Unmarshal err", "err", err)
		// log.Println(err)
	}

	//用切片来获取json消息
	result, ok := tresult.(map[string]interface{})
	if !ok {
		log_email.Warn("type err", "err", "invalid result")
		// log.Println("invalid result")
	}

	//获取json中code的值
	icode, err := JsonMsgProcess("code", result)
	if err != nil {
		log_email.Warn("JsonMsgProcess err", "err", err)
		// log.Println(err)
		return
	}

	//根据code值判断短信发送是否成功
	if icode != 200 {
		req_err, err := JsonMsgProcess("error", result)
		if err != nil {
			log_email.Warn("JsonMsgProcess err", "err ", err)
			// log.Println(err)
			return
		}
		errstr := req_err.(string)
		log_email.Warn("email send failed", "err", errstr)
		// log.Println("email send failed : " + errstr)
	}
}
