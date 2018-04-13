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

var log_sms = l.New("module", "main/sms")

func SmsSend() {
	//从信道读取数据
	mailInfo := <-ChSms
	account := mailInfo[0]
	currency := mailInfo[1]
	amount := mailInfo[2]
	rea := strings.EqualFold(mailInfo[3], "true")
	errMsg := mailInfo[4]
	phone_numbers := Conf.Receipts.Sms

	//循环每个号码，进行发短信操作
	for i := 0; i < len(phone_numbers); i++ {
		smsSend(phone_numbers[i], account, currency, amount, rea, errMsg)
	}
}

func smsSend(phone_number string, account string, currency string, amount string, rea bool, errMsg string) {
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
	smsUrl := Conf.Api.Sms

	// //以表单形式发送消息给sms的url
	body, err := HttpPostFormReq(smsUrl, url.Values{"mobile": {phone_number}, "codetype": {"notice"}, "vparam": {sendStr}})
	if err != nil {
		log_sms.Warn("HttpPostFormReq err", "err", err)
		// log.Println(err)
	}

	//将返回的不确定类型的json消息解析出来
	var tresult interface{}
	err = json.Unmarshal(body, &tresult)
	if err != nil {
		log_sms.Warn("Unmarshal err", "err", err)
		// log.Println(err)
	}

	//用切片来获取json消息
	result, ok := tresult.(map[string]interface{})
	if !ok {
		log_sms.Warn("invalid result")
		// log.Println("invalid result")
	}

	//获取json中code的值
	icode, err := JsonMsgProcess("code", result)
	if err != nil {
		log_sms.Warn("JsonMsgProcess err", "err", err)
		// log.Println(err)
		return
	}

	//根据code值判断短信发送是否成功
	if icode != 200 {
		req_err, err := JsonMsgProcess("error", result)
		if err != nil {
			log_sms.Warn("JsonMsgProcess err", "err", err)
			// log.Println(err)
			return
		}
		errstr := req_err.(string)
		log_sms.Warn("sms send failed", "err", errstr)
		// log.Println("sms send failed : " + errstr)
	} else {
		log_sms.Info("sms send succeeded", "phone_number", phone_number,
			"account", account, "currency", currency, "amount", amount)
	}
}

//表单消息的http请求，返回值为json格式消息的[]byte值和error
func HttpPostFormReq(api_url string, msg url.Values) ([]byte, error) {
	resp, err := http.PostForm(api_url, msg)
	if nil != resp {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	//读取json消息
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

//Json消息的http请求，返回值为json格式消息的[]byte值和error
func HttpPostJsonReq(api_url string, msg string) ([]byte, error) {
	resp, err := http.Post(api_url, "application/json", strings.NewReader(msg))
	if nil != resp {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	//读取json消息
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

//解析任意格式的json消息，传入key和json的map，返回interface{}和err
func JsonMsgProcess(key string, result map[string]interface{}) (interface{}, error) {
	var msg interface{}
	var err error
	switch t := result[key].(type) {
	case nil:
		err = fmt.Errorf("req[\"" + key + "\"] type is nil")
		return nil, err
	case float64:
		msg = (int)(t)
	case string:
		msg = t
	case bool:
		msg = t
	case []interface{}:
		msg = t
	case map[string]interface{}:
		msg = t
	default:
		err = fmt.Errorf("req[\"" + key + "\"] is unknow type")
		return nil, err
	}
	return msg, err
}
