package gfcollection

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
)

func (app *GfcollectionApplication) checkInitPlatform(pubkey []byte, platformPubkey []byte) error {
	if !isOriginalAdmin(pubkey) {
		return ErrNotAdmin
	}
	_, _, exists := app.state.Get([]byte(KeyPlatform()))
	if exists {
		return ErrPlatformExist
	}
	return nil
}

func (app *GfcollectionApplication) initPlatform(pubkey []byte, platformPubkey []byte, info string, instructionId int64) (*Response, error) {
	fmt.Println("initPlatform==")
	err := app.checkInitPlatform(pubkey, platformPubkey)
	if err != nil {
		return nil, err
	}
	platform := &Platform{}
	platform.Pubkey = platformPubkey[:]
	platform.Info = info

	save, err := MarshalMessage(platform)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyPlatform()), save)
	event := &EventInitPlatform{}
	event.PlatformKey = platformPubkey
	return &Response{Value: &Response_InitPlatform{&ResponseInitPlatform{InstructionId: instructionId, Event: &Event{Value: &Event_InitPlatform{event}}}}}, nil
}

func (app *GfcollectionApplication) checkSetBank(pubkey []byte, bankId string, bankPubkey []byte) error {
	_, _, exists := app.state.Get([]byte(KeyBank(bankId)))
	if exists {
		return ErrBankExists
	}
	_, _, exists = app.state.Get([]byte(KeyPlatform()))
	if !exists {
		return ErrPlatformNotExist
	}
	return nil
}

func (app *GfcollectionApplication) setBank(pubkey []byte, bankId string, bankName string, bankPubkey []byte, instructionId int64) (*Response, error) {
	fmt.Println("SETBANK==")
	bank := &Bank{}
	bank.Pubkey = bankPubkey[:]
	bank.BankId = bankId
	bank.BankName = bankName
	save, err := MarshalMessage(bank)
	if err != nil {
		return nil, err
	}
	fmt.Println("SETBANK", bank)
	app.state.Set([]byte(KeyBank(bankId)), save)
	_, value, exists := app.state.Get([]byte(KeyBanks()))
	if !exists {
		banks := &Banks{}
		banks.Banks = append(banks.Banks, bankId)
		save, err := MarshalMessage(banks)
		if err != nil {
			return nil, err
		}
		app.state.Set([]byte(KeyBanks()), save)
	}
	var banks Banks
	err = UnmarshalMessage(value, &banks)
	if err != nil {
		return nil, err
	}
	banks.Banks = append(banks.Banks, bankId)
	fmt.Println("BANKS :", banks)
	save, err = MarshalMessage(&banks)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyBanks()), save)
	fmt.Println("BANKS", banks)
	event := &EventSetBank{}
	event.BankId = bankId
	return &Response{Value: &Response_SetBank{&ResponseSetBank{InstructionId: instructionId, Event: &Event{Value: &Event_SetBank{event}}}}}, nil
}

func (app *GfcollectionApplication) checkSetCompany(pubkey []byte, companyId, companyName string, companyPubkey []byte) error {
	_, _, exists := app.state.Get([]byte(KeyCompany(companyId)))
	if exists {
		return ErrCompanyExists
	}
	_, value, exists := app.state.Get([]byte(KeyPlatform()))
	if !exists {
		return ErrPlatformNotExist
	}
	var plat Platform
	err := UnmarshalMessage(value, &plat)
	if err != nil {
		return ErrStorage
	}
	if !bytes.Equal(pubkey, plat.Pubkey) {
		return ErrNoRight
	}
	return nil
}

func (app *GfcollectionApplication) setCompany(pubkey []byte, companyId, companyName string, companyPubkey []byte, instructionId int64) (*Response, error) {
	fmt.Println("setCompany==")
	company := &Company{}
	company.CompanyId = companyId
	company.CompanyName = companyName
	company.CompanyPubkey = companyPubkey[:]
	save, err := MarshalMessage(company)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyCompany(companyId)), save)
	fmt.Println("setcompany", company)
	event := &EventSetCompany{}
	event.CompanyId = companyId
	return &Response{Value: &Response_SetCompany{&ResponseSetCompany{InstructionId: instructionId, Event: &Event{Value: &Event_SetCompany{event}}}}}, nil
}

func (app *GfcollectionApplication) checkAddCompany(pubkey []byte, uid string, companyId string, companyPubkey []byte, weight int32) error {
	_, value, exists := app.state.Get([]byte(KeyBank(uid)))
	if !exists {
		return ErrBankNotExists
	}
	var bank Bank
	err := UnmarshalMessage(value, &bank)
	if err != nil {
		return ErrStorage
	}
	if bank.BankId != uid {
		return ErrNoRight
	}
	if weight > int32(10000) || weight < int32(1) {
		return ErrWrongWeight
	}
	_, _, exists = app.state.Get([]byte(KeyCompany(companyId)))
	if exists {
		return ErrCompanyExists
	}

	return nil
}

func (app *GfcollectionApplication) addCompany(pubkey []byte, uid string, companyId, companyName, companyArea string, companyPubkey []byte, weight int32, instructionId int64) (*Response, error) {
	fmt.Println("addCompany==")
	_, value, exists := app.state.Get([]byte(KeyBank(uid)))
	if !exists {
		return nil, ErrBankNotExists
	}
	var bank Bank
	err := UnmarshalMessage(value, &bank)
	if err != nil {
		return nil, ErrStorage
	}
	err = app.checkArea(companyArea)
	if err != nil {
		return nil, err
	}
	fmt.Println("company pubkey ", companyPubkey)
	fmt.Println("company pubkey ", companyPubkey[:])
	company := &Company{}
	company.CompanyId = companyId
	company.CompanyName = companyName
	company.CompanyArea = companyArea
	company.CompanyPubkey = companyPubkey[:]

	save, err := MarshalMessage(company)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyCompany(companyId)), save)
	if !bank.IsCompanyOk {
		bank.IsCompanyOk = true
		save, err := MarshalMessage(&bank)
		if err != nil {
			return nil, err
		}
		app.state.Set([]byte(KeyBank(uid)), save)
	}
	fmt.Println("addcompany", company)

	_, value, exists = app.state.Get([]byte(KeyAreaList()))
	if !exists {
		fmt.Println("不存在  新建arealist")
		return nil, ErrAreaListNotExist
	}
	areaList := &AreaList{}
	err = UnmarshalMessage(value, areaList)
	if err != nil {
		return nil, err
	}
	fmt.Println("AREALIST COMPANY BEFORE", areaList.AreaCompany)
	check := false
	for k, v := range areaList.AreaCompany {
		if companyArea == v.Area {
			check = true
			ch := false
			for _, vv := range v.Companys {
				if vv.CompanyId == companyId {
					ch = true
					break
				}
			}
			if !ch {
				var weightCompany WeightCompany
				weightCompany.CompanyId = companyId
				weightCompany.FirstWeight = weight
				_, _, exists = app.state.Get([]byte(KeyAreaBool(companyArea)))
				//未分派算法
				if !exists {
					if len(v.Companys) == 0 {
						weightCompany.Weight = int32(10000)
					} else {
						total := int32(0)
						totalFirst := int32(0)
						for _, vvv := range v.Companys {
							totalFirst += vvv.FirstWeight
						}
						for kkk, vvv := range v.Companys {
							ff := fmt.Sprintf("%0.f", float64(int(vvv.FirstWeight))*10000/float64(int(totalFirst+weight)))
							w, _ := strconv.Atoi(ff)
							areaList.AreaCompany[k].Companys[kkk].Weight = int32(w)
							total += int32(w)
						}

						weightCompany.Weight = int32(10000) - total
					}

				} else {
					//已分派算法
					weightCompany.Weight = weight
					total := int32(0)
					l := len(v.Companys)
					for kkk, vvv := range v.Companys {
						if kkk+1 < l {
							ff := fmt.Sprintf("%0.f", (float64(int(vvv.Weight))*float64(10000-int(weight)))/float64(10000))
							w, _ := strconv.Atoi(ff)
							areaList.AreaCompany[k].Companys[kkk].Weight = int32(w)
							total += int32(w)
						} else {
							areaList.AreaCompany[k].Companys[kkk].Weight = int32(10000) - weight - total
						}

					}
				}

				areaList.AreaCompany[k].Companys = append(areaList.AreaCompany[k].Companys, &weightCompany)
			}
			break
		}
	}
	fmt.Println("AREALIST COMPANY AFTER", areaList.AreaCompany)
	if !check {
		return nil, ErrException
	}
	save, err = MarshalMessage(areaList)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyAreaList()), save)
	fmt.Println("AREALIST COMPANY", areaList.AreaCompany)
	fmt.Println("arealist:", areaList)
	event := &EventAddCompany{}
	event.CompanyId = companyId
	return &Response{Value: &Response_AddCompany{&ResponseAddCompany{InstructionId: instructionId, Event: &Event{Value: &Event_AddCompany{event}}}}}, nil
	// fmt.Println("addCompany==")
	// _, value, exists := app.state.Get([]byte(KeyBank(uid)))
	// if !exists {
	// 	return nil, ErrBankNotExists
	// }
	// var bank Bank
	// err := UnmarshalMessage(value, &bank)
	// if err != nil {
	// 	return nil, ErrStorage
	// }
	// err = app.checkArea(companyArea)
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Println("company pubkey ", companyPubkey)
	// fmt.Println("company pubkey ", companyPubkey[:])
	// company := &Company{}
	// company.CompanyId = companyId
	// company.CompanyName = companyName
	// company.CompanyArea = companyArea
	// company.CompanyPubkey = companyPubkey[:]

	// save, err := MarshalMessage(company)
	// if err != nil {
	// 	return nil, err
	// }
	// app.state.Set([]byte(KeyCompany(companyId)), save)
	// if !bank.IsCompanyOk {
	// 	bank.IsCompanyOk = true
	// 	save, err := MarshalMessage(&bank)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	app.state.Set([]byte(KeyBank(uid)), save)
	// }
	// fmt.Println("addcompany", company)

	// _, value, exists = app.state.Get([]byte(KeyAreaList()))
	// if !exists {
	// 	fmt.Println("不存在  新建arealist")
	// 	return nil, ErrAreaListNotExist
	// }
	// areaList := &AreaList{}
	// err = UnmarshalMessage(value, areaList)
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Println("AREALIST COMPANY BEFORE", areaList.AreaCompany)
	// check := false
	// for k, v := range areaList.AreaCompany {
	// 	if companyArea == v.Area {
	// 		check = true
	// 		ch := false
	// 		for _, vv := range v.Companys {
	// 			if vv.CompanyId == companyId {
	// 				ch = true
	// 				break
	// 			}
	// 		}
	// 		if !ch {
	// 			var weightCompany WeightCompany
	// 			weightCompany.CompanyId = companyId
	// 			if len(v.Companys) == 0 {
	// 				//weightCompany.FirstWeight = weight
	// 				weightCompany.Weight = 10000
	// 				// }
	// 				// else if len(v.Companys) == 1 && v.Companys[0].FirstWeight > 0 {
	// 				// 	weightCompany.Weight = (weight * 10000) / (v.Companys[0].FirstWeight + weight)
	// 				// 	areaList.AreaCompany[k].Companys[0].Weight = (v.Companys[0].FirstWeight * 10000) / (v.Companys[0].FirstWeight + weight)
	// 			} else {
	// 				total := int32(0)
	// 				for kkk, vvv := range v.Companys {
	// 					//var w int32
	// 					ff := fmt.Sprintf("%0.f", float64(int(vvv.Weight))*10000/float64(10000+int(weight)))
	// 					w, _ := strconv.Atoi(ff)
	// 					areaList.AreaCompany[k].Companys[kkk].Weight = int32(w)
	// 					total += int32(w)
	// 				}
	// 				weightCompany.Weight = int32(10000) - total
	// 			}

	// 			areaList.AreaCompany[k].Companys = append(areaList.AreaCompany[k].Companys, &weightCompany)
	// 		}
	// 		break
	// 	}
	// }
	// fmt.Println("AREALIST COMPANY AFTER", areaList.AreaCompany)
	// if !check {
	// 	return nil, ErrException
	// }
	// save, err = MarshalMessage(areaList)
	// if err != nil {
	// 	return nil, err
	// }
	// app.state.Set([]byte(KeyAreaList()), save)
	// fmt.Println("AREALIST COMPANY", areaList.AreaCompany)
	// fmt.Println("arealist:", areaList)
	// event := &EventAddCompany{}
	// event.CompanyId = companyId
	// return &Response{Value: &Response_AddCompany{&ResponseAddCompany{InstructionId: instructionId, Event: &Event{Value: &Event_AddCompany{event}}}}}, nil
}

func (app *GfcollectionApplication) checkAddCaseConf(pubkey []byte, uid string, caseConfId string) error {

	_, value, exists := app.state.Get([]byte(KeyBank(uid)))
	if !exists {
		return ErrBankNotExists
	}
	var bank Bank
	err := UnmarshalMessage(value, &bank)
	if err != nil {
		return ErrStorage
	}
	if bank.BankId != uid {
		return ErrNoRight
	}
	for _, v := range bank.CaseConfList {
		if v.CaseConfId == caseConfId {
			return ErrCaseConfExists
		}
	}
	return nil
}

func (app *GfcollectionApplication) addCaseConf(pubkey []byte, uid string, caseConfId string, caseMinAmount int64, caseMaxAmount int64, expireDays int32, overdueDays int32, rate int32, instructionId int64) (*Response, error) {
	fmt.Println("addCaseConf==")
	_, value, exists := app.state.Get([]byte(KeyBank(uid)))
	if !exists {
		return nil, ErrBankNotExists
	}
	var bank Bank
	err := UnmarshalMessage(value, &bank)
	if err != nil {
		return nil, ErrStorage
	}
	caseConf := &CaseConf{}
	caseConf.BankId = uid
	caseConf.CaseConfId = caseConfId
	caseConf.CaseMinAmount = caseMinAmount
	caseConf.CaseMaxAmount = caseMaxAmount
	caseConf.ExpireDays = expireDays
	caseConf.OverdueDays = overdueDays
	caseConf.Rate = rate
	bank.CaseConfList = append(bank.CaseConfList, caseConf)
	save, err := MarshalMessage(&bank)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyBank(uid)), save)
	fmt.Println("addCaseConf", bank)
	event := &EventAddCaseConf{}
	event.CaseConfId = caseConfId
	return &Response{Value: &Response_AddCaseConf{&ResponseAddCaseConf{InstructionId: instructionId, Event: &Event{Value: &Event_AddCaseConf{event}}}}}, nil
}

func (app *GfcollectionApplication) checkEditCaseConf(pubkey []byte, uid string, caseConfId string) error {

	_, value, exists := app.state.Get([]byte(KeyBank(uid)))
	if !exists {
		return ErrBankNotExists
	}
	fmt.Println("checkEditCaseConf BANK", value)
	var bank Bank
	err := UnmarshalMessage(value, &bank)
	if err != nil {
		return ErrStorage
	}
	if bank.BankId != uid {
		return ErrNoRight
	}
	check := false
	for _, v := range bank.CaseConfList {
		if v.CaseConfId == caseConfId {
			check = true
			break
		}
	}
	if !check {
		return ErrCaseConfNotExists
	}
	return nil
}

func (app *GfcollectionApplication) editCaseConf(pubkey []byte, uid string, caseConfId string, caseMinAmount int64, caseMaxAmount int64, expireDays int32, overdueDays int32, rate int32, instructionId int64) (*Response, error) {
	fmt.Println("editCaseConf==")
	_, value, exists := app.state.Get([]byte(KeyBank(uid)))
	if !exists {
		return nil, ErrBankNotExists
	}
	var bank Bank
	err := UnmarshalMessage(value, &bank)
	if err != nil {
		return nil, ErrStorage
	}
	fmt.Println("editCaseConf:before", bank)
	check := false
	for i, v := range bank.CaseConfList {
		if v.CaseConfId == caseConfId {
			check = true
			bank.CaseConfList[i].CaseMinAmount = caseMinAmount
			bank.CaseConfList[i].CaseMaxAmount = caseMaxAmount
			bank.CaseConfList[i].ExpireDays = expireDays
			bank.CaseConfList[i].OverdueDays = overdueDays
			bank.CaseConfList[i].Rate = rate
			break
		}
	}
	if !check {
		return nil, ErrCaseConfNotExists
	}
	fmt.Println("editCaseConf:after", bank)
	save, err := MarshalMessage(&bank)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyBank(uid)), save)

	event := &EventEditCaseConf{}
	event.CaseConfId = caseConfId
	return &Response{Value: &Response_EditCaseConf{&ResponseEditCaseConf{InstructionId: instructionId, Event: &Event{Value: &Event_EditCaseConf{event}}}}}, nil
}

func (app *GfcollectionApplication) checkDelCaseConf(pubkey []byte, uid string, caseConfId string) error {
	_, value, exists := app.state.Get([]byte(KeyBank(uid)))
	if !exists {
		return ErrBankNotExists
	}
	var bank Bank
	err := UnmarshalMessage(value, &bank)
	if err != nil {
		return ErrStorage
	}
	if bank.BankId != uid {
		return ErrNoRight
	}
	check := false
	for _, v := range bank.CaseConfList {
		if v.CaseConfId == caseConfId {
			check = true
			break
		}
	}
	if !check {
		return ErrCaseConfNotExists
	}
	return nil
}

func (app *GfcollectionApplication) delCaseConf(pubkey []byte, uid string, caseConfId string, instructionId int64) (*Response, error) {
	fmt.Println("delCaseConf==")
	_, value, exists := app.state.Get([]byte(KeyBank(uid)))
	if !exists {
		return nil, ErrBankNotExists
	}
	var bank Bank
	err := UnmarshalMessage(value, &bank)
	if err != nil {
		return nil, ErrStorage
	}
	fmt.Println("editCaseConf:before", bank)
	check := false
	for k, v := range bank.CaseConfList {
		if v.CaseConfId == caseConfId {
			check = true
			temp := bank.CaseConfList[:]
			bank.CaseConfList = temp[:k]
			bank.CaseConfList = append(bank.CaseConfList, temp[k+1:]...)
			break
		}
	}
	if !check {
		return nil, ErrCaseConfNotExists
	}
	fmt.Println("editCaseConf:after", bank)
	save, err := MarshalMessage(&bank)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyBank(uid)), save)
	event := &EventDelCaseConf{}
	event.CaseConfId = caseConfId
	return &Response{Value: &Response_DelCaseConf{&ResponseDelCaseConf{InstructionId: instructionId, Event: &Event{Value: &Event_DelCaseConf{event}}}}}, nil
}

func (app *GfcollectionApplication) checkApplyCaseConf(pubkey []byte, uid string, caseConfIds []string) error {

	_, value, exists := app.state.Get([]byte(KeyBank(uid)))
	if !exists {
		return ErrBankNotExists
	}
	var bank Bank
	err := UnmarshalMessage(value, &bank)
	if err != nil {
		return ErrStorage
	}
	if bank.BankId != uid {
		return ErrNoRight
	}

	for _, vv := range caseConfIds {
		check := false
		for _, v := range bank.CaseConfList {
			if v.CaseConfId == vv {
				check = true
				break
			}
		}
		if !check {
			return ErrCaseConfNotExists
		}
	}

	return nil
}

func (app *GfcollectionApplication) applyCaseConf(pubkey []byte, uid string, caseConfIds []string, isApply bool, instructionId int64) (*Response, error) {
	fmt.Println("applyCaseConf==")
	_, value, exists := app.state.Get([]byte(KeyBank(uid)))
	if !exists {
		return nil, ErrBankNotExists
	}
	var bank Bank
	err := UnmarshalMessage(value, &bank)
	if err != nil {
		return nil, ErrStorage
	}
	fmt.Println("editCaseConf:before", bank)
	bank.IsApplyCompanyConf = isApply
	bank.IsBankConfOk = true
	for _, vv := range caseConfIds {
		check := false
		for k, v := range bank.CaseConfList {
			if v.CaseConfId == vv {
				check = true
				bank.CaseConfList[k].IsApply = true
				break
			}
		}
		if !check {
			return nil, ErrCaseConfNotExists
		}
	}
	fmt.Println("editCaseConf:after", bank)
	save, err := MarshalMessage(&bank)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyBank(uid)), save)

	event := &EventApplyCaseConf{}
	event.CaseConfIds = caseConfIds
	return &Response{Value: &Response_ApplyCaseConf{&ResponseApplyCaseConf{InstructionId: instructionId, Event: &Event{Value: &Event_ApplyCaseConf{event}}}}}, nil
}

func (app *GfcollectionApplication) checkImportCase(pubkey []byte, uid string, caseId string) error {
	_, value, exists := app.state.Get([]byte(KeyBank(uid)))
	if !exists {
		return ErrBankNotExists
	}
	var bank Bank
	err := UnmarshalMessage(value, &bank)
	if err != nil {
		return ErrStorage
	}
	if bank.BankId != uid {
		return ErrNoRight
	}
	_, _, exists = app.state.Get([]byte(KeyCase(caseId)))
	if exists {
		return ErrCaseExists
	}
	return nil
}

func (app *GfcollectionApplication) importCase(pubkey []byte, uid string, bankCard, caseArea, caseId, caseIdCard, caseOwner, contract string, debtAmount, fees, originalAmount int64, overdueDays int32, instructionId int64) (*Response, error) {
	fmt.Println("importCase==")
	_, value, exists := app.state.Get([]byte(KeyBank(uid)))
	if !exists {
		return nil, ErrBankNotExists
	}
	var bank Bank
	err := UnmarshalMessage(value, &bank)
	if err != nil {
		return nil, ErrStorage
	}
	err = app.checkArea(caseArea)
	if err != nil {
		return nil, err
	}

	c := &Case{}
	c.BankCard = bankCard
	c.CaseArea = caseArea
	c.CaseId = caseId
	c.CaseOwner = caseOwner
	c.Contract = contract
	c.DebtAmount = debtAmount
	c.Fees = fees
	c.OriginalAmount = originalAmount
	c.OverdueDays = overdueDays
	err = app.importCasePool(c, bank.BankId)
	if err != nil {
		return nil, err
	}
	//bank.WaitingList = append(bank.WaitingList, c)
	save, err := MarshalMessage(c)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyCase(caseId)), save)

	result, err := app.autoDeliver()
	if err != nil {
		return nil, err
	}

	// save, err = MarshalMessage(&bank)
	// if err != nil {
	// 	return nil, err
	// }
	// app.state.Set([]byte(KeyBank(pubkey)), save)

	event := &EventImportCase{}
	event.CaseId = caseId
	event.Result = result
	return &Response{Value: &Response_ImportCase{&ResponseImportCase{InstructionId: instructionId, Event: &Event{Value: &Event_ImportCase{event}}}}}, nil
}

func (app *GfcollectionApplication) checkImportCaseList(pubkey []byte, uid string, caseList []*Case) error {
	_, value, exists := app.state.Get([]byte(KeyBank(uid)))
	if !exists {
		return ErrBankNotExists
	}
	var bank Bank
	err := UnmarshalMessage(value, &bank)
	if err != nil {
		fmt.Println("111111111")
		return ErrStorage
	}
	if bank.BankId != uid {
		return ErrNoRight
	}
	return nil
}

func (app *GfcollectionApplication) importCaseList(pubkey []byte, uid string, caseList []*Case, instructionId int64) (*Response, error) {
	fmt.Println("importCaseList==")
	_, value, exists := app.state.Get([]byte(KeyBank(uid)))
	if !exists {
		return nil, ErrBankNotExists
	}
	var bank Bank
	err := UnmarshalMessage(value, &bank)
	if err != nil {
		fmt.Println("22222222")
		return nil, ErrStorage
	}
	for _, v := range caseList {
		err = app.checkArea(v.CaseArea)
		if err != nil {
			return nil, err
		}
		// _, value, exists := app.state.Get([]byte(KeyCase(v.CaseId)))

		var c Case
		c.BankCard = v.BankCard
		c.CaseArea = v.CaseArea
		c.CaseId = v.CaseId
		c.CaseOwner = v.CaseOwner
		c.Contract = v.Contract
		c.DebtAmount = v.DebtAmount
		c.Fees = v.Fees
		c.OriginalAmount = v.OriginalAmount
		c.OverdueDays = v.OverdueDays
		c.BankId = uid
		c.CaseState = v.CaseState
		c.CompanyId = v.CompanyId
		// if !exists {
		// 	save, err := MarshalMessage(&c)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	app.state.Set([]byte(KeyCase(v.CaseId)), save)
		// }
		fmt.Println("导入的案件", c)
		err = app.importCasePool(&c, bank.BankId)
		if err != nil {
			fmt.Println("99999999")
			fmt.Println("99999999", err)
			return nil, err
		}
		fmt.Println("no error 1")
		// save, err := MarshalMessage(&c)
		// if err != nil {
		// 	fmt.Println("100000000")
		// 	return nil, err
		// }
		// app.state.Set([]byte(KeyCase(v.CaseId)), save)
	}

	result, err := app.autoDeliver()
	if err != nil {
		return nil, err
	}

	event := &EventImportCaseList{}
	event.CaseList = caseList
	event.Result = result
	return &Response{Value: &Response_ImportCaseList{&ResponseImportCaseList{InstructionId: instructionId, Event: &Event{Value: &Event_ImportCaseList{event}}}}}, nil
}

func (app *GfcollectionApplication) checkAddCompanyConf(pubkey []byte, uid string, companyConfId string) error {
	_, value, exists := app.state.Get([]byte(KeyCompany(uid)))

	if !exists {
		return ErrCompanyNotExists
	}
	var company Company
	err := UnmarshalMessage(value, &company)
	if err != nil {
		return err
	}
	fmt.Println("COMPANY", company)
	fmt.Println(pubkey, company.CompanyPubkey)
	if !bytes.Equal(pubkey, company.CompanyPubkey) {
		return ErrNoRight
	}
	fmt.Println("cid:", company.CompanyConf)
	if company.CompanyConf != nil && company.CompanyConf.CompanyConfId != "" {
		return ErrCompanyConfExists
	}
	return nil
}

func (app *GfcollectionApplication) addCompanyConf(pubkey []byte, uid, companyConfId, companyConfName string, isAutoAdd bool, maxAmount, maxReceive, minAmount int64, overdueDays int32, rate int32, instructionId int64) (*Response, error) {
	fmt.Println("addCompanyConf==")
	_, value, exists := app.state.Get([]byte(KeyCompany(uid)))

	if !exists {
		return nil, ErrCompanyNotExists
	}
	var company Company
	err := UnmarshalMessage(value, &company)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(pubkey, company.CompanyPubkey) {
		return nil, ErrNoRight
	}
	if company.CompanyConf != nil && company.CompanyConf.CompanyConfId != "" {
		return nil, ErrCompanyConfExists
	}
	conf := &CompanyConf{}
	conf.CompanyConfId = companyConfId
	conf.CompanyConfName = companyConfName
	conf.IsAutoAdd = isAutoAdd
	conf.MaxAmount = maxAmount
	conf.MaxReceive = maxReceive
	conf.MinAmount = minAmount
	conf.OverdueDays = overdueDays
	conf.Rate = rate
	company.CompanyConf = conf
	save, err := MarshalMessage(&company)
	if err != nil {
		return nil, err
	}
	fmt.Println("addCompanyConf", company)
	app.state.Set([]byte(KeyCompany(uid)), save)

	event := &EventAddCompanyConf{}

	return &Response{Value: &Response_AddCompanyConf{&ResponseAddCompanyConf{InstructionId: instructionId, Event: &Event{Value: &Event_AddCompanyConf{event}}}}}, nil
}

func (app *GfcollectionApplication) checkEditCompanyConf(pubkey []byte, uid string, companyConfId string) error {
	_, value, exists := app.state.Get([]byte(KeyCompany(uid)))

	if !exists {
		return ErrCompanyNotExists
	}
	var company Company
	err := UnmarshalMessage(value, &company)
	if err != nil {
		return err
	}
	if !bytes.Equal(pubkey, company.CompanyPubkey) {
		return ErrNoRight
	}
	if company.CompanyConf == nil {
		return ErrCompanyConfNotExists
	}
	if company.CompanyConf.CompanyConfId != companyConfId {
		return ErrCompanyConfExists
	}
	return nil
}

func (app *GfcollectionApplication) editCompanyConf(pubkey []byte, uid, companyConfId, companyConfName string, isAutoAdd bool, maxAmount, maxReceive, minAmount int64, overdueDays int32, rate int32, instructionId int64) (*Response, error) {
	fmt.Println("editCompanyConf==")
	_, value, exists := app.state.Get([]byte(KeyCompany(uid)))

	if !exists {
		return nil, ErrCompanyNotExists
	}
	var company Company
	err := UnmarshalMessage(value, &company)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(pubkey, company.CompanyPubkey) {
		return nil, ErrNoRight
	}
	fmt.Println("editCompanyConf:before", company)
	if company.CompanyConf.CompanyConfId != companyConfId {
		return nil, ErrCompanyConfNotExists
	}

	company.CompanyConf.CompanyConfName = companyConfName
	company.CompanyConf.IsAutoAdd = isAutoAdd
	company.CompanyConf.MaxAmount = maxAmount
	company.CompanyConf.MaxReceive = maxReceive
	company.CompanyConf.MinAmount = minAmount
	company.CompanyConf.OverdueDays = overdueDays
	company.CompanyConf.Rate = rate

	fmt.Println("editCompanyConf:after", company)
	save, err := MarshalMessage(&company)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyCompany(uid)), save)
	event := &EventEditCompanyConf{}
	event.CompanyConfId = companyConfId
	return &Response{Value: &Response_EditCompanyConf{&ResponseEditCompanyConf{InstructionId: instructionId, Event: &Event{Value: &Event_EditCompanyConf{event}}}}}, nil
}

func (app *GfcollectionApplication) checkDelCompanyConf(pubkey []byte, uid, companyConfId string) error {
	_, value, exists := app.state.Get([]byte(KeyCompany(uid)))

	if !exists {
		return ErrCompanyNotExists
	}
	var company Company
	err := UnmarshalMessage(value, &company)
	if err != nil {
		return err
	}
	if !bytes.Equal(pubkey, company.CompanyPubkey) {
		return ErrNoRight
	}
	if company.CompanyConf == nil {
		return ErrCompanyConfNotExists
	}
	if company.CompanyConf.CompanyConfId != companyConfId {
		return ErrCompanyConfNotExists
	}
	return nil
}

func (app *GfcollectionApplication) delCompanyConf(pubkey []byte, uid, companyConfId string, instructionId int64) (*Response, error) {
	fmt.Println("delCompanyConf==")
	_, value, exists := app.state.Get([]byte(KeyCompany(uid)))

	if !exists {
		return nil, ErrCompanyNotExists
	}
	var company Company
	err := UnmarshalMessage(value, &company)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(pubkey, company.CompanyPubkey) {
		return nil, ErrNoRight
	}
	fmt.Println("delCompanyConf:after", company)
	company.CompanyConf = &CompanyConf{}
	fmt.Println("delCompanyConf:after", company)
	save, err := MarshalMessage(&company)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyCompany(uid)), save)
	event := &EventDelCompanyConf{}
	event.CompanyConfId = companyConfId
	return &Response{Value: &Response_DelCompanyConf{&ResponseDelCompanyConf{InstructionId: instructionId, Event: &Event{Value: &Event_DelCompanyConf{event}}}}}, nil
}

func (app *GfcollectionApplication) checkDelayCaseList(pubkey []byte, uid string, caseList []*CaseDelay) error {
	_, value, exists := app.state.Get([]byte(KeyCompany(uid)))

	if !exists {
		return ErrCompanyNotExists
	}
	var company Company
	err := UnmarshalMessage(value, &company)
	if err != nil {
		return err
	}
	if !bytes.Equal(pubkey, company.CompanyPubkey) {
		return ErrNoRight
	}

	for _, vv := range caseList {
		check := false
		for _, v := range company.CollectingList {
			if v.CaseId == vv.CaseId {
				check = true
				break
			}
		}
		if !check {
			for _, v := range company.DelayedList {
				if v.CaseId == vv.CaseId {
					check = true
					break
				}
			}
		}
		if !check {
			return ErrCompanyConfNotExists
		}
	}

	return nil
}

func (app *GfcollectionApplication) delayCaseList(pubkey []byte, uid string, caseList []*CaseDelay, instructionId int64) (*Response, error) {
	fmt.Println("delayCaseList==")
	_, value, exists := app.state.Get([]byte(KeyCompany(uid)))

	if !exists {
		return nil, ErrCompanyNotExists
	}
	var company Company
	err := UnmarshalMessage(value, &company)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(pubkey, company.CompanyPubkey) {
		return nil, ErrNoRight
	}
	fmt.Println("delayCaseList:after", company)
	tempCollecting := company.CollectingList[len(company.CollectingList):]
	delayList := company.DelayedList[len(company.DelayedList):]
	fmt.Println("TYPE", reflect.TypeOf(tempCollecting))

	for _, v := range company.CollectingList {

		check := false
		for _, vv := range caseList {
			if v.CaseId == vv.CaseId {
				check = true
				_, value, exists := app.state.Get([]byte(KeyCase(v.CaseId)))
				if !exists {
					return nil, ErrCaseNotExists
				}
				var c Case
				err := UnmarshalMessage(value, &c)
				if err != nil {
					return nil, err
				}
				delayList = append(delayList, &c)
				c.DelayDays += vv.Days
				save, err := MarshalMessage(&c)
				if err != nil {
					return nil, err
				}
				app.state.Set([]byte(KeyCase(v.CaseId)), save)
				break
			}
		}

		if !check {
			var c Case
			c.CaseId = v.CaseId
			c.BankCard = v.BankCard
			c.CaseArea = v.CaseArea
			c.CaseId = v.CaseId
			c.CaseOwner = v.CaseOwner
			c.Contract = v.Contract
			c.DebtAmount = v.DebtAmount
			c.Fees = v.Fees
			c.OriginalAmount = v.OriginalAmount
			c.OverdueDays = v.OverdueDays
			c.CaseState = v.CaseState
			c.BankId = v.BankId
			c.CompanyId = v.CompanyId
			tempCollecting = append(tempCollecting, &c)
		}
	}
	company.CollectingList = tempCollecting[:]
	for _, v := range company.DelayedList {
		for _, vv := range caseList {
			if v.CaseId == vv.CaseId {
				_, value, exists := app.state.Get([]byte(KeyCase(v.CaseId)))
				if !exists {
					return nil, ErrCaseNotExists
				}
				var c Case
				err := UnmarshalMessage(value, &c)
				if err != nil {
					return nil, err
				}
				c.DelayDays += vv.Days
				save, err := MarshalMessage(&c)
				if err != nil {
					return nil, err
				}
				app.state.Set([]byte(KeyCase(v.CaseId)), save)
				break
			}
		}
	}
	company.DelayedList = append(company.DelayedList, delayList[:]...)
	fmt.Println("delayCaseList:after", company)
	save, err := MarshalMessage(&company)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyCompany(uid)), save)
	event := &EventDelayCaseList{}
	event.CaseList = caseList
	return &Response{Value: &Response_DelayCaseList{&ResponseDelayCaseList{InstructionId: instructionId, Event: &Event{Value: &Event_DelayCaseList{event}}}}}, nil
}

func (app *GfcollectionApplication) checkCancelCaseList(pubkey []byte, uid string, caseList []string) error {
	if len(caseList) == 0 {
		return ErrNoEmpty
	}
	_, value, exists := app.state.Get([]byte(KeyCompany(uid)))

	if !exists {
		return ErrCompanyNotExists
	}
	var company Company
	err := UnmarshalMessage(value, &company)
	if err != nil {
		return err
	}
	if !bytes.Equal(pubkey, company.CompanyPubkey) {
		return ErrNoRight
	}
	for _, v := range caseList {
		check := false
		// for _, vv := range company.CollectingList {
		// 	if vv.CaseId == v {
		// 		check = true
		// 		break
		// 	}
		// }
		// if !check {
		// 	for _, vv := range company.DelayedList {
		// 		if vv.CaseId == v {
		// 			check = true
		// 			break
		// 		}
		// 	}
		// }
		if !check {
			for _, vv := range company.DeliverList {
				if vv.CaseId == v {
					check = true
					break
				}
			}
		}
		if !check {
			return ErrCaseNotExists
		}
	}
	return nil
}

func (app *GfcollectionApplication) cancelCaseList(pubkey []byte, uid string, caseList []string, instructionId int64) (*Response, error) {
	fmt.Println("cancelCaseList==")
	_, value, exists := app.state.Get([]byte(KeyCompany(uid)))

	if !exists {
		return nil, ErrCompanyNotExists
	}
	var company Company
	err := UnmarshalMessage(value, &company)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(pubkey, company.CompanyPubkey) {
		return nil, ErrNoRight
	}
	fmt.Println("cancelCaseList:before", company)
	tempCollecting := make([]*Case, 0)
	tempDelay := make([]*Case, 0)
	tempDeliver := make([]*Case, 0)
	for k, v := range company.CollectingList {
		check := false
		for _, vv := range caseList {
			if v.CaseId == vv {
				check = true
				break
			}
		}
		if !check {
			tempCollecting = append(tempCollecting, company.CollectingList[k])
		}
	}

	for k, v := range company.DelayedList {
		check := false
		for _, vv := range caseList {
			if v.CaseId == vv {
				check = true
				break
			}
		}
		if !check {
			tempDelay = append(tempDelay, company.DelayedList[k])
		}
	}
	for k, v := range company.DeliverList {
		check := false
		for _, vv := range caseList {
			if v.CaseId == vv {
				check = true
				break
			}
		}
		if !check {
			tempDeliver = append(tempDeliver, company.DeliverList[k])
		}
	}
	company.CollectingList = tempCollecting
	company.DelayedList = tempDelay
	company.DeliverList = tempDeliver
	fmt.Println("cancelCaseList:after", company)
	save, err := MarshalMessage(&company)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyCompany(uid)), save)
	for _, v := range caseList {
		_, value, exists := app.state.Get([]byte(KeyCase(v)))
		if !exists {
			return nil, ErrCaseNotExists
		}
		var scase Case
		err := UnmarshalMessage(value, &scase)
		if err != nil {
			return nil, err
		}
		err = app.reBackCasePool(&scase)
		if err != nil {
			return nil, err
		}
		scase.CompanyId = ""
		scase.CaseState = CaseState_CS_WAITED
		save, err := MarshalMessage(&scase)
		if err != nil {
			return nil, err
		}
		app.state.Set([]byte(KeyCase(v)), save)
	}

	event := &EventCancelCaseList{}
	event.CaseList = caseList
	return &Response{Value: &Response_CancelCaseList{&ResponseCancelCaseList{InstructionId: instructionId, Event: &Event{Value: &Event_CancelCaseList{event}}}}}, nil
}

func (app *GfcollectionApplication) checkFinishCaseList(pubkey []byte, uid string, caseList []string) error {
	if len(caseList) == 0 {
		return ErrNoEmpty
	}
	_, value, exists := app.state.Get([]byte(KeyCompany(uid)))

	if !exists {
		return ErrCompanyNotExists
	}
	var company Company
	err := UnmarshalMessage(value, &company)
	if err != nil {
		return err
	}
	if !bytes.Equal(pubkey, company.CompanyPubkey) {
		return ErrNoRight
	}
	for _, v := range caseList {
		check := false
		for _, vv := range company.CollectingList {
			if vv.CaseId == v {
				check = true
				break
			}
		}
		if !check {
			for _, vv := range company.DelayedList {
				if vv.CaseId == v {
					check = true
					break
				}
			}
		}
		if !check {
			return ErrCaseNotExists
		}
	}
	return nil
}

func (app *GfcollectionApplication) finishCaseList(pubkey []byte, uid string, caseList []string, instructionId int64) (*Response, error) {
	fmt.Println("finishCaseList==")
	_, value, exists := app.state.Get([]byte(KeyCompany(uid)))

	if !exists {
		return nil, ErrCompanyNotExists
	}
	var company Company
	err := UnmarshalMessage(value, &company)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(pubkey, company.CompanyPubkey) {
		return nil, ErrNoRight
	}
	fmt.Println("finishCaseList:before", company)
	tempCollecting := make([]*Case, 0)
	tempDelay := make([]*Case, 0)
	for k, v := range company.CollectingList {
		check := false
		for _, vv := range caseList {
			if v.CaseId == vv {
				check = true
				break
			}
		}
		if !check {
			tempCollecting = append(tempCollecting, company.CollectingList[k])
		}
	}

	for k, v := range company.DelayedList {
		check := false
		for _, vv := range caseList {
			if v.CaseId == vv {
				check = true
				break
			}
		}
		if !check {
			tempDelay = append(tempDelay, company.DelayedList[k])
		}
	}
	company.CollectingList = tempCollecting[:]
	company.DelayedList = tempDelay[:]
	save, err := MarshalMessage(&company)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyPackConfs()), save)
	for _, v := range caseList {
		_, value, exists := app.state.Get([]byte(KeyCase(v)))
		if !exists {
			return nil, ErrCaseNotExists
		}
		var scase Case
		err := UnmarshalMessage(value, &scase)
		if err != nil {
			return nil, err
		}
		scase.CaseState = CaseState_CS_DONE
		save, err := MarshalMessage(&scase)
		if err != nil {
			return nil, err
		}
		app.state.Set([]byte(KeyCase(v)), save)
	}
	fmt.Println("finishCaseList:after", company)
	result, err := app.autoDeliver()
	if err != nil {
		return nil, err
	}
	event := &EventFinishCaseList{}
	event.CaseList = caseList
	event.Result = result
	return &Response{Value: &Response_FinishCaseList{&ResponseFinishCaseList{InstructionId: instructionId, Event: &Event{Value: &Event_FinishCaseList{event}}}}}, nil
}

func (app *GfcollectionApplication) checkCollectCaseList(pubkey []byte, uid string, caseList []string) error {
	if len(caseList) == 0 {
		return ErrNoEmpty
	}
	_, value, exists := app.state.Get([]byte(KeyCompany(uid)))

	if !exists {
		return ErrCompanyNotExists
	}
	var company Company
	err := UnmarshalMessage(value, &company)
	if err != nil {
		return err
	}
	if !bytes.Equal(pubkey, company.CompanyPubkey) {
		return ErrNoRight
	}
	for _, v := range caseList {
		check := false
		for _, vv := range company.DeliverList {
			if vv.CaseId == v {
				check = true
				break
			}
		}
		if !check {
			return ErrCaseNotExists
		}
	}
	return nil
}

func (app *GfcollectionApplication) collectCaseList(pubkey []byte, uid string, caseList []string, instructionId int64) (*Response, error) {
	fmt.Println("collectCaseList==")
	_, value, exists := app.state.Get([]byte(KeyCompany(uid)))

	if !exists {
		return nil, ErrCompanyNotExists
	}
	var company Company
	err := UnmarshalMessage(value, &company)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(pubkey, company.CompanyPubkey) {
		return nil, ErrNoRight
	}
	fmt.Println("collectCaseList:before", company)
	tempDeliver := make([]*Case, 0)
	for k, v := range company.DeliverList {
		check := false
		for _, vv := range caseList {
			if v.CaseId == vv {
				check = true
				break
			}
		}
		if !check {
			tempDeliver = append(tempDeliver, company.DeliverList[k])
		}
	}

	company.DeliverList = tempDeliver

	for _, v := range caseList {
		_, value, exists := app.state.Get([]byte(KeyCase(v)))
		if !exists {
			return nil, ErrCaseNotExists
		}
		var scase Case
		err := UnmarshalMessage(value, &scase)
		if err != nil {
			return nil, err
		}
		scase.CaseState = CaseState_CS_COLLECTIING
		scase.CompanyId = uid
		company.CollectingList = append(company.CollectingList, &scase)
		save, err := MarshalMessage(&scase)
		if err != nil {
			return nil, err
		}
		app.state.Set([]byte(KeyCase(v)), save)
	}
	fmt.Println("collectCaseList:after", company)
	save, err := MarshalMessage(&company)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyCompany(uid)), save)
	event := &EventFinishCaseList{}
	event.CaseList = caseList
	return &Response{Value: &Response_FinishCaseList{&ResponseFinishCaseList{InstructionId: instructionId, Event: &Event{Value: &Event_FinishCaseList{event}}}}}, nil
}

func (app *GfcollectionApplication) checkSwitchCase(pubkey []byte, uid string, caseId, companyId string) error {
	_, value, exists := app.state.Get([]byte(KeyBank(uid)))
	if !exists {
		return ErrBankNotExists
	}
	var bank Bank
	err := UnmarshalMessage(value, &bank)
	if err != nil {
		return ErrStorage
	}
	if bank.BankId != uid {
		return ErrNoRight
	}
	_, _, exists = app.state.Get([]byte(KeyCase(caseId)))
	if !exists {
		return ErrCaseNotExists
	}
	_, _, exists = app.state.Get([]byte(KeyCompany(companyId)))
	if !exists {
		return ErrCompanyNotExists
	}
	return nil
}

func (app *GfcollectionApplication) switchCase(pubkey []byte, uid string, caseId, companyId string, instructionId int64) (*Response, error) {
	fmt.Println("switchCase==")
	_, value, exists := app.state.Get([]byte(KeyCase(caseId)))
	if !exists {
		return nil, ErrCaseNotExists
	}
	var c Case
	err := UnmarshalMessage(value, &c)
	if err != nil {
		return nil, ErrStorage
	}
	_, value, exists = app.state.Get([]byte(KeyCompany(companyId)))
	if !exists {
		return nil, ErrCompanyNotExists
	}
	var tocompany Company
	err = UnmarshalMessage(value, &tocompany)
	if err != nil {
		return nil, ErrStorage
	}
	_, value, exists = app.state.Get([]byte(KeyCompany(c.CompanyId)))
	if !exists {
		return nil, ErrCompanyNotExists
	}
	var oldcompany Company
	err = UnmarshalMessage(value, &oldcompany)
	if err != nil {
		return nil, ErrStorage
	}
	check := false
	add := make([]*Case, 0)
	for k, v := range oldcompany.DeliverList {
		if v.CaseId == caseId {
			temp := oldcompany.DeliverList[:]
			oldcompany.DeliverList = temp[:k]
			oldcompany.DeliverList = append(oldcompany.DeliverList, temp[k+1:]...)
			add = append(add, temp[k:k+1]...)
			fmt.Println("add==", add)
			check = true
			break
		}
	}
	if !check {
		return nil, ErrCaseNotExists
	}
	fmt.Println("tocompany.DeliverList", tocompany.DeliverList)
	tocompany.DeliverList = append(tocompany.DeliverList, add...)
	save, err := MarshalMessage(&tocompany)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyCompany(companyId)), save)
	save, err = MarshalMessage(&oldcompany)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyCompany(c.CompanyId)), save)

	event := &EventSwitchCase{}
	event.CaseId = caseId
	event.CompanyId = companyId
	return &Response{Value: &Response_SwitchCase{&ResponseSwitchCase{InstructionId: instructionId, Event: &Event{Value: &Event_SwitchCase{event}}}}}, nil
}

func (app *GfcollectionApplication) checkUpdateWeight(pubkey []byte, uid string, weightList []*CompanyWeight) error {
	_, _, exists := app.state.Get([]byte(KeyBank(uid)))

	if !exists {
		return ErrBankNotExists
	}
	var sum int32
	for _, v := range weightList {
		_, _, exists := app.state.Get([]byte(KeyCompany(v.CompanyId)))
		if !exists {
			return ErrCompanyNotExists
		}
		sum += v.Weight
	}
	if sum != int32(10000) {
		return ErrException
	}
	return nil
}

func (app *GfcollectionApplication) updateWeight(pubkey []byte, uid string, weightList []*CompanyWeight, instructionId int64) (*Response, error) {
	//fmt.Println("updateWeight==")
	_, value, exists := app.state.Get([]byte(KeyAreaList()))
	if !exists {
		return nil, ErrException
	}
	var areaList AreaList
	err := UnmarshalMessage(value, &areaList)
	if err != nil {
		return nil, err
	}
	//fmt.Println("updateWeight before", areaList)
	for _, v := range weightList {
		for kk, vv := range areaList.AreaCompany {
			if v.Area == vv.Area {
				for kkk, vvv := range vv.Companys {
					if v.CompanyId == vvv.CompanyId {
						areaList.AreaCompany[kk].Companys[kkk].Weight = v.Weight
						break
					}
				}
			}
		}
	}
	//fmt.Println("updateWeight after", areaList)
	save, err := MarshalMessage(&areaList)
	if err != nil {
		return nil, ErrStorage
	}
	app.state.Set([]byte(KeyAreaList()), save)

	event := &EventUpdateWeight{}

	return &Response{Value: &Response_UpdateWeight{&ResponseUpdateWeight{InstructionId: instructionId, Event: &Event{Value: &Event_UpdateWeight{event}}}}}, nil
}

func (app *GfcollectionApplication) checkDeliverCaseList(pubkey []byte, uid string, caseList []string) error {
	if len(caseList) == 0 {
		return ErrNoEmpty
	}
	_, value, exists := app.state.Get([]byte(KeyBank(uid)))
	if !exists {
		return ErrBankNotExists
	}
	var bank Bank
	err := UnmarshalMessage(value, &bank)
	if err != nil {
		return err
	}
	if !bytes.Equal(pubkey, bank.Pubkey) {
		return ErrNoRight
	}

	for _, v := range caseList {
		_, value, exists := app.state.Get([]byte(KeyCase(v)))
		if !exists {
			return ErrCaseNotExists
		}
		var c Case
		err := UnmarshalMessage(value, &c)
		if err != nil {
			return err
		}
		fmt.Println("case state ", c.CaseState)
		if c.CaseState != CaseState_CS_DELIVER {
			return ErrWrongState
		}

	}
	return nil
}

func (app *GfcollectionApplication) deliverCaseList(pubkey []byte, uid string, caseList []string, instructionId int64) (*Response, error) {
	fmt.Println("deliverCaseList==")

	_, _, exists := app.state.Get([]byte(KeyBank(uid)))

	if !exists {
		return nil, ErrBankNotExists
	}
	succList := make([]string, 0)
	for _, v := range caseList {
		_, value, exists := app.state.Get([]byte(KeyCase(v)))
		if !exists {
			return nil, ErrCaseNotExists
		}
		var c Case
		err := UnmarshalMessage(value, &c)
		if err != nil {
			return nil, err
		}
		_, value, exists = app.state.Get([]byte(KeyCompany(c.CompanyId)))

		if !exists {
			return nil, ErrCompanyNotExists
		}

		var company Company
		err = UnmarshalMessage(value, &company)
		if err != nil {
			return nil, err
		}
		_, _, exists = app.state.Get([]byte(KeyAreaBool(company.CompanyArea)))
		if !exists {
			app.state.Set([]byte(KeyAreaBool(company.CompanyArea)), []byte(company.CompanyArea))
		}

		fmt.Println("company:undeliverCaseList:before", len(company.UnDeliverList))
		tempRest := make([]*Case, 0)
		tempDeliver := make([]*Case, 0)

		for kk, vv := range company.UnDeliverList {
			check := false
			for _, vvv := range caseList {
				if vv.CaseId == vvv && vv.CompanyId == c.CompanyId {
					check = true
					tempDeliver = append(tempDeliver, company.UnDeliverList[kk])
					succList = append(succList, vvv)
					break
				}
			}
			if !check {
				tempRest = append(tempRest, company.UnDeliverList[kk])
			}
		}
		company.UnDeliverList = tempRest
		fmt.Println("company:undeliverCaseList:AFTER", len(company.UnDeliverList))

		fmt.Println("company:deliverCaseList:before", len(company.DeliverList))
		company.DeliverList = append(company.DeliverList, tempDeliver...)
		fmt.Println("company:deliverCaseList:after", len(company.DeliverList))
		save, err := MarshalMessage(&company)
		if err != nil {
			return nil, err
		}
		app.state.Set([]byte(KeyCompany(c.CompanyId)), save)

	}

	event := &EventDeliverCaseList{}
	event.CaseIds = succList
	return &Response{Value: &Response_DeliverCaseList{&ResponseDeliverCaseList{InstructionId: instructionId, Event: &Event{Value: &Event_DeliverCaseList{event}}}}}, nil
}
