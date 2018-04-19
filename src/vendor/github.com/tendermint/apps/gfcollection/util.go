package gfcollection

import (
	//"bytes"
	"dev.33.cn/33/crypto"
	_ "dev.33.cn/33/crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

var ErrDupInstructionId = errors.New("ErrDupInstructionId")
var ErrWrongMessageType = errors.New("wrong message type")
var ErrSignWrongLength = errors.New("wrong length")
var ErrAdminExist = errors.New("admin exist")
var ErrAdminNotExist = errors.New("admin not exist")
var ErrUserExist = errors.New("user exist")
var ErrUserNotExist = errors.New("user not exist")
var ErrLoanExist = errors.New("loan exist")
var ErrLoanNotExist = errors.New("loan not exist")
var ErrPackNotExist = errors.New("pack not exist")
var ErrStorage = errors.New("storage error")
var ErrNoRight = errors.New("no right")
var ErrNotAdmin = errors.New("not admin")
var ErrWrongPubkey = errors.New("wrong pubkey")
var ErrNotCreator = errors.New("not creator")
var ErrWrongState = errors.New("wrong state")
var ErrWrongLoanState = errors.New("wrong loan state")
var ErrRmbNotEnough = errors.New("rmb not enough")
var ErrWrongUserType = errors.New("wrong user type")
var ErrWrongPackAmount = errors.New("err pack amount")
var ErrPlatformExist = errors.New("platform exist")
var ErrPlatformNotExist = errors.New("platform not exist")
var ErrPackageNotExist = errors.New("package not exist")
var ErrPackageExist = errors.New("package exist")
var ErrWrongAdminType = errors.New("err admin type")
var ErrSign = errors.New("errsign")
var ErrPackConfNotExist = errors.New("packconf not exist")
var ErrPackConfExist = errors.New("packconf exist")
var ErrWrongAmount = errors.New("wrong amount")
var ErrWrongPackConfState = errors.New("wrong packconf state")
var ErrWrongPackNums = errors.New("wrong packnums")
var ErrBankExists = errors.New("bank exists")
var ErrBankNotExists = errors.New("bank not exists")
var ErrCompanyExists = errors.New("company exists")
var ErrCompanyNotExists = errors.New("company not exists")
var ErrCompanyConfExists = errors.New("company conf exists")
var ErrCompanyConfNotExists = errors.New("company conf not exists")
var ErrCaseConfExists = errors.New("caseconf exists")
var ErrCaseConfNotExists = errors.New("caseconf not exists")
var ErrCaseExists = errors.New("case exists")
var ErrCaseNotExists = errors.New("case not exists")
var ErrException = errors.New("exception error")
var ErrNoEmpty = errors.New("list not empty")
var ErrAreaPoolNotExists = errors.New("area pool not exists")
var ErrAreaListNotExist = errors.New("area list not exist")
var ErrWrongWeight = errors.New("wrong weight")

func pad32(i int32) string {
	return fmt.Sprintf("%010d", i)
}

func pad64(i int64) string {
	return fmt.Sprintf("%020d", i)
}

var adminList = map[string]bool{
	"b15a4f6c5c1163b5f80715c9bd87d5118ec4b5668cb29f148eeceec61ddeadc2": true,
}

var bankList = map[string]bool{
	"4a246cd2a3f41b2bc1d071d2db159a388cd6f5c3547ea592d401f270073133d7": true,
}

func isBank(uid []byte) bool {
	_, ok := bankList[hex.EncodeToString(uid)]
	return ok
}

func (app *GfcollectionApplication) checkArea(area string) error {
	fmt.Println("checkarea:", area)
	_, value, exists := app.state.Get([]byte(KeyAreaList()))
	if !exists {
		areaList := &AreaList{}
		areaCompany := &AreaCompany{}
		areaCompany.Area = area
		areaList.AreaCompany = append(areaList.AreaCompany, areaCompany)
		save, err := MarshalMessage(areaList)
		if err != nil {
			fmt.Println("3333333")
			return err
		}
		app.state.Set([]byte(KeyAreaList()), save)
		return nil
	}
	areaList := &AreaList{}
	err := UnmarshalMessage(value, areaList)
	if err != nil {
		fmt.Println("44444444")
		return err
	}
	check := false
	for _, v := range areaList.AreaCompany {
		if area == v.Area {
			check = true
			break
		}
	}
	if !check {
		areaCompany := &AreaCompany{}
		areaCompany.Area = area
		areaList.AreaCompany = append(areaList.AreaCompany, areaCompany)
		save, err := MarshalMessage(areaList)
		if err != nil {
			fmt.Println("55555555")
			return err
		}
		app.state.Set([]byte(KeyAreaList()), save)
	}
	return nil
}

func (app *GfcollectionApplication) importCasePool(c *Case, bank string) error {
	_, value, exists := app.state.Get([]byte(KeyAreaCasePool(c.CaseArea, bank)))
	fmt.Println("555", exists, c)
	if !exists {
		var pool AreaCasePool
		pool.BankId = bank
		pool.Area = c.CaseArea
		pool.WaitingList = append(pool.WaitingList, c)
		save, err := MarshalMessage(&pool)
		if err != nil {
			return err
		}
		app.state.Set([]byte(KeyAreaCasePool(c.CaseArea, bank)), save)
		fmt.Println("POOL", pool)
		s, err := MarshalMessage(c)
		if err != nil {
			return err
		}
		fmt.Println("案件编号已保存", c.CaseId)
		app.state.Set([]byte(KeyCase(c.CaseId)), s)
		return nil
	}
	var poo AreaCasePool
	err := UnmarshalMessage(value, &poo)
	if err != nil {
		return err
	}
	_, _, exists = app.state.Get([]byte(KeyCase(c.CaseId)))
	if !exists {
		s, err := MarshalMessage(c)
		if err != nil {
			return err
		}
		fmt.Println("案件编号已保存", c.CaseId)
		app.state.Set([]byte(KeyCase(c.CaseId)), s)

		poo.WaitingList = append(poo.WaitingList, c)
		save, err := MarshalMessage(&poo)
		if err != nil {
			return err
		}
		app.state.Set([]byte(KeyAreaCasePool(c.CaseArea, bank)), save)
	}

	return nil

}

func (app *GfcollectionApplication) reBackCasePool(c *Case) error {
	_, value, exists := app.state.Get([]byte(KeyAreaCasePool(c.CaseArea, c.BankId)))
	if !exists {
		return ErrAreaPoolNotExists
	}
	pool := &AreaCasePool{}
	err := UnmarshalMessage(value, pool)
	if err != nil {
		return err
	}
	pool.WaitingList = append(pool.WaitingList, c)
	save, err := MarshalMessage(pool)
	if err != nil {
		return err
	}
	app.state.Set([]byte(KeyAreaCasePool(c.CaseArea, c.BankId)), save)
	return nil
}

func (app *GfcollectionApplication) autoDeliver() (*DeliverResult, error) {
	fmt.Println("开始分发==========================")
	result := &DeliverResult{}
	_, value, exists := app.state.Get([]byte(KeyAreaList()))
	if !exists {
		return nil, nil
	}
	var arealist AreaList
	err := UnmarshalMessage(value, &arealist)
	if err != nil {
		return nil, err
	}
	_, value, exists = app.state.Get([]byte(KeyBanks()))
	var banks Banks
	err = UnmarshalMessage(value, &banks)
	if err != nil {
		return nil, err
	}
	for _, b := range banks.Banks { //银行
		_, value, exists = app.state.Get([]byte(KeyBank(b)))
		if !exists {
			return nil, ErrBankNotExists
		}
		var bank Bank
		err = UnmarshalMessage(value, &bank)
		if err != nil {
			return nil, err
		}
		restFull := make([]*Case, 0)

		for _, v := range arealist.AreaCompany { //地区
			fmt.Println("area:", v.Area, "bank:", bank.BankId)
			_, value, exists := app.state.Get([]byte(KeyAreaCasePool(v.Area, bank.BankId)))
			if !exists {
				fmt.Println("地区案件池不存在")
				continue
			}
			weight := 0
			for _, vv := range v.Companys {
				weight += int(vv.Weight)
			}
			if weight > 10000 {
				fmt.Println("地区权重异常")
				continue
			}
			var areaPool AreaCasePool
			err = UnmarshalMessage(value, &areaPool)
			if err != nil {
				return nil, err
			}
			number := len(areaPool.WaitingList)
			fmt.Println("待分发数量：", number)
			// tempCase := areaPool.WaitingList[len(areaPool.WaitingList):]
			// restCase := areaPool.WaitingList[len(areaPool.WaitingList):]
			tempCase := make([]*Case, 0)
			restCase := make([]*Case, 0)
			index := make([]string, 0)
			if len(v.Companys) == 0 {
				restFull = append(restFull, areaPool.WaitingList...)
				continue
			}
			//同一个地区案件数 委外机构比小于2    不自动派发
			if len(areaPool.WaitingList)/len(v.Companys) < 2 {
				continue
			}
			//fmt.Println("待分配的案件：", areaPool.WaitingList)
			for kk, vv := range areaPool.WaitingList {
				for _, vvv := range bank.CaseConfList {
					if vvv.IsApply && vv.DebtAmount >= vvv.CaseMinAmount && vv.DebtAmount <= vvv.CaseMaxAmount && vv.OverdueDays == vvv.OverdueDays {
						check := false
						for _, vvvv := range index {
							tempL := strings.Split(vvvv, "_")
							if len(tempL) != 2 {
								panic("exception index")
							}
							tempvv, err := strconv.Atoi(tempL[0])
							if err != nil {
								return nil, err
							}
							if tempvv == kk+1 {
								check = true
								break
							}
						}
						if !check {
							index = append(index, strconv.Itoa(kk+1)+"_"+vvv.CaseConfId)
						}
					}
				}

			}
			fmt.Println("待分发顺序", index)
			number = len(index)
			// shuffleIndex := Shuffle(index)
			// fmt.Println("洗牌顺序", shuffleIndex)
			w := 0
			ccc := make([]string, 0)
			ccc1 := make([]string, 0)
			for _, vv := range v.Companys {
				fmt.Println("公司：", vv.CompanyId)
				if vv.Weight == 0 {
					fmt.Println("公司没有权重，暂时不派发")
					continue
				}
				var deliverCompany DeliverCompany
				deliverCompany.Area = v.Area
				deliverCompany.CompanyId = vv.CompanyId
				deliverCompany.Weight = vv.Weight
				tempIndex := make([]string, 0)
				var company Company
				_, value, exists = app.state.Get([]byte(KeyCompany(vv.CompanyId)))
				if !exists {
					return nil, ErrCompanyNotExists
				}
				err = UnmarshalMessage(value, &company)
				if err != nil {
					return nil, err
				}
				nowNum := len(company.CollectingList) + len(company.DelayedList) + len(company.DeliverList) + len(company.UnDeliverList)
				fmt.Println("公司", vv.CompanyId, " 已接受的案件数量：", nowNum)
				if company.CompanyConf != nil {
					fmt.Println("公司收单规则", company.CompanyConf)
					if int32(nowNum) >= int32(company.CompanyConf.MaxReceive) {
						fmt.Println("公司已派发数量已满足")
						continue
					}
				}
				//获取权重分配案件列表下标
				fmt.Println((number*w)/10000, (number*(int(vv.Weight)+w))/10000)
				for kkk, vvv := range index {
					if kkk+1 > (number*w)/10000 && kkk+1 <= (number*(int(vv.Weight)+w))/10000 {
						//deliverCompany.Indexs = append(deliverCompany.Indexs, vvv)
						tempIndex = append(tempIndex, vvv)
					}
				}
				fmt.Println("公司分发的临时案件列表", tempIndex)
				for kkk, vvv := range areaPool.WaitingList {
					check := false
					if company.CompanyConf != nil && len(company.CompanyId) > 0 {
						if nowNum+len(deliverCompany.CaseIds) >= int(company.CompanyConf.MaxAmount) {
							fmt.Println("---公司已派发数量已满足,停止分发")
							break
						}
					}

					for _, vvvv := range tempIndex {
						tempL := strings.Split(vvvv, "_")
						if len(tempL) != 2 {
							return nil, ErrException
						}
						tempv := tempL[0]
						tempId := tempL[1]
						tempvv, err := strconv.Atoi(tempv)
						if err != nil {
							return nil, err
						}
						if kkk+1 == tempvv {
							//是否设置委外规则
							if company.CompanyConf != nil {
								//是否启用委外规则
								if bank.IsApplyCompanyConf {
									if vvv.OverdueDays == company.CompanyConf.OverdueDays && vvv.DebtAmount >= company.CompanyConf.MinAmount && vvv.DebtAmount <= company.CompanyConf.MaxAmount {
										deliverCompany.CaseIds = append(deliverCompany.CaseIds, vvv.CaseId+"_"+tempId)
										check = true
									}
								} else {
									deliverCompany.CaseIds = append(deliverCompany.CaseIds, vvv.CaseId+"_"+tempId)
									check = true
								}
							} else {
								deliverCompany.CaseIds = append(deliverCompany.CaseIds, vvv.CaseId+"_"+tempId)
								check = true
							}
						}
					}
					if check {
						ccc = append(ccc, vvv.CaseId)
						tempCase = append(tempCase, areaPool.WaitingList[kkk])
						//fmt.Println("#", tempCase)
						ccc1 = append(ccc1, areaPool.WaitingList[kkk].CaseId)
						areaPool.WaitingList[kkk].CompanyId = vv.CompanyId
						//areaPool.DeliverList = append(areaPool.DeliverList, areaPool.WaitingList[kkk])
						company.UnDeliverList = append(company.UnDeliverList, areaPool.WaitingList[kkk])
						_, value, exists := app.state.Get([]byte(KeyCase(vvv.CaseId)))
						fmt.Println("分发的案件编号", vvv.CaseId)
						if !exists {
							return nil, ErrCaseNotExists
						}
						var c Case
						err := UnmarshalMessage(value, &c)
						if err != nil {
							return nil, err
						}
						c.CompanyId = vv.CompanyId
						c.CaseState = CaseState_CS_DELIVER
						save, err := MarshalMessage(&c)
						if err != nil {
							return nil, err
						}
						fmt.Println("Case 详情:", c)
						app.state.Set([]byte(KeyCase(vvv.CaseId)), save)
					}
				}
				fmt.Println("$$$$$$$$$$$$$$$$$$$$", deliverCompany)
				result.DeliverList = append(result.DeliverList, &deliverCompany)

				csave, err := MarshalMessage(&company)
				if err != nil {
					return nil, ErrStorage
				}
				app.state.Set([]byte(KeyCompany(vv.CompanyId)), csave)
				w += int(vv.Weight)
			}
			fmt.Println("############", tempCase)
			fmt.Println("此区域下自动分发ID", ccc)
			mm := make([]string, 0)
			rr := make([]string, 0)
			tt := make([]string, 0)
			for kk, vv := range areaPool.WaitingList {
				mm = append(mm, vv.CaseId)
				check := false
				for _, tvv := range tempCase {
					if vv.CaseId == tvv.CaseId {
						check = true
						break
					}
				}
				if !check {
					restCase = append(restCase, areaPool.WaitingList[kk])
				}
			}
			for _, vv := range restCase {
				rr = append(rr, vv.CaseId)
			}
			fmt.Println("###########2", tempCase)
			for _, vv := range tempCase {
				tt = append(tt, vv.CaseId)
			}
			if len(areaPool.WaitingList) != len(restCase)+len(tempCase) {
				fmt.Println("自动分派异常+++++++++++++++++++++++++++++++++", "案件池数量-", len(areaPool.WaitingList), "未分派数量-", len(restCase), "分派数量-", len(tempCase))
				fmt.Println(mm)
				fmt.Println(rr)
				fmt.Println(ccc)
				fmt.Println(ccc1)
				fmt.Println(tt)
			}
			//fmt.Println(restCase)
			fmt.Println("未分发的案件列表", len(restCase))
			//fmt.Println("已分发的案件列表", areaPool.DeliverList)
			restFull = append(restFull, restCase...)
			areaPool.WaitingList = restCase
			save, err := MarshalMessage(&areaPool)
			if err != nil {
				return nil, err
			}
			app.state.Set([]byte(KeyAreaCasePool(v.Area, bank.BankId)), save)
		}
		fmt.Println("全辖域分派开始=======================================")
		fmt.Println("全辖域分派案件数量，", len(restFull))
		for _, v := range arealist.AreaCompany {
			if v.Area == "-1" {
				mapRestWaiting := make(map[string][]*Case)
				weight := 0
				for _, vv := range v.Companys {
					weight += int(vv.Weight)
				}
				if weight != 10000 {
					fmt.Println("总权重:", weight)
					fmt.Println("==全辖域权重异常==")
					break
				}
				number := len(restFull)
				fmt.Println("全辖域待分配的案件数量：", number)

				index := make([]string, 0)
				if len(v.Companys) == 0 {
					continue
				}

				for kk, vv := range restFull {
					for _, vvv := range bank.CaseConfList {
						if vvv.IsApply && vv.DebtAmount >= vvv.CaseMinAmount && vv.DebtAmount <= vvv.CaseMaxAmount && vv.OverdueDays == vvv.OverdueDays {
							check := false
							for _, vvvv := range index {
								tempL := strings.Split(vvvv, "_")
								if len(tempL) != 2 {
									panic("exception index")
								}
								tempvv, err := strconv.Atoi(tempL[0])
								if err != nil {
									return nil, err
								}
								if tempvv == kk+1 {
									check = true
									break
								}
							}
							if !check {
								index = append(index, strconv.Itoa(kk+1)+"_"+vvv.CaseConfId)
							}

						}
					}
				}
				fmt.Println("待分发顺序", index)
				number = len(index)
				w := 0
				//tempCase := restFull[len(restFull):]
				tempCase := make([]*Case, 0)
				//restCase := restFull[len(restFull):]
				for _, vv := range v.Companys {
					if vv.Weight == 0 || vv.Weight > 10000 {
						fmt.Println("weight:", vv.Weight)
						fmt.Println("==全辖域权重异常==")
						continue
					}
					var deliverCompany DeliverCompany
					deliverCompany.Area = v.Area
					deliverCompany.CompanyId = vv.CompanyId
					deliverCompany.Weight = vv.Weight
					tempIndex := make([]string, 0)
					var company Company
					_, value, exists = app.state.Get([]byte(KeyCompany(vv.CompanyId)))
					if !exists {
						return nil, ErrCompanyNotExists
					}
					err = UnmarshalMessage(value, &company)
					if err != nil {
						return nil, err
					}
					nowNum := len(company.CollectingList) + len(company.DelayedList) + len(company.DeliverList) + len(company.UnDeliverList)
					fmt.Println("全辖域公司", company.CompanyId, "已接受的案件池数量", nowNum)
					if company.CompanyConf != nil {
						if int32(nowNum) >= int32(company.CompanyConf.MaxReceive) {
							fmt.Println("公司已派发数量已满足")
							continue
						}
					}
					//获取权重分配案件列表下标
					fmt.Println((number*w)/10000, (number*(int(vv.Weight)+w))/10000)
					for kkk, vvv := range index {
						if kkk+1 > (number*w)/10000 && kkk+1 <= (number*(int(vv.Weight)+w))/10000 {
							//deliverCompany.Indexs = append(deliverCompany.Indexs, vvv)
							tempIndex = append(tempIndex, vvv)
						}
					}
					fmt.Println("全辖域公司分发的临时案件列表", tempIndex)

					for kkk, vvv := range restFull {
						check := false
						if company.CompanyConf != nil && len(company.CompanyId) > 0 {
							if nowNum+len(deliverCompany.CaseIds) >= int(company.CompanyConf.MaxAmount) {
								fmt.Println("已达到最大接受量")
								break
							}
						}

						for _, vvvv := range tempIndex {
							tempL := strings.Split(vvvv, "_")
							if len(tempL) != 2 {
								return nil, ErrException
							}
							tempv := tempL[0]
							tempId := tempL[1]
							tempvv, err := strconv.Atoi(tempv)
							if err != nil {
								return nil, err
							}
							if kkk+1 == tempvv {
								//是否设置委外规则
								if company.CompanyConf != nil && company.CompanyId != "" {
									//是否启用委外规则
									if bank.IsApplyCompanyConf {
										if vvv.OverdueDays == company.CompanyConf.OverdueDays && vvv.DebtAmount >= company.CompanyConf.MinAmount && vvv.DebtAmount <= company.CompanyConf.MaxAmount {
											deliverCompany.CaseIds = append(deliverCompany.CaseIds, vvv.CaseId+"_"+tempId)
											check = true
										}
									} else {
										deliverCompany.CaseIds = append(deliverCompany.CaseIds, vvv.CaseId+"_"+tempId)
										check = true
									}
								} else {
									deliverCompany.CaseIds = append(deliverCompany.CaseIds, vvv.CaseId+"_"+tempId)
									check = true
								}
							}
						}
						if check {
							tempCase = append(tempCase, restFull[kkk])
							restFull[kkk].CompanyId = vv.CompanyId
							//areaPool.DeliverList = append(areaPool.DeliverList, restFull[kkk])
							company.UnDeliverList = append(company.UnDeliverList, restFull[kkk])
							mcheck := false
							for mk, _ := range mapRestWaiting {
								if mk == restFull[kkk].CaseArea {
									mcheck = true
								}
							}
							if !mcheck {
								mapRestWaiting[restFull[kkk].CaseArea] = make([]*Case, 0)
							}
							mapRestWaiting[restFull[kkk].CaseArea] = append(mapRestWaiting[restFull[kkk].CaseArea], restFull[kkk])
							_, value, exists := app.state.Get([]byte(KeyCase(vvv.CaseId)))
							fmt.Println("分发的案件编号", vvv.CaseId)
							if !exists {
								return nil, ErrCaseNotExists
							}
							var c Case
							err := UnmarshalMessage(value, &c)
							if err != nil {
								return nil, err
							}
							c.CompanyId = vv.CompanyId
							c.CaseState = CaseState_CS_DELIVER
							save, err := MarshalMessage(&c)
							if err != nil {
								return nil, err
							}
							fmt.Println("Case 详情:", c)
							app.state.Set([]byte(KeyCase(vvv.CaseId)), save)
						}
					}
					fmt.Println("&&&&&&&&&&&&&&&&&&&&", deliverCompany)
					result.DeliverList = append(result.DeliverList, &deliverCompany)
					csave, err := MarshalMessage(&company)
					if err != nil {
						return nil, ErrStorage
					}
					app.state.Set([]byte(KeyCompany(vv.CompanyId)), csave)
					w += int(vv.Weight)
				}
				for mk, mv := range mapRestWaiting {

					_, value, exists := app.state.Get([]byte(KeyAreaCasePool(mk, bank.BankId)))
					if !exists {
						fmt.Println(mk, "案件池不存在")
						continue
					}
					var pool AreaCasePool
					err := UnmarshalMessage(value, &pool)
					if err != nil {
						return nil, ErrStorage
					}
					mrest := make([]*Case, 0)

					for pk, pv := range pool.WaitingList {
						mcheck := false
						for _, mvv := range mv {
							if pv.CaseId == mvv.CaseId {
								mcheck = true
								break
							}
						}
						if !mcheck {
							mrest = append(mrest, pool.WaitingList[pk])
						}
					}
					pool.WaitingList = mrest
					pool.DeliverList = append(pool.DeliverList, mv...)
					save, err := MarshalMessage(&pool)
					if err != nil {
						return nil, ErrStorage
					}
					app.state.Set([]byte(KeyAreaCasePool(mk, bank.BankId)), save)
				}
			}
		}
	}
	fmt.Println("=============================自动打包结果===============================")
	fmt.Println(result)
	fmt.Println("=============================自动打包结果===============================")
	return result, nil
}

func (app *GfcollectionApplication) saveReceipt(instructionId int64, receipt *Receipt) error {
	save, err := MarshalMessage(receipt)
	if err != nil {
		return err
	}
	app.state.Set(reqkey(instructionId), save)
	return err
}

func (app *GfcollectionApplication) checkInstructionId(instructionId int64) error {
	key := reqkey(instructionId)
	_, _, exists := app.state.Get(key)
	if !exists {
		return nil
	}
	return ErrDupInstructionId
}

func reqkey(n int64) []byte {
	s := fmt.Sprintf("reqkey_%d", n)
	return []byte(s)
}

func KeyAreaList() string {
	return "arealist"
}

func KeyBanks() string {
	return "banks"
}

func KeyAreaCasePool(area string, bank string) string {
	fmt.Println("pool:" + fmt.Sprintf("%s", area) + fmt.Sprintf("%s", bank))
	return "pool:" + fmt.Sprintf("%s", area) + fmt.Sprintf("%s", bank)
}

func KeyAdmins() string {
	return "administrators"
}

func KeyUser(id string) string {
	return "user:" + fmt.Sprintf("%s", id)
}

func KeyCompany(id string) string {
	return "company:" + fmt.Sprintf("%s", id)
}

func KeyLoan(id string) string {
	return "loan:" + fmt.Sprintf("%s", id)
}

func KeyPackConf(id string) string {
	return "packconf:" + fmt.Sprintf("%s", id)
}

func KeyCase(id string) string {
	return "case:" + fmt.Sprintf("%s", id)
}

func KeyAreaBool(area string) string {
	return "case:" + fmt.Sprintf("%s", area)
}

func KeyCaseConf(id string) string {
	return "caseconf:" + fmt.Sprintf("%s", id)
}

func KeyPackConfs() string {
	return "packconfs"
}

func KeyPackId() string {
	return "packid_"
}

func KeyBank(bank string) string {
	return "bank:" + fmt.Sprintf("%s", bank)
}

func KeyPackages(id string) string {
	return "pack:" + fmt.Sprintf("%s", id)
}

func KeyPlatform() string {
	return "platform"
}

func isOriginalAdmin(str []byte) bool {
	_, ok := adminList[hex.EncodeToString(str)]
	return ok
}

func CheckSign(data []byte, uid []byte, sign []byte) error {
	c, err := crypto.New("ed25519")
	if err != nil {
		return err
	}
	if len(sign) != 64 {
		return ErrSignWrongLength
	}
	sig, err := c.SignatureFromBytes(sign)
	if err != nil {
		return err
	}
	pub, err := c.PubKeyFromBytes(uid[:])
	if err != nil {
		return err
	}
	if !pub.VerifyBytes(data, sig) {
		return ErrSign
	}
	return nil
}

func Signdata(privKey []byte, data []byte) []byte {
	c, err := crypto.New("ed25519")
	if err != nil {
		panic(err)
	}
	priv, err := c.PrivKeyFromBytes(privKey)
	if err != nil {
		panic(err)
	}
	sig := priv.Sign(data)
	return sig.Bytes()
}

func generateRandomNumber(start int, end int, count int) []int {
	//范围检查
	if end < start || (end-start) < count {
		return nil
	}
	//存放结果的slice
	nums := make([]int, 0)
	numsort := make([]int, 0)
	sum := 0
	//随机数生成器，加入时间戳保证每次生成的随机数不一样
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for len(numsort) < count {
		//生成随机数
		num := r.Intn((end - start)) + start
		sum++
		//查重
		// mid := binarySearchIndex(numsort, num)
		// if mid == -4 {
		// 	numsort = append(numsort, num)
		// 	nums = append(nums, num)
		// } else if mid == -3 {
		// 	temp := make([]int, 0)
		// 	temp = append(temp, num)
		// 	temp = append(temp, numsort[:]...)
		// 	numsort = temp
		// 	nums = append(nums, num)
		// } else if mid == -2 {
		// 	numsort = append(numsort, num)
		// 	nums = append(nums, num)
		// } else if mid == -1 {
		// } else {
		// 	temp := make([]int, 0)
		// 	temp = append(temp, numsort[:mid+1]...)
		// 	temp = append(temp, num)
		// 	//fmt.Println(numsort[mid+1:])
		// 	temp = append(temp, numsort[mid+1:]...)
		// 	numsort = temp
		// 	nums = append(nums, num)
		// }
		exist := false
		for _, v := range nums {
			if v == num {
				exist = true
				break
			}
		}
		if !exist {
			nums = append(nums, num)
			numsort = append(numsort, num)
		}
		//fmt.Println("位置下标", mid)

		//fmt.Println("有序列表", numsort)
		//fmt.Println("无序列表", nums)

	}
	//fmt.Println("有序列表", numsort)
	//fmt.Println("无序列表", nums)
	fmt.Println("运行了", sum, "次随机数")
	return nums
}

func binarySearch(sortedList []int, lookingFor int) int {
	var lo int = 0
	var hi int = len(sortedList) - 1
	for lo <= hi {
		var mid int = lo + (hi-lo)/2
		var midValue int = sortedList[mid]
		if midValue == lookingFor {
			return midValue
		} else if midValue > lookingFor {
			hi = mid - 1
		} else {
			lo = mid + 1
		}
	}
	return -1
}

func binarySearchIndex(sortedList []int, lookingFor int) int {
	if len(sortedList) == 0 {
		return -4
	}
	var lo int = 0
	var hi int = len(sortedList) - 1
	if lookingFor < sortedList[lo] {
		return -3
	} else if lookingFor > sortedList[hi] {
		return -2
	}
	for lo <= hi {
		var mid int = lo + (hi-lo)/2
		var midValue int = sortedList[mid]
		if midValue == lookingFor {
			return -1
		} else if midValue > lookingFor {
			hi = mid - 1
		} else {
			if sortedList[mid+1] < lookingFor {
				lo = mid + 1
			} else if sortedList[mid+1] > lookingFor {
				return mid
			} else {
				return -1
			}

		}
	}
	return -1
}

func BubbleSort(values []int) {
	flag := true
	for i := 0; i < len(values)-1; i++ {
		flag = true
		for j := 0; j < len(values)-i-1; j++ {
			if values[j] > values[j+1] {
				values[j], values[j+1] = values[j+1], values[j]
				flag = false
			}
		}
		if flag == true {
			// 如果已经顺序对了，就不用继续冒泡排序了。
			break
		}
	}
}

func quickSort(values []int, left, right int) {

	temp := values[left]
	p := left
	i, j := left, right

	for i <= j {
		for j >= p && values[j] >= temp {
			j--
		}
		if j >= p {
			values[p] = values[j]
			p = j
		}

		for i <= p && values[i] <= temp {
			i++
		}
		if i <= p {
			values[p] = values[i]
			p = i
		}

	}

	values[p] = temp

	if p-left > 1 {
		quickSort(values, left, p-1)
	}
	if right-p > 1 {
		quickSort(values, p+1, right)
	}

}

//插入排序（排序10000个整数，用时约30ms）
func insertSort(nums []int) {
	for i := 1; i < len(nums); i++ {
		if nums[i] < nums[i-1] {
			j := i - 1
			temp := nums[i]
			for j >= 0 && nums[j] > temp {
				nums[j+1] = nums[j]
				j--
			}
			nums[j+1] = temp
		}
	}
}

func Shuffle(vals []int) []int {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	ret := make([]int, len(vals))
	perm := r.Perm(len(vals))
	for i, index := range perm {
		ret[i] = vals[index]
	}
	return ret
}
