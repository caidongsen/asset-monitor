package gwpoints

import (
	"bytes"
	//"log"
)

func (app *GwpointsApplication) checkInitPlatform(pubkey []byte) error {
	if !isOriginalAdmin(pubkey) {
		return ErrNotAdmin
	}
	_, _, exists := app.state.Get([]byte(KeyPlatform()))
	if exists {
		return ErrPlatformExist
	}
	return nil
}

func (app *GwpointsApplication) initPlatform(pubkey []byte, platformPubkey []byte, info []byte, instructionId int64) (*Response, error) {
	err := app.checkInitPlatform(pubkey)
	if err != nil {
		return nil, err
	}
	platform := &Platform{}
	platform.Pubkey = platformPubkey[:]
	platform.Info = info[:]
	save, err := MarshalMessage(platform)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyPlatform()), save)
	event := &EventInitPlatform{}
	event.Pubkey = platformPubkey
	return &Response{Value: &Response_InitPlatform{&ResponseInitPlatform{InstructionId: instructionId, Event: &Event{Value: &Event_InitPlatform{event}}}}}, nil
}

func (app *GwpointsApplication) checkUserCreate(pubkey []byte, userUid int64) error {
	_, _, exists := app.state.Get([]byte(KeyUser(userUid)))
	if exists {
		return ErrUserExist
	}
	return nil
}

func (app *GwpointsApplication) userCreate(pubkey []byte, userUid int64, userPubkey []byte, info []byte, instructionId int64) (*Response, error) {
	err := app.checkUserCreate(pubkey, userUid)
	if err != nil {
		return nil, err
	}
	user := &User{}
	user.Pubkey = userPubkey[:]
	user.Info = info[:]
	save, err := MarshalMessage(user)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyUser(userUid)), save)
	event := &EventUserCreate{}
	event.UserUid = userUid
	return &Response{Value: &Response_UserCreate{&ResponseUserCreate{InstructionId: instructionId, Event: &Event{Value: &Event_UserCreate{event}}}}}, nil
}

func (app *GwpointsApplication) checkChangePubkey(userUid int64, oldPubkey, newPubkey []byte) error {
	_, value, exists := app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return ErrUserNotExist
	}
	var user User
	err := UnmarshalMessage(value, &user)
	if err != nil {
		return ErrStorage
	}
	if !bytes.Equal(user.Pubkey, oldPubkey) {
		return ErrNoRight
	}
	return nil
}

func (app *GwpointsApplication) changePubkey(userUid int64, oldPubkey, newPubkey []byte, instructionId int64) (*Response, error) {
	_, value, exists := app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return nil, ErrUserNotExist
	}
	var user User
	err := UnmarshalMessage(value, &user)
	if err != nil {
		return nil, ErrStorage
	}
	if !bytes.Equal(user.Pubkey, oldPubkey) {
		return nil, ErrNoRight
	}
	copy(user.Pubkey[:], newPubkey[:])
	save, err := MarshalMessage(&user)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyUser(userUid)), save)
	event := &EventChangePubkey{}
	event.NewPubkey = newPubkey[:]
	return &Response{Value: &Response_ChangePubkey{&ResponseChangePubkey{InstructionId: instructionId, Event: &Event{Value: &Event_ChangePubkey{event}}}}}, nil
}

func (app *GwpointsApplication) checkDistributeGwpoints(pubkey []byte, gwPoints int64) error {
	if !app.isAdmin(pubkey) {
		return ErrNotAdmin
	}
	if gwPoints <= 0 {
		return ErrWrongPoints
	}

	return nil
}

func (app *GwpointsApplication) distributeGwpoints(pubkey []byte, userUid int64, userPubkey []byte, gwPoints int64, instructionId int64) (*Response, error) {
	err := app.checkDistributeGwpoints(pubkey, gwPoints)
	if err != nil {
		return nil, err
	}
	_, value, exists := app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return nil, ErrUserNotExist
	}
	var user User
	err = UnmarshalMessage(value, &user)
	if err != nil {
		return nil, ErrStorage
	}
	if !bytes.Equal(userPubkey, user.Pubkey) {
		return nil, ErrWrongPubkey
	}
	user.GwPoints += gwPoints
	save, err := MarshalMessage(&user)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyUser(userUid)), save)
	event := &EventDistributeGwpoints{}
	event.GwPoints = gwPoints
	return &Response{Value: &Response_DistributeGwpoints{&ResponseDistributeGwpoints{InstructionId: instructionId, Event: &Event{Value: &Event_DistributeGwpoints{event}}}}}, nil
}

func (app *GwpointsApplication) checkSetCompanyExchangeRate(pubkey []byte) error {
	if !app.isAdmin(pubkey) {
		return ErrNotAdmin
	}

	return nil
}

func (app *GwpointsApplication) setCompanyExchangeRate(pubkey []byte, companyId int64, companyNum, gwNum int32, info []byte, instructionId int64) (*Response, error) {
	err := app.checkSetCompanyExchangeRate(pubkey)
	if err != nil {
		return nil, err
	}
	_, value, exists := app.state.Get([]byte(KeyPlatform()))
	if !exists {
		return nil, ErrPlatformNotExist
	}
	var platform Platform
	err = UnmarshalMessage(value, &platform)
	if err != nil {
		return nil, ErrStorage
	}
	//log.Println("debug setCompanyExchangeRate 1", companyId, companyNum, gwNum)
	//log.Println("debug setCompanyExchangeRate 2", platform)
	bHas := false
	for _, v := range platform.CompanyStatistics {
		//log.Println("debug setCompanyExchangeRate 3", v)
		if v.Id == companyId {
			v.CompanyNum = companyNum
			v.GwNum = gwNum
			copy(v.Info[:], info[:])
			bHas = true
			break
		}
	}
	if !bHas {
		companyStatistics := &CompanyStatistics{}
		companyStatistics.Id = companyId
		companyStatistics.CompanyNum = companyNum
		companyStatistics.GwNum = gwNum
		companyStatistics.Info = info[:]
		platform.CompanyStatistics = append(platform.CompanyStatistics, companyStatistics)
		//log.Println("debug setCompanyExchangeRate 4", companyStatistics)
	}
	save, err := MarshalMessage(&platform)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyPlatform()), save)
	//log.Println("debug setCompanyExchangeRate 5", platform)
	event := &EventSetCompanyExchangeRate{}
	event.CompanyNum = companyNum
	event.GwNum = gwNum
	return &Response{Value: &Response_SetCompanyExchangeRate{&ResponseSetCompanyExchangeRate{InstructionId: instructionId, Event: &Event{Value: &Event_SetCompanyExchangeRate{event}}}}}, nil
}

func (app *GwpointsApplication) checkBuyGwpoints(pubkey []byte, userUid int64) error {
	_, value, exists := app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return ErrUserNotExist
	}
	var user User
	err := UnmarshalMessage(value, &user)
	if err != nil {
		return ErrStorage
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return ErrNoRight
	}

	return nil
}

func (app *GwpointsApplication) buyGwpoints(pubkey []byte, userUid int64, companyId int64, companyPoints int64, instructionId int64) (*Response, error) {
	_, value, exists := app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return nil, ErrUserNotExist
	}
	var user User
	err := UnmarshalMessage(value, &user)
	if err != nil {
		return nil, ErrStorage
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return nil, ErrNoRight
	}
	_, value, exists = app.state.Get([]byte(KeyPlatform()))
	if !exists {
		return nil, ErrPlatformNotExist
	}
	var platform Platform
	err = UnmarshalMessage(value, &platform)
	if err != nil {
		return nil, ErrStorage
	}
	companyNum := int64(-1)
	gwNum := int64(-1)
	for _, v := range platform.CompanyStatistics {
		//log.Println(v)
		if v.Id == companyId {
			companyNum = int64(v.CompanyNum)
			gwNum = int64(v.GwNum)
			v.In += companyPoints
			v.CompanyPoints += companyPoints
			break
		}
	}
	if companyNum == -1 {
		return nil, ErrWrongCompanyId
	}
	user.GwPoints += (companyPoints * gwNum) / companyNum
	save, err := MarshalMessage(&platform)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyPlatform()), save)
	save, err = MarshalMessage(&user)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyUser(userUid)), save)
	event := &EventBuyGwpoints{}
	event.Points = (companyPoints * gwNum) / companyNum
	return &Response{Value: &Response_BuyGwpoints{&ResponseBuyGwpoints{InstructionId: instructionId, Event: &Event{Value: &Event_BuyGwpoints{event}}}}}, nil
}

func (app *GwpointsApplication) checkSellGwpoints(pubkey []byte, userUid int64) error {
	_, value, exists := app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return ErrUserNotExist
	}
	var user User
	err := UnmarshalMessage(value, &user)
	if err != nil {
		return ErrStorage
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return ErrNoRight
	}

	return nil
}

func (app *GwpointsApplication) sellGwpoints(pubkey []byte, userUid int64, companyId int64, companyPoints int64, instructionId int64) (*Response, error) {
	_, value, exists := app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return nil, ErrUserNotExist
	}
	var user User
	err := UnmarshalMessage(value, &user)
	if err != nil {
		return nil, ErrStorage
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return nil, ErrNoRight
	}
	_, value, exists = app.state.Get([]byte(KeyPlatform()))
	if !exists {
		return nil, ErrPlatformNotExist
	}
	var platform Platform
	err = UnmarshalMessage(value, &platform)
	if err != nil {
		return nil, ErrStorage
	}
	companyNum := int64(-1)
	gwNum := int64(-1)
	for _, v := range platform.CompanyStatistics {
		if v.Id == companyId {
			companyNum = int64(v.CompanyNum)
			gwNum = int64(v.GwNum)
			v.GwPoints += (companyPoints * gwNum) / companyNum
			v.CompanyPoints -= companyPoints
			v.Out += companyPoints
			break
		}
	}
	if companyNum == -1 {
		return nil, ErrWrongCompanyId
	}
	user.GwPoints -= (companyPoints * gwNum) / companyNum
	save, err := MarshalMessage(&platform)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyPlatform()), save)
	save, err = MarshalMessage(&user)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyUser(userUid)), save)
	event := &EventSellGwPoints{}
	event.Points = (companyPoints * gwNum) / companyNum
	return &Response{Value: &Response_SellGwPoints{&ResponseSellGwPoints{InstructionId: instructionId, Event: &Event{Value: &Event_SellGwPoints{event}}}}}, nil
}

func (app *GwpointsApplication) checkBuyGoods(pubkey []byte, userUid int64) error {
	_, value, exists := app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return ErrUserNotExist
	}
	var user User
	err := UnmarshalMessage(value, &user)
	if err != nil {
		return ErrStorage
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return ErrNoRight
	}
	return nil
}

func (app *GwpointsApplication) buyGoods(pubkey []byte, userUid int64, gwPoints int64, orderId, productName string, productNum int32, instructionId int64) (*Response, error) {
	_, value, exists := app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return nil, ErrUserNotExist
	}
	var user User
	err := UnmarshalMessage(value, &user)
	if err != nil {
		return nil, ErrStorage
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return nil, ErrNoRight
	}
	user.GwPoints -= gwPoints
	order := &Order{}
	order.Id = orderId
	order.Name = productName
	order.Num = productNum
	user.Orders = append(user.Orders, order)
	save, err := MarshalMessage(&user)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyUser(userUid)), save)
	event := &EventBuyGoods{}
	event.Points = gwPoints
	return &Response{Value: &Response_BuyGoods{&ResponseBuyGoods{InstructionId: instructionId, Event: &Event{Value: &Event_BuyGoods{event}}}}}, nil
}

func (app *GwpointsApplication) checkClear(pubkey []byte) error {
	if !app.isAdmin(pubkey) {
		return ErrNotAdmin
	}

	return nil
}

func (app *GwpointsApplication) clear(pubkey []byte, companyId int64, instructionId int64) (*Response, error) {
	err := app.checkClear(pubkey)
	if err != nil {
		return nil, err
	}
	_, value, exists := app.state.Get([]byte(KeyPlatform()))
	if !exists {
		return nil, ErrPlatformNotExist
	}
	var platform Platform
	err = UnmarshalMessage(value, &platform)
	if err != nil {
		return nil, ErrStorage
	}
	var tmpCompanyPoints, tmpIn, tmpOut, tmpGwPoints int64
	for _, v := range platform.CompanyStatistics {
		if v.Id == companyId {
			tmpCompanyPoints = v.CompanyPoints
			tmpIn = v.In
			tmpOut = v.Out
			tmpGwPoints = v.GwPoints
			v.CompanyPoints = 0
			v.In = 0
			v.Out = 0
			v.GwPoints = 0
			break
		}
	}
	save, err := MarshalMessage(&platform)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyPlatform()), save)
	event := &EventClear{}
	event.CompanyPoints = tmpCompanyPoints
	event.In = tmpIn
	event.Out = tmpOut
	event.GwPoints = tmpGwPoints
	return &Response{Value: &Response_Clear{&ResponseClear{InstructionId: instructionId, Event: &Event{Value: &Event_Clear{event}}}}}, nil
}

func (app *GwpointsApplication) checkSyncPoints(pubkey []byte, gwPoints int64) error {
	if gwPoints <= 0 {
		return ErrWrongPoints
	}

	return nil
}

func (app *GwpointsApplication) syncPoints(pubkey []byte, userUid int64, userPubkey []byte, gwPoints int64, added int32, instructionId int64) (*Response, error) {
	err := app.checkSyncPoints(pubkey, gwPoints)
	if err != nil {
		return nil, err
	}
	_, value, exists := app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return nil, ErrUserNotExist
	}
	var user User
	err = UnmarshalMessage(value, &user)
	if err != nil {
		return nil, ErrStorage
	}
	if !bytes.Equal(userPubkey, user.Pubkey) {
		return nil, ErrWrongPubkey
	}
	if !bytes.Equal(pubkey, user.Pubkey) {
		return nil, ErrNoRight
	}
	if added == 1 {
		user.GwPoints += gwPoints
	}
	if added == 2 {
		user.GwPoints -= gwPoints
	}
	save, err := MarshalMessage(&user)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyUser(userUid)), save)
	event := &EventSyncPoints{}
	event.GwPoints = gwPoints
	return &Response{Value: &Response_SyncPoints{&ResponseSyncPoints{InstructionId: instructionId, Event: &Event{Value: &Event_SyncPoints{event}}}}}, nil
}
