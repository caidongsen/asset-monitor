package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var account string

func monitor() {
	for _, v := range Conf.Accounts {
		account = v
		Acc := `{"uid":` + `"` + account + `"` + "}"
		for {
			<-CRechargeOK
			body, _ := Post(Acc)
			Parsing(body)
			time.Sleep(3 * time.Second)
		}
	}
}

/****************post请求*******************/
func Post(str string) ([]byte, error) {
	resp, err := http.Post(Conf.Api.WalletInfo,
		"application/json",
		strings.NewReader(str))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return body, nil
}

/*******************解析*********************/
func Parsing(body []byte) error {
	conf := InitConfig()
	h := make(map[string][]string)
	str := make(map[string]float64)
	var r interface{}
	err := json.Unmarshal([]byte(body), &r)
	if err != nil {
		fmt.Printf("err was %v", err)
		return err
	}
	data, ok := r.(map[string]interface{})
	if ok {
		for k, v := range data {
			switch v1 := v.(type) {
			case interface{}:
				for m1, n1 := range v1.(map[string]interface{}) {
					switch v2 := n1.(type) {
					case interface{}:
						for m2, n2 := range v2.(map[string]interface{}) {
							switch v3 := n2.(type) {
							case interface{}:
								switch v4 := v3.(type) {
								case float64:
									switch m1 {
									case "1":
										if m2 == "active" {
											str["CNY"] = v4
											if v4/1e8 < float64((*conf).Charges["CNY"].MinActiveAllowed) {
												h[account] = append(h[account], "CNY")
											}
										}
									case "2":
										if m2 == "active" {
											str["BTC"] = v4
											if v4/1e8 < float64((*conf).Charges["BTC"].MinActiveAllowed) {
												h[account] = append(h[account], "BTC")
											}
										}
									case "3":
										if m2 == "active" {
											str["BTY"] = v4
											if v4/1e8 < float64((*conf).Charges["BTY"].MinActiveAllowed) {
												h[account] = append(h[account], "BTY")
											}
										}
									case "4":
										if m2 == "active" {
											str["ETH"] = v4
											if v4/1e8 < float64((*conf).Charges["ETH"].MinActiveAllowed) {
												h[account] = append(h[account], "ETH")
											}
										}
									case "5":
										if m2 == "active" {
											str["ETC"] = v4
											if v4/1e8 < float64((*conf).Charges["ETC"].MinActiveAllowed) {
												h[account] = append(h[account], "ETC")
											}
										}
									case "7":
										if m2 == "active" {
											str["SC"] = v4
											if v4/1e8 < float64((*conf).Charges["SC"].MinActiveAllowed) {
												h[account] = append(h[account], "SC")
											}
										}
									case "8":
										if m2 == "active" {
											str["ZEC"] = v4
											if v4/1e8 < float64((*conf).Charges["ZEC"].MinActiveAllowed) {
												h[account] = append(h[account], "ZEC")
											}
										}
									case "9":
										if m2 == "active" {
											str["BTS"] = v4
											if v4/1e8 < float64((*conf).Charges["BTS"].MinActiveAllowed) {
												h[account] = append(h[account], "BTS")
											}
										}
									case "10":
										if m2 == "active" {
											str["LTC"] = v4
											if v4/1e8 < float64((*conf).Charges["LTC"].MinActiveAllowed) {
												h[account] = append(h[account], "LTC")
											}
										}
									case "11":
										if m2 == "active" {
											str["BCC"] = v4
											if v4/1e8 < float64((*conf).Charges["BCC"].MinActiveAllowed) {
												h[account] = append(h[account], "BCC")
											}
										}
									case "15":
										if m2 == "active" {
											str["USDT"] = v4
											if v4/1e8 < float64((*conf).Charges["USDT"].MinActiveAllowed) {
												h[account] = append(h[account], "USDT")
											}
										}
									case "17":
										if m2 == "active" {
											str["DCR"] = v4
											if v4/1e8 < float64((*conf).Charges["DCR"].MinActiveAllowed) {
												h[account] = append(h[account], "DCR")
											}
										}
									}
								}
							}
						}
					}
				}
			default:
				fmt.Println(k, "is another type not handle yet")
			}
		}
	}
	CData <- h
	return nil
}
func ParsingWeb(body []byte) (map[string]float64, error) {
	conf := InitConfig()
	h := make(map[string][]string)
	str := make(map[string]float64)
	var r interface{}
	err := json.Unmarshal([]byte(body), &r)
	if err != nil {
		fmt.Printf("err was %v", err)
		return nil, err
	}
	data, ok := r.(map[string]interface{})
	if ok {
		for k, v := range data {
			switch v1 := v.(type) {

			case interface{}:
				for m1, n1 := range v1.(map[string]interface{}) {
					switch v2 := n1.(type) {
					case interface{}:
						//	fmt.Printf("%s币种:\n", m1)
						for m2, n2 := range v2.(map[string]interface{}) {
							switch v3 := n2.(type) {
							case interface{}:
								//	fmt.Printf("%s的数量为%.2f  ", m2, v3)
								switch v4 := v3.(type) {
								case float64:
									switch m1 {
									case "1":
										if m2 == "active" {
											str["CNY"] = v4 / 1e+8
											if v4/1e8 < float64((*conf).Charges["CNY"].MinActiveAllowed) {
												h[account] = append(h[account], "CNY")
											}
										}
									case "2":
										if m2 == "active" {
											str["BTC"] = v4 / 1e+8
											if v4/1e8 < float64((*conf).Charges["BTC"].MinActiveAllowed) {
												h[account] = append(h[account], "BTC")
											}
										}

									case "3":
										if m2 == "active" {
											str["BTY"] = v4 / 1e+8
											if v4/1e8 < float64((*conf).Charges["BTY"].MinActiveAllowed) {
												h[account] = append(h[account], "BTY")
											}
										}
									case "4":
										if m2 == "active" {
											str["ETH"] = v4 / 1e+8
											if v4/1e8 < float64((*conf).Charges["ETH"].MinActiveAllowed) {
												h[account] = append(h[account], "ETH")
											}
										}
									case "5":
										if m2 == "active" {
											str["ETC"] = v4 / 1e+8
											if v4/1e8 < float64((*conf).Charges["ETC"].MinActiveAllowed) {
												h[account] = append(h[account], "ETC")
											}
										}
									case "7":
										if m2 == "active" {
											str["SC"] = v4 / 1e+8
											if v4/1e8 < float64((*conf).Charges["SC"].MinActiveAllowed) {
												h[account] = append(h[account], "SC")
											}
										}
									case "8":
										if m2 == "active" {
											str["ZEC"] = v4 / 1e+8
											if v4/1e8 < float64((*conf).Charges["ZEC"].MinActiveAllowed) {
												h[account] = append(h[account], "ZEC")
											}
										}
									case "9":
										if m2 == "active" {
											str["BTS"] = v4 / 1e+8
											if v4/1e8 < float64((*conf).Charges["BTS"].MinActiveAllowed) {
												h[account] = append(h[account], "BTS")
											}
										}
									case "10":
										if m2 == "active" {
											str["LTC"] = v4 / 1e+8
											if v4/1e8 < float64((*conf).Charges["LTC"].MinActiveAllowed) {
												h[account] = append(h[account], "LTC")
											}
										}
									case "11":
										if m2 == "active" {
											str["BCC"] = v4 / 1e+8
											if v4/1e8 < float64((*conf).Charges["BCC"].MinActiveAllowed) {
												h[account] = append(h[account], "BCC")
											}
										}
									case "15":
										if m2 == "active" {
											str["USDT"] = v4 / 1e+8
											if v4/1e8 < float64((*conf).Charges["USDT"].MinActiveAllowed) {
												h[account] = append(h[account], "USDT")
											}
										}
									case "17":
										if m2 == "active" {
											str["DCR"] = v4 / 1e+8
											if v4/1e8 < float64((*conf).Charges["DCR"].MinActiveAllowed) {
												h[account] = append(h[account], "DCR")
											}
										}
									}
								}
							}
						}
					}
				}
			default:
				fmt.Println(k, "is another type not handle yet")
			}
		}
	}
	return str, nil

}
