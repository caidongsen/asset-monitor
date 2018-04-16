package main

import (
	"html/template"
	"net/http"
	"regexp"
	"strings"

	l "github.com/inconshreveable/log15"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var log_web = l.New("module", "main/web")

type Asset struct {
	UserId     string
	UserAsset  map[string]float64
	UserRecord []Record
}

type Record struct {
	CoinName string
	Time     string
	Number   string
}

func home(w http.ResponseWriter, r *http.Request) { // 用户查询界面
	t := template.New("home.html")
	t, _ = t.ParseFiles("home.html")
	t.ExecuteTemplate(w, "home", "")
}

func GetuUserRecord(uid string) []Record {
	RWMutex.Lock()
	userRecord := make([]Record, 0)
	//iter := Db.NewIterator(nil, nil)
	seekStr := uid + "_" //前缀查询数据
	iter := Db.NewIterator(util.BytesPrefix([]byte(seekStr)), nil)
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		all := strings.Split(string(key), "_")
		if string(value) == "true" {
			// do something for the key
			record := Record{
				CoinName: all[1],
				Time:     all[2],
				Number:   all[3],
			}
			userRecord = append(userRecord, record)
		}
	}
	iter.Release()
	err := iter.Error()
	RWMutex.Unlock()
	if err != nil {
		log_web.Warn("err ", err)
		// srvlog := log.New("iter", "is not close")
		// //log.Println(err)
		// srvlog.Warn("There is something wrong")
		return nil
	}
	return userRecord
}

func process(w http.ResponseWriter, r *http.Request) { // 用户信息页面
	//对获取的表单进行识别，防止恶意代码
	if m, _ := regexp.MatchString(`^([0-9a-z]{64})$`, r.FormValue("uid")); !m {
		t := template.New("account_fa.html")
		t, _ = t.ParseFiles("account_fa.html")
		t.ExecuteTemplate(w, "account_fa", "")
		return
	}

	PostContent := `{"uid":"` + r.FormValue("uid") + `"}`
	//读取账户信息
	body, err := HttpPostJsonReq(Conf.Api.WalletInfo, PostContent)
	if err != nil {
		log_web.Warn("err ", err)
		// srvlog := log.New("web :", err)
		// srvlog.Warn("There is something wrong")
		return
	}

	// body := Post(PostContent)
	if string(body) == "params error: invalid uid" {
		//若不存在该账户，则执行show_
		t := template.New("account_fa.html")
		t, _ = t.ParseFiles("account_fa.html")
		t.ExecuteTemplate(w, "account_fa", "")
		return
	}
	var userAsset = make(map[string]float64, 0)
	userAsset, err = ParsingWeb(body)
	//userAsset, err := ParsingWeb_(body)
	if err != nil {
		log_web.Warn("err ", err)
		// srvlog := log.New("parsing :", err)
		// srvlog.Warn("There is something wrong")
		return
	}

	var userRecord = GetuUserRecord(r.FormValue("uid"))

	//data为放到页面中的数据
	data := Asset{
		UserId:     r.FormValue("uid"),
		UserAsset:  userAsset,
		UserRecord: userRecord,
	}

	t := template.New("account_ac.html")
	t, _ = t.ParseFiles("account_ac.html")
	t.ExecuteTemplate(w, "account_ac", data)
	return
}
