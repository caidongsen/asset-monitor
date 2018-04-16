package main

import (
	l "github.com/inconshreveable/log15"
	"github.com/syndtr/goleveldb/leveldb"
	"net/http"
	"sync"
)

var Version string = "0.1"
var log_main = l.New("module", "main")
var Conf = InitConfig()
var Mutex sync.Mutex
var RWMutex sync.RWMutex
var Db *leveldb.DB
var ChData = make(chan map[string][]string, 100) //查询数据信道
var ChMail = make(chan []string, 100)            //发送邮件信道
var ChSms = make(chan []string, 100)             //发送短信信道
var ChRechargeOK = make(chan bool)
var Path = "../db"
var reChargeMsg = NewRecharge()
var reChargeResult RechargeResult
var CoinIdCharge = InitMapCoinIdCharge()

func main() {
	log_main.Info("project version:" + Version)
	// log.Println("version : ", Version)
	var err error
	Db, err = leveldb.OpenFile(Path, nil)
	if err != nil {
		log_main.Info("err", err)
		// log.Println(err)
	}
	defer Db.Close()
	go monitor()
	go recharge()

	ChRechargeOK <- true
	server := http.Server{
		Addr: "127.0.0.1:8080",
	}

	http.HandleFunc("/home", home)
	http.HandleFunc("/process", process)
	server.ListenAndServe()
}
