package mideaBill

import (
	"bytes"
    "strings"
	"time"
)

/*
数据库说明
票：  票号(key)-->票信息(value)
企业：组织代码(key)-->企业信息(value)
用户：用户公钥(key)-->用户信息(value)
      用户名(key)-->用户信息(value)
*/

var (
	formatDate     = "2006-01-02"
	formatDateTime = "2006-01-02 15:04:05"
)

// 检查--用户身份校验
func (app *MideaBillApplication) checkUserSign(req *Request) error {
	if req.UserName == "" {
		return ErrUserNameIsNull
	}

	_, value, exists := app.state.Get(KeyUser(req.UserName))
	if !exists {
		return ErrUserNotExist
	}

	userInfo := &UserInfo{}
	err := UnmarshalMessage(value, userInfo)
	if err != nil {
		return ErrStorage
	}

	if !bytes.Equal(req.Pubkey,  userInfo.UserPublicKey) {
		return ErrPubkeyNotMatch
	}

	return nil
}

// 检查-初始化平台
func (app *MideaBillApplication) checkInitPlatform (req *Request) error {
	_, _, exists := app.state.Get(KeyPlatform())
	if exists {
		return ErrPlatformIsInit
	}

	initPlatform := req.GetInitPlatform()
	if initPlatform == nil {
		return ErrGetValueIsNull
	}

	if req.UserName != SdkUser {
		return ErrUserNameIsNotSDKUser
	}
	if initPlatform.UserName == "" {
		return ErrUserNameIsNull
	}
	if len(initPlatform.UserPublicKey) != 32 {
		return ErrUserPublicKey
	}
	if initPlatform.EntCode == "" {
		return ErrEntCodeIsNull
	}
	if initPlatform.EntName == "" {
		return ErrEntNameIsNull
	}
	if initPlatform.SdkUserName == "" {
		return ErrSDKUserNameIsNull
	}
	if initPlatform.SdkUserName != SdkUser {
		return ErrSDKUserNameIsNotMatch
	}
	if len(initPlatform.SdkPublicKey) != 32 {
		return ErrSDKPublicKey
	}

	return nil
}

// 初始化平台
func (app *MideaBillApplication) initPlatform(req *Request) (*Response, error) {
	initPlatform := req.GetInitPlatform()
	if initPlatform == nil {
		return nil, ErrGetValueIsNull
	}

    user := &UserInfo{}
	user.UserName = initPlatform.UserName
	user.UserPublicKey  = initPlatform.UserPublicKey
	user.UserPublicKeyList = append(user.UserPublicKeyList, user.UserPublicKey)
	user.EntCode = initPlatform.EntCode
	user.EntName = initPlatform.EntName
	user.EntPublicKey = GetEntPublicKey(initPlatform.EntCode)
	user.CreateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)
	user.UpdateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)

	save, err := MarshalMessage(user)
	if err != nil {
		return nil, err
	}

	app.state.Set(KeyUser(user.UserName), save)

	sdkUser := &UserInfo{}
	sdkUser.UserName = initPlatform.SdkUserName
	sdkUser.UserPublicKey  = initPlatform.SdkPublicKey
	sdkUser.UserPublicKeyList = append(sdkUser.UserPublicKeyList, sdkUser.UserPublicKey)
	sdkUser.EntCode = initPlatform.EntCode
	sdkUser.EntName = initPlatform.EntName
	sdkUser.EntPublicKey = GetEntPublicKey(initPlatform.EntCode)
	sdkUser.CreateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)
	sdkUser.UpdateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)

	save, err = MarshalMessage(sdkUser)
	if err != nil {
		return nil, err
	}

	app.state.Set(KeyUser(sdkUser.UserName), save)

	ent := &EntInfo{}
	ent.EntCode = initPlatform.EntCode
	ent.EntName = initPlatform.EntName
	ent.EntPublicKey = GetEntPublicKey(ent.EntCode)
	ent.CreateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)
	ent.UpdateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)

	save, err = MarshalMessage(ent)
	if err != nil {
		return nil, err
	}

	app.state.Set(KeyEnt(ent.EntCode), save)
	app.state.Set(KeyPlatform(), nil)

	return &Response{Value: &Response_InitPlatform{&ResponseInitPlatform{InstructionId: req.InstructionId}}}, nil
}

// 检查--注册用户
func (app *MideaBillApplication) checkRegisterUser(req *Request) error {
	registerUserUser := req.GetRegisterUser()
	if registerUserUser == nil {
		return ErrGetValueIsNull
	}

	// 用户信息都是必填字段，不可为空
	if registerUserUser.EntCode == "" {
		return ErrEntCodeIsNull
	}

	_, _, exists := app.state.Get(KeyUser(req.UserName))
	if exists {
		return ErrUserExist
	}

	return nil
}

// 用户注册
func (app *MideaBillApplication) registerUser(req *Request) (*Response, error) {
	registerUser := req.GetRegisterUser()
	if registerUser == nil {
		return nil, ErrGetValueIsNull
	}

	err := app.checkRegisterUser(req)
	if err != nil {
		return nil, err
	}

	user := &UserInfo{}
	user.UserName = req.UserName
	user.UserPublicKey = req.GetPubkey()
	user.UserPublicKeyList = append(user.UserPublicKeyList, user.UserPublicKey)
	user.EntCode = registerUser.EntCode
	user.EntPublicKey = nil
	user.CreateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)
	user.UpdateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)

	save, err := MarshalMessage(user)
	if err != nil {
		return nil, err
	}

    // key-value :   用户名--用户信息
	app.state.Set(KeyUser(user.UserName), save)

	return &Response{Value: &Response_RegisterUser{&ResponseRegisterUser{InstructionId: req.InstructionId}}}, nil
}

// 检查--添加用户
func (app *MideaBillApplication) checkAddUser(req *Request) error {
	addUser := req.GetAddUser()
	if addUser == nil {
		return ErrGetValueIsNull
	}

	// 用户信息都是必填字段，不可为空
	if addUser.UserName == "" {
		return ErrUserNameIsNull
	}
	if addUser.Operator == "" {
		return ErrOperatorIsNull
	}
	if len(addUser.UserPublicKey) != 32 {
		return ErrUserPublicKey
	}
	
	if req.UserName != SdkUser {
		return ErrUserNameIsNotSDKUser
	}

	// 判断是否为企业管理员添加用户
	_, valueUserInfo, exists := app.state.Get(KeyUser(addUser.Operator))
	if !exists {
		return ErrUserNotExist
	}
	userInfo := &UserInfo{}
	err := UnmarshalMessage(valueUserInfo, userInfo)
	if err != nil {
		return ErrStorage
	}

	_, _, exists = app.state.Get(KeyUser(addUser.UserName))
	if exists {
		return ErrUserExist
	}

	return nil
}

// 添加用户
func (app *MideaBillApplication) addUser(req *Request) (*Response, error) {
	addUser := req.GetAddUser()
	if addUser == nil {
		return nil, ErrGetValueIsNull
	}

	err := app.checkAddUser(req)
	if err != nil {
		return nil, err
	}

	// 查询企业信息
	_, valueUserInfo, exists := app.state.Get(KeyUser(addUser.Operator))
	if !exists {
		return nil, ErrUserNotExist
	}
	userInfo := &UserInfo{}
	err = UnmarshalMessage(valueUserInfo, userInfo)
	if err != nil {
		return nil, ErrStorage
	}

	user := &UserInfo{}
	user.UserName = addUser.UserName
	user.UserPublicKey = addUser.UserPublicKey
	user.UserPublicKeyList = append(user.UserPublicKeyList, user.UserPublicKey)
	user.EntName = userInfo.EntName
	user.EntCode = userInfo.EntCode
	user.EntPublicKey = userInfo.EntPublicKey
	user.CreateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)
	user.UpdateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)

	save, err := MarshalMessage(user)
	if err != nil {
		return nil, err
	}

    // key-value :   用户名--用户信息
	app.state.Set(KeyUser(user.UserName), save)

	return &Response{Value: &Response_AddUser{&ResponseAddUser{InstructionId: req.InstructionId}}}, nil
}

// 检查--用户密码修改
func (app *MideaBillApplication) checkUserPwdModify(req *Request) error {
	userPwdModify := req.GetUserPwdModify()
	if userPwdModify == nil {
		return ErrGetValueIsNull
	}

	//新公钥是必填字段，不可为空
	if len(userPwdModify.UserPublicKey) != 32 {
		return ErrUserPublicKey
	}

	_, _, exists := app.state.Get(KeyUser(req.UserName))
	if !exists {
		return ErrUserNotExist
	}

	return nil
}

// 用户密码修改
func (app *MideaBillApplication) userPwdModify(req *Request) (*Response, error) {
	userPwdModify := req.GetUserPwdModify()
	if userPwdModify == nil {
		return nil, ErrGetValueIsNull
	}

	err := app.checkUserPwdModify(req)
	if err != nil {
		return nil, err
	}

	_, value, exists := app.state.Get(KeyUser(req.UserName))
	if !exists {
		return nil, ErrUserNotExist
	}
	user := &UserInfo{}
	err = UnmarshalMessage(value, user)
	if err != nil {
		return nil, ErrStorage
	}
	user.UserPublicKey = userPwdModify.UserPublicKey
	user.UserPublicKeyList = append(user.UserPublicKeyList, user.UserPublicKey)
	user.UpdateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)

	save, err := MarshalMessage(user)
	if err != nil {
		return nil, err
	}

    // key-value :   用户名--用户信息
	app.state.Set(KeyUser(req.UserName), save)

	return &Response{Value: &Response_UserPwdModify{&ResponseUserPwdModify{InstructionId: req.InstructionId}}}, nil
}

// 检查--用户密码重置
func (app *MideaBillApplication) checkUserPwdReset(req *Request) error {
	userPwdReset := req.GetUserPwdReset()
	if userPwdReset == nil {
		return ErrGetValueIsNull
	}

	if userPwdReset.UserName == "" {
		return ErrUserNameIsNull
	}
	if len(userPwdReset.UserPublicKey) != 32 {
		return ErrUserPublicKey
	}

	//_, _, exists := app.state.Get(KeyUser(req.UserName))
	if req.UserName != SdkUser {
		return ErrUserNameIsNotSDKUser
	}

	_, _, exists := app.state.Get(KeyUser(userPwdReset.UserName))
	if !exists {
		return ErrUserNotExist
	}

	return nil
}

// 用户密码重置
func (app *MideaBillApplication) userPwdReset(req *Request) (*Response, error) {
	userPwdReset := req.GetUserPwdReset()
	if userPwdReset == nil {
		return nil, ErrGetValueIsNull
	}

	err := app.checkUserPwdReset(req)
	if err != nil {
		return nil, err
	}

	//根据用户名查询用户信息
	_, value, exists := app.state.Get(KeyUser(userPwdReset.UserName))
	if !exists {
		return nil, ErrUserNotExist
	}

	user := &UserInfo{}
	err = UnmarshalMessage(value, user)
	if err != nil {
		return nil, ErrStorage
	}
	user.UserPublicKey = userPwdReset.UserPublicKey
	user.UserPublicKeyList = append(user.UserPublicKeyList, user.UserPublicKey)
	user.UpdateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)

	save, err := MarshalMessage(user)
	if err != nil {
		return nil, err
	}

    // key-value :   用户名--用户信息
	app.state.Set(KeyUser(user.UserName), save)

	return &Response{Value: &Response_UserPwdReset{&ResponseUserPwdReset{InstructionId: req.InstructionId}}}, nil
}

//5-检查--企业认证审核通过
func (app *MideaBillApplication) checkEntIdentifyCheck(req *Request) error {
	ent := req.GetEntIdentifyCheck()
	if ent == nil {
		return ErrGetValueIsNull
	}

	// 企业信息都是必填字段，不可为空
	if ent.UserName == "" {
		return ErrUserNameIsNull
	}
	if ent.EntName == "" {
		return ErrEntNameIsNull
	}

	_, _, exists := app.state.Get(KeyUser(ent.UserName))
	if !exists {
		return ErrUserNotExist
	}

	_, _, exists = app.state.Get(KeyEnt(ent.EntCode))
	if exists {
		return ErrEntExist
	}
    
	return nil
}

//5-企业认证审核通过
func (app *MideaBillApplication) entIdentifyCheck(req *Request) (*Response, error) {
	entIdentify := req.GetEntIdentifyCheck()
    if entIdentify == nil {
		return nil, ErrGetValueIsNull
	}

	err := app.checkEntIdentifyCheck(req)
	if err != nil {
		return nil, err
	}

	_, value, exists := app.state.Get(KeyUser(entIdentify.UserName))
	if !exists {
		return nil, ErrUserNotExist
	}

	user := &UserInfo{}
	err = UnmarshalMessage(value, user)
	if err != nil {
		return nil, ErrStorage
	}
    
	user.EntCode = entIdentify.EntCode
	user.EntName = entIdentify.EntName
	user.EntPublicKey = GetEntPublicKey(user.EntCode)
	save, err := MarshalMessage(user)
	if err != nil {
		return nil, err
	}
	app.state.Set(KeyUser(entIdentify.UserName), save)

	ent := &EntInfo{}
	ent.EntCode = entIdentify.EntCode
	ent.EntName = entIdentify.EntName
	ent.EntPublicKey = GetEntPublicKey(ent.EntCode)
	ent.CreateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)
	ent.UpdateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)

	save, err = MarshalMessage(ent)
	if err != nil {
		return nil, err
	}

    // key-value :   组织机构代码--企业信息
	app.state.Set(KeyEnt(ent.EntCode), save) 

	return &Response{Value: &Response_EntIdentifyCheck{&ResponseEntIdentifyCheck{InstructionId: req.InstructionId}}}, nil
}

//6-检查--企业信息修改
func (app *MideaBillApplication) checkEntInfoModify(req *Request) error {
	entInfoModify := req.GetEntInfoModify()
	if entInfoModify == nil {
		return ErrGetValueIsNull
	}

	//当前用户必须是企业管理员
	_, value, exists := app.state.Get(KeyUser(req.UserName))
	if !exists {
		return ErrUserNotExist
	}
	user := &UserInfo{}
	err := UnmarshalMessage(value, user)
	if err != nil {
		return err
	}

	_, _, exists = app.state.Get(KeyEnt(user.EntCode))
	if !exists {
		return ErrEntNotExist
	}

	return nil
}

//6-企业信息修改
func (app *MideaBillApplication) entInfoModify(req *Request) (*Response, error) {
	entInfoModify := req.GetEntInfoModify()
	if entInfoModify == nil {
		return nil, ErrGetValueIsNull
	}

	err := app.checkEntInfoModify(req)
	if err != nil {
		return nil, err
	}

	//查询对应的企业信息
	_, value, exists := app.state.Get(KeyUser(req.UserName))
	if !exists {
		return nil, ErrUserNotExist
	}
	user := &UserInfo{}
	err = UnmarshalMessage(value, user)
	if err != nil {
		return nil, err
	}

	_, valueEnt, exists := app.state.Get(KeyEnt(user.EntCode))
	if !exists {
		return nil, ErrEntNotExist
	}

	ent := &EntInfo{}
	err = UnmarshalMessage(valueEnt, ent)
	if err != nil {
		return nil, ErrStorage
	}

	if entInfoModify.EntName != "" {
		ent.EntName = entInfoModify.EntName
		user.EntName = entInfoModify.EntName
	}

	user.UpdateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)
	save, err := MarshalMessage(user)
	if err != nil {
		return nil, err
	}

    //存入数据库 key-value :   用户名--用户信息
	app.state.Set(KeyUser(user.UserName), save) 

	ent.UpdateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)
	save, err = MarshalMessage(ent)
	if err != nil {
		return nil, err
	}

	//存入数据库 key-value :   组织代码--企业信息
	app.state.Set(KeyEnt(user.EntCode), save) 

	return &Response{Value: &Response_EntInfoModify{&ResponseEntInfoModify{InstructionId: req.InstructionId}}}, nil
}

//7-检查--开票申请
func (app *MideaBillApplication) checkApplyBill(req *Request) error {
	applyBill := req.GetApplyBill()
	if applyBill == nil {
		return ErrGetValueIsNull
	}

	if applyBill.MideaDraftId == "" {
		return ErrMideaDraftIdIsNull
	}
	if applyBill.MideaDraftAmount <= 0 {
		return ErrMideaDraftAmountIsNull
	}
    _, err := time.Parse(formatDate, applyBill.IssueBillDay)
    if err != nil {
    	return ErrIssueBillDayIsErr
    }
	_, err = time.Parse(formatDate, applyBill.ExpireDay)
    if err != nil {
    	return ErrExpireDayIsErr
    }
/*
	if applyBill.PayNum == "" {
		return ErrPayNumIsNull
	}
*/
	if applyBill.RecvBillEntName == "" {
		return ErrRecvBillEntNameIsNull
	}
	if applyBill.RecvBillEntCode == "" {
		return ErrRecvBillEntCodeIsNull
	}

	_, value, exists := app.state.Get(KeyUser(req.UserName))
	if !exists {
		return ErrUserNotExist
	}
	user := &UserInfo{}
	err = UnmarshalMessage(value, user)
	if err != nil {
		return ErrStorage
	}

	//检查开票企业是否为成员企业
	_, value, exists = app.state.Get(KeyEnt(user.EntCode))
	if !exists {
		return ErrEntNotExist 
	}

	ent := &EntInfo{}
	err = UnmarshalMessage(value, ent)
	if err != nil {
		return ErrStorage
	}

	//检查票是否存在
	_, _, exists = app.state.Get(KeyDraft(applyBill.MideaDraftId))
	if exists {
		return ErrBillIsExists
	}

	_, value, exists = app.state.Get(KeyEnt(applyBill.RecvBillEntCode))
	if !exists {
		return ErrEntNotExist 
	}

	ent = &EntInfo{}
	err = UnmarshalMessage(value, ent)
	if err != nil {
		return ErrStorage
	}

    if ent.EntName != applyBill.RecvBillEntName {
    	return ErrEntNameNotMatch 
    }
    if ent.EntCode != applyBill.RecvBillEntCode {
		return ErrEntCodeNotMatch 
    }

	return nil
}

// 开票申请
func (app *MideaBillApplication) applyBill(req *Request) (*Response, error) {
	applyBill := req.GetApplyBill()
	if applyBill == nil {
		return nil, ErrGetValueIsNull
	}

	err := app.checkApplyBill(req)
	if err != nil {
		return nil, err
	}

	//查询开票企业信息
	_, value, exists := app.state.Get(KeyUser(req.UserName))
	if !exists {
		return nil, ErrUserNotExist
	}
	user := &UserInfo{}
	err = UnmarshalMessage(value, user)
	if err != nil {
		return nil, ErrStorage
	}

	//查询收票企业的公钥
	_, value, exists = app.state.Get(KeyEnt(applyBill.RecvBillEntCode))
	if !exists {
		return nil, ErrEntNotExist
	}
	ent := &EntInfo{}
	err = UnmarshalMessage(value, ent)
	if err != nil {
		return nil, ErrStorage
	}

	bill := &MideaBill{}
	bill.MideaDraftId = applyBill.MideaDraftId
	bill.MideaDraftAmount = applyBill.MideaDraftAmount
	bill.IssueBillDay = applyBill.IssueBillDay
	bill.ExpireDay = applyBill.ExpireDay
	bill.IssueBillEntName = user.EntName
	bill.IssueBillEntCode = user.EntCode
	bill.IssueBillPublicKey = user.EntPublicKey
	bill.PayNum = applyBill.PayNum
	bill.RecvBillEntName = applyBill.RecvBillEntName
	bill.RecvBillEntCode = applyBill.RecvBillEntCode
	bill.RecvBillPublicKey = ent.EntPublicKey
	bill.BillState = BillState_BillApplyWaitSign
	bill.CreateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)
	bill.UpdateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)

	save, err := MarshalMessage(bill)
	if err != nil {
		return nil, err
	}
	app.state.Set(KeyDraft(bill.MideaDraftId), save)

	return &Response{Value: &Response_ApplyBill{&ResponseApplyBill{InstructionId: req.InstructionId}}}, nil
}

// 检查--开票签收
func (app *MideaBillApplication) checkApplyBillSign(req *Request) error {
	billSign := req.GetApplyBillSign()
	if billSign == nil {
		return ErrGetValueIsNull
	}

	//检查输入是否为空
	if billSign.MideaDraftId == "" {
		return ErrMideaDraftIdIsNull
	}

	//检查票的'状态'、票的'待收票企业'是不是本用户所在企业
	_, value, exists := app.state.Get(KeyUser(req.UserName))
	if !exists {
		return ErrUserNotExist
	}
	user := &UserInfo{}
	err := UnmarshalMessage(value, user)
	if err != nil {
		return ErrStorage
	}

	_, value, exists = app.state.Get(KeyDraft(billSign.MideaDraftId))
	if !exists {
		return ErrBillIsNotExists
	}

	bill := &MideaBill{}
	err = UnmarshalMessage(value, bill)
	if err != nil {
		return ErrStorage
	}

	if !bytes.Equal(bill.RecvBillPublicKey, user.EntPublicKey) {
		return ErrRecvBillPublicKeyNotMatch
	}
	if bill.BillState != BillState_BillApplyWaitSign {
		return ErrBillStateNotMatch
	}

	return nil
}

// 开票签收
func (app *MideaBillApplication) applyBillSign(req *Request) (*Response, error) {
	applyBill := req.GetApplyBillSign()
	if applyBill == nil {
		return nil, ErrGetValueIsNull
	}

	err := app.checkApplyBillSign(req)
	if err != nil {
		return nil, err
	}

	//通过票号查询票信息
	_, value, exists := app.state.Get(KeyDraft(applyBill.MideaDraftId))
	if !exists {
		return nil, ErrBillIsNotExists
	}
	bill := &MideaBill{}
	err = UnmarshalMessage(value, bill)
	if err != nil {
		return nil, ErrStorage
	}

	bill.SignDay = time.Unix(req.OperatorTime, 0).Format(formatDate)
	bill.BillState = BillState_BillNormalOwn
	bill.UpdateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)

	save, err := MarshalMessage(bill)
	if err != nil {
		return nil, err
	}
	app.state.Set(KeyDraft(bill.MideaDraftId), save)

	return &Response{Value: &Response_ApplyBillSign{&ResponseApplyBillSign{InstructionId: req.InstructionId}}}, nil
}

// 检查--开票拒绝
func (app *MideaBillApplication) checkApplyBillSignRefuse(req *Request) error {
	billSignRefuse := req.GetApplyBillSignRefuse()
	if billSignRefuse == nil {
		return ErrGetValueIsNull
	}

	//检查输入是否为空
	if billSignRefuse.MideaDraftId == "" {
		return ErrMideaDraftIdIsNull
	}

	//检查票的'状态'、票的'待收票企业'是不是本用户所在企业
	_, value, exists := app.state.Get(KeyUser(req.UserName))
	if !exists {
		return ErrUserNotExist
	}
	user := &UserInfo{}
	err := UnmarshalMessage(value, user)
	if err != nil {
		return ErrStorage
	}

	_, value, exists = app.state.Get(KeyDraft(billSignRefuse.MideaDraftId))
	if !exists {
		return ErrBillIsNotExists
	}

	bill := &MideaBill{}
	err = UnmarshalMessage(value, bill)
	if err != nil {
		return ErrStorage
	}

	if !bytes.Equal(bill.RecvBillPublicKey, user.EntPublicKey) {
		return ErrRecvBillPublicKeyNotMatch
	}
	if bill.BillState != BillState_BillApplyWaitSign {
		return ErrBillStateNotMatch
	}

	return nil
}

// 开票拒绝
func (app *MideaBillApplication) applyBillSignRefuse(req *Request) (*Response, error) {
	billSignRefuse := req.GetApplyBillSignRefuse()
	if billSignRefuse == nil {
		return nil, ErrGetValueIsNull
	}

	err := app.checkApplyBillSignRefuse(req)
	if err != nil {
		return nil, err
	}

	//通过票号查询票信息
	_, value, exists := app.state.Get(KeyDraft(billSignRefuse.MideaDraftId))
	if !exists {
		return nil, ErrBillIsNotExists
	}
	bill := &MideaBill{}
	err = UnmarshalMessage(value, bill)
	if err != nil {
		return nil, ErrStorage
	}

	bill.BillState = BillState_BillCancer
	bill.UpdateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)

	save, err := MarshalMessage(bill)
	if err != nil {
		return nil, err
	}
	app.state.Set(KeyDraft(bill.MideaDraftId), save)

	return &Response{Value: &Response_ApplyBillSignRefuse{&ResponseApplyBillSignRefuse{InstructionId: req.InstructionId}}}, nil
}

// 检查--开票待签收撤回
func (app *MideaBillApplication) checkApplyBillSignCancle(req *Request) error {
	billSignCancle := req.GetApplyBillSignCancle()
	if billSignCancle == nil {
		return ErrGetValueIsNull
	}

	//检查输入是否为空
	if billSignCancle.MideaDraftId == "" {
		return ErrMideaDraftIdIsNull
	}

	//检查票的'状态'、票的'待收票企业'是不是本用户所在企业
	_, value, exists := app.state.Get(KeyUser(req.UserName))
	if !exists {
		return ErrUserNotExist
	}
	user := &UserInfo{}
	err := UnmarshalMessage(value, user)
	if err != nil {
		return ErrStorage
	}

	_, value, exists = app.state.Get(KeyDraft(billSignCancle.MideaDraftId))
	if !exists {
		return ErrBillIsNotExists
	}

	bill := &MideaBill{}
	err = UnmarshalMessage(value, bill)
	if err != nil {
		return ErrStorage
	}

	if !bytes.Equal(bill.IssueBillPublicKey, user.EntPublicKey) {
		return ErrIssueBillPublicKeyNotMatch
	}
	if bill.BillState != BillState_BillApplyWaitSign {
		return ErrBillStateNotMatch
	}

	return nil
}

// 开票待签收撤回
func (app *MideaBillApplication) applyBillSignCancle(req *Request) (*Response, error) {
	applyBillSignCancle := req.GetApplyBillSignCancle()
	if applyBillSignCancle == nil {
		return nil, ErrGetValueIsNull
	}

	err := app.checkApplyBillSignCancle(req)
	if err != nil {
		return nil, err
	}

	//通过票号查询票信息
	_, value, exists := app.state.Get(KeyDraft(applyBillSignCancle.MideaDraftId))
	if !exists {
		return nil, ErrBillIsNotExists
	}
	bill := &MideaBill{}
	err = UnmarshalMessage(value, bill)
	if err != nil {
		return nil, ErrStorage
	}

	bill.BillState = BillState_BillCancer
	bill.UpdateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)

	save, err := MarshalMessage(bill)
	if err != nil {
		return nil, err
	}
	app.state.Set(KeyDraft(bill.MideaDraftId), save)

	return &Response{Value: &Response_ApplyBillSignCancle{&ResponseApplyBillSignCancle{InstructionId: req.InstructionId}}}, nil
}

// 检查--整转
func (app *MideaBillApplication) checkBillTotalTransfer(req *Request) error {
	billTotalTransfer := req.GetBillTotalTransfer()
	if billTotalTransfer == nil {
		return ErrGetValueIsNull
	}

	//检查输入是否为空
	if billTotalTransfer.MideaDraftId == "" {
		return ErrMideaDraftIdIsNull
	}
    if billTotalTransfer.WaitRecvBillEntCode == "" {
		return ErrWaitRecvBillEntCodeIsNull
	}

	//检查票的'状态'、票的'收票企业'是不是本用户所在企业
	_, value, exists := app.state.Get(KeyUser(req.UserName))
	if !exists {
		return ErrUserNotExist
	}

	user := &UserInfo{}
	err := UnmarshalMessage(value, user)
	if err != nil {
		return ErrStorage
	}

	_, value, exists = app.state.Get(KeyDraft(billTotalTransfer.MideaDraftId))
	if !exists {
		return ErrBillIsNotExists
	}

	bill := &MideaBill{}
	err = UnmarshalMessage(value, bill)
	if err != nil {
		return ErrStorage
	}

	if !bytes.Equal(bill.RecvBillPublicKey, user.EntPublicKey) {
		return ErrRecvBillPublicKeyNotMatch
	}
    if bill.BillState != BillState_BillNormalOwn {
		return ErrBillStateNotMatch
	}

	_, value, exists = app.state.Get(KeyEnt(billTotalTransfer.WaitRecvBillEntCode))
	if !exists {
		return ErrEntNotExist
	}
	ent := &EntInfo{}
	err = UnmarshalMessage(value, ent)
	if err != nil {
		return ErrStorage
	}

	if billTotalTransfer.WaitRecvBillEntCode != ent.EntCode {
		return ErrWaitRecvBillEntCodeIsNull
	}

	return nil
}

// 整转
func (app *MideaBillApplication) billTotalTransfer(req *Request) (*Response, error) {
	billTotalTransfer := req.GetBillTotalTransfer()
	if billTotalTransfer == nil {
		return nil, ErrGetValueIsNull
	}

	err := app.checkBillTotalTransfer(req)
	if err != nil {
		return nil, err
	}

	//通过票号查询票信息
	_, value, exists := app.state.Get(KeyDraft(billTotalTransfer.MideaDraftId))
	if !exists {
		return nil, ErrBillIsNotExists 
	}

	bill := &MideaBill{}
	err = UnmarshalMessage(value, bill)
	if err != nil {
		return nil, ErrStorage
	}

	bill.WaitRecvBillEntCode = billTotalTransfer.WaitRecvBillEntCode

	//通过组织代码查询待收票企业公钥
	_, value, exists = app.state.Get(KeyEnt(billTotalTransfer.WaitRecvBillEntCode))
	if !exists {
		return nil, ErrEntNotExist
	}
	ent := &EntInfo{}
	err = UnmarshalMessage(value, ent)
	if err != nil {
		return nil, ErrStorage
	}

	bill.WaitRecvBillPublicKey = ent.EntPublicKey
	bill.BillState = BillState_BillTransferWaitSign
	bill.UpdateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)

	save, err := MarshalMessage(bill)
	if err != nil {
		return nil, err
	}
	app.state.Set(KeyDraft(bill.MideaDraftId), save)

	return &Response{Value: &Response_BillTotalTransfer{&ResponseBillTotalTransfer{InstructionId: req.InstructionId}}}, nil
}

// 检查--转让拆分
func (app *MideaBillApplication) checkBillPartTransfer(req *Request) error {
	billPartTransfer := req.GetBillPartTransfer()
	if billPartTransfer == nil {
		return ErrGetValueIsNull
	}

	if billPartTransfer.MideaDraftId == "" {
		return ErrMideaDraftIdIsNull
	}

	billSubList := billPartTransfer.GetSub()
	if len(billSubList) <= 0 {
		return ErrBillSubListIsNull
	}

	//检查票的'状态'、票的'收票企业'是不是本用户所在企业
	_, value, exists := app.state.Get(KeyUser(req.UserName))
	if !exists {
		return ErrUserNotExist
	}
	user := &UserInfo{}
	err := UnmarshalMessage(value, user)
	if err != nil {
		return ErrStorage
	}

	_, value, exists = app.state.Get(KeyDraft(billPartTransfer.MideaDraftId))
	if !exists {
		return ErrBillIsNotExists
	}
	bill := &MideaBill{}
	err = UnmarshalMessage(value, bill)
	if err != nil {
		return ErrStorage
	}

	if !bytes.Equal(bill.RecvBillPublicKey, user.EntPublicKey) {
		return ErrRecvBillPublicKeyNotMatch
	}
    if bill.BillState != BillState_BillNormalOwn {
		return ErrBillStateNotMatch
	}

    var draftAmount int64

    for _, subBill := range billSubList {
    	_, _, exists := app.state.Get(KeyDraft(subBill.MideaDraftId))
    	if exists {
    		return ErrBillIsExists
    	}
        _, value, exists = app.state.Get(KeyEnt(subBill.WaitRecvBillEntCode))
        if !exists {
                return ErrEntNotExist
        }
        ent := &EntInfo{}
        err := UnmarshalMessage(value, ent)
        if err != nil {
            return ErrStorage
        }

        if subBill.MideaDraftId == "" {
        	return ErrMideaDraftIdIsNull
        }
        if subBill.MideaDraftAmount <= 0 {
        	return ErrMideaDraftAmountIsNull
        }
	    if subBill.WaitRecvBillEntCode != ent.EntCode {
	    	return ErrWaitRecvBillEntCodeIsNull
        }
        draftAmount += subBill.MideaDraftAmount
    }

    if bill.MideaDraftAmount != draftAmount {
	    return ErrMideaDraftAmountNotMatch
    }

	return nil
}

// 转让拆分
func (app *MideaBillApplication) billPartTransfer(req *Request) (*Response, error) {
	billPartTransfer := req.GetBillPartTransfer()
	if billPartTransfer == nil {
		return nil, ErrGetValueIsNull
	}

	err := app.checkBillPartTransfer(req)
	if err != nil {
		return nil, err
	}

	//通过票号查询母票信息
	_, value, exists := app.state.Get(KeyDraft(billPartTransfer.MideaDraftId))
	if !exists {
		return nil, ErrBillIsNotExists
	}
	bill := &MideaBill{}
	err = UnmarshalMessage(value, bill)
	if err != nil {
		return nil, ErrStorage
	}

	for _, w := range billPartTransfer.GetSub() {
		//通过票号查询子票组织结构代码
		_, value, exists := app.state.Get(KeyEnt(w.WaitRecvBillEntCode))
		if !exists {
			return nil, ErrEntNotExist
		}
		ent := &EntInfo{}
		err := UnmarshalMessage(value, ent)
		if err != nil {
			return nil, ErrStorage
		}

		//记录母票的每一个子票票号
		bill.NextMideaDraftId = append(bill.NextMideaDraftId, w.MideaDraftId)

		//每次循环创建一张新票
		newBill := &MideaBill{}
		newBill.MideaDraftId = w.MideaDraftId
		newBill.MideaDraftAmount = w.MideaDraftAmount
		newBill.IssueBillDay			= bill.IssueBillDay
		newBill.ExpireDay				= bill.ExpireDay
		newBill.IssueBillEntName		= bill.IssueBillEntName
		newBill.IssueBillEntCode		= bill.IssueBillEntCode
		newBill.IssueBillPublicKey		= bill.IssueBillPublicKey
		newBill.PayNum					= bill.PayNum
		newBill.RecvBillEntName			= bill.RecvBillEntName		//收票企业暂时和母票一样
		newBill.RecvBillEntCode			= bill.RecvBillEntCode
		newBill.RecvBillPublicKey		= bill.RecvBillPublicKey
		
		// 判断待收票企业等于原收票企业，则是自己持有，状态为 正常持有
		if bytes.Equal(bill.RecvBillPublicKey, ent.EntPublicKey) {
			newBill.SignDay = time.Unix(req.OperatorTime, 0).Format(formatDate)
			newBill.BillState = BillState_BillNormalOwn
		} else {
			newBill.WaitRecvBillEntName = ent.EntName
			newBill.WaitRecvBillEntCode = ent.EntCode
			newBill.WaitRecvBillPublicKey = ent.EntPublicKey
			newBill.BillState = BillState_BillTransferWaitSign
		}
		
		newBill.PreMideaDraftId = bill.MideaDraftId //上一级美汇票号--母票票号
		newBill.CreateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)
		newBill.UpdateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)

		//存入子票
		save, err := MarshalMessage(newBill)
		if err != nil {
			return nil, err
		}
		app.state.Set(KeyDraft(w.MideaDraftId), save)
	}

	//对母票进行操作   状态置为--已拆转
	bill.BillState = BillState_BillTransferOk
	bill.UpdateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)
	//修改原母票信息并存入数据库
	save, err := MarshalMessage(bill)
	if err != nil {
		return nil, err
	}
	app.state.Set(KeyDraft(bill.MideaDraftId), save)

	return &Response{Value: &Response_BillPartTransfer{&ResponseBillPartTransfer{InstructionId: req.InstructionId}}}, nil
}

// 检查--转让签收
func (app *MideaBillApplication) checkBillTransferSign(req *Request) error {
	billTransferSign := req.GetBillTransferSign()
	if billTransferSign == nil {
		return ErrGetValueIsNull
	}

	if billTransferSign.MideaDraftId == "" {
		return ErrMideaDraftIdIsNull
	}

	_, value, exists := app.state.Get(KeyUser(req.UserName))
	if !exists {
		return ErrUserNotExist
	}
	user := &UserInfo{}
	err := UnmarshalMessage(value, user)
	if err != nil {
		return ErrStorage
	}

	_, value, exists = app.state.Get(KeyDraft(billTransferSign.MideaDraftId))
	if !exists {
		return ErrBillIsNotExists
	}
	bill := &MideaBill{}
	err = UnmarshalMessage(value, bill)
	if err != nil {
		return ErrStorage
	}

	if !bytes.Equal(bill.WaitRecvBillPublicKey, user.EntPublicKey) {
		return ErrWaitRecvBillPublicKeyNotMatch
	}
    if bill.BillState != BillState_BillTransferWaitSign {
		return ErrBillStateNotMatch
	}

	return nil
}

// 转让签收
func (app *MideaBillApplication) billTransferSign(req *Request) (*Response, error) {
	billTransferSign := req.GetBillTransferSign()
	if billTransferSign == nil {
		return nil, ErrGetValueIsNull
	}

	err := app.checkBillTransferSign(req)
	if err != nil {
		return nil, err
	}

	//通过票号查询票信息
	_, value, exists := app.state.Get(KeyDraft(billTransferSign.MideaDraftId))
	if !exists {
		return nil, ErrBillIsNotExists
	}
	bill := &MideaBill{}
	err = UnmarshalMessage(value, bill)
	if err != nil {
		return nil, ErrStorage
	}

	bill.SignDay = time.Unix(req.OperatorTime, 0).Format(formatDate)
	//'收票企业'赋值给‘原收票企业’
	bill.LastBillEntName = bill.RecvBillEntName
	bill.LastBillEntCode = bill.RecvBillEntCode
	bill.LastBillPublicKey = bill.RecvBillPublicKey
	//'待收票企业'赋值给‘收票企业’
	bill.RecvBillEntName = bill.WaitRecvBillEntName
	bill.RecvBillEntCode = bill.WaitRecvBillEntCode
	bill.RecvBillPublicKey = bill.WaitRecvBillPublicKey
	//'待收票企业'字段清空
	bill.WaitRecvBillEntName = ""
	bill.WaitRecvBillEntCode = ""
	bill.WaitRecvBillPublicKey = nil

	bill.BillState = BillState_BillNormalOwn
	bill.UpdateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)

	save, err := MarshalMessage(bill)
	if err != nil {
		return nil, err
	}
	app.state.Set(KeyDraft(bill.MideaDraftId), save)

	return &Response{Value: &Response_BillTransferSign{&ResponseBillTransferSign{InstructionId: req.InstructionId}}}, nil
}

// 检查--转让拒绝
func (app *MideaBillApplication) checkBillTransferRefuse(req *Request) error {
	billTransferRefuse := req.GetBillTransferRefuse()
	if billTransferRefuse == nil {
		return ErrGetValueIsNull
	}

	if billTransferRefuse.MideaDraftId == "" {
		return ErrMideaDraftIdIsNull
	}

	//检查票的'状态'、票的'待收票企业'是不是本用户所在企业
	_, value, exists := app.state.Get(KeyUser(req.UserName))
	if !exists {
		return ErrUserNotExist
	}
	user := &UserInfo{}
	err := UnmarshalMessage(value, user)
	if err != nil {
		return ErrStorage
	}

	_, value, exists = app.state.Get(KeyDraft(billTransferRefuse.MideaDraftId))
	if !exists {
		return ErrBillIsNotExists
	}
	bill := &MideaBill{}
	err = UnmarshalMessage(value, bill)
	if err != nil {
		return ErrStorage
	}

	if !bytes.Equal(bill.WaitRecvBillPublicKey, user.EntPublicKey) {
		return ErrWaitRecvBillPublicKeyNotMatch
	}
    if bill.BillState != BillState_BillTransferWaitSign {
		return ErrBillStateNotMatch
	}

	return nil
}

// 转让拒绝
func (app *MideaBillApplication) billTransferRefuse(req *Request) (*Response, error) {
	billTransferRefuse := req.GetBillTransferRefuse()
	if billTransferRefuse == nil {
		return nil, ErrGetValueIsNull
	}

	err := app.checkBillTransferRefuse(req)
	if err != nil {
		return nil, err
	}

	//通过票号查询票信息
	_, value, exists := app.state.Get(KeyDraft(billTransferRefuse.MideaDraftId))
	if !exists {
		return nil, ErrBillIsNotExists
	}
	bill := &MideaBill{}
	err = UnmarshalMessage(value, bill)
	if err != nil {
		return nil, ErrStorage
	}

	//'待收票企业'字段清空
	bill.WaitRecvBillEntName = ""
	bill.WaitRecvBillEntCode = ""
	bill.WaitRecvBillPublicKey = nil
	bill.BillState = BillState_BillNormalOwn
	bill.UpdateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)

	save, err := MarshalMessage(bill)
	if err != nil {
		return nil, err
	}
	app.state.Set(KeyDraft(bill.MideaDraftId), save)

	return &Response{Value: &Response_BillTransferRefuse{&ResponseBillTransferRefuse{InstructionId: req.InstructionId}}}, nil
}

// 检查-转让待签收撤回
func (app *MideaBillApplication) checkBillTransferCancle(req *Request) error {
	billTransferCancle := req.GetBillTransferCancle()
	if billTransferCancle == nil {
		return ErrGetValueIsNull
	}

	if billTransferCancle.MideaDraftId == "" {
		return ErrMideaDraftIdIsNull
	}

	//检查票的'状态'、票的'待收票企业'是不是本用户所在企业
	_, value, exists := app.state.Get(KeyUser(req.UserName))
	if !exists {
		return ErrUserNotExist
	}
	user := &UserInfo{}
	err := UnmarshalMessage(value, user)
	if err != nil {
		return ErrStorage
	}

	_, value, exists = app.state.Get(KeyDraft(billTransferCancle.MideaDraftId))
	if !exists {
		return ErrBillIsNotExists
	}
	bill := &MideaBill{}
	err = UnmarshalMessage(value, bill)
	if err != nil {
		return ErrStorage
	}

	if !bytes.Equal(bill.RecvBillPublicKey, user.EntPublicKey) {
		return ErrRecvBillPublicKeyNotMatch
	}
    if bill.BillState != BillState_BillTransferWaitSign {
		return ErrBillStateNotMatch
	}

	return nil
}

// 转让待签收撤回
func (app *MideaBillApplication) billTransferCancle(req *Request) (*Response, error) {
	billTransferCancle := req.GetBillTransferCancle()
	if billTransferCancle == nil {
		return nil, ErrGetValueIsNull
	}

	err := app.checkBillTransferCancle(req)
	if err != nil {
		return nil, err
	}

	//通过票号查询票信息
	_, value, exists := app.state.Get(KeyDraft(billTransferCancle.MideaDraftId))
	if !exists {
		return nil, ErrBillIsNotExists
	}
	bill := &MideaBill{}
	err = UnmarshalMessage(value, bill)
	if err != nil {
		return nil, ErrStorage
	}

	//'待收票企业'字段清空
	bill.WaitRecvBillEntName = ""
	bill.WaitRecvBillEntCode = ""
	bill.WaitRecvBillPublicKey = nil
	bill.BillState = BillState_BillNormalOwn
	bill.UpdateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)

	save, err := MarshalMessage(bill)
	if err != nil {
		return nil, err
	}
	app.state.Set(KeyDraft(bill.MideaDraftId), save)

	return &Response{Value: &Response_BillTransferCancle{&ResponseBillTransferCancle{InstructionId: req.InstructionId}}}, nil
}

// 检查--转让到期未签收兑付
func (app *MideaBillApplication) checkBillTransferForcePay(req *Request) error {
	billTransferForcePay := req.GetBillTransferForcePay()
	if billTransferForcePay == nil {
		return ErrGetValueIsNull
	}

	if billTransferForcePay.MideaDraftId == "" {
		return ErrMideaDraftIdIsNull
	}

	_, value, exists := app.state.Get(KeyUser(req.UserName))
	if !exists {
	   return ErrUserNotExist
	}
	user := &UserInfo{}
	err := UnmarshalMessage(value, user)
	if err != nil{
	   return ErrStorage
	}

	// 检查票的'状态'
	_, value, exists = app.state.Get(KeyDraft(billTransferForcePay.MideaDraftId))
	if !exists {
		return ErrBillIsNotExists
	}
	bill := &MideaBill{}
	err = UnmarshalMessage(value, bill)
	if err != nil {
		return ErrStorage
	}

	if strings.Compare(bill.ExpireDay, time.Now().Format(formatDate)) < 0 {
		return ErrExpireDayNotMatch
	}
    if bill.BillState != BillState_BillTransferWaitSign {
		return ErrBillStateNotMatch
	}

	return nil
}

// 转让到期未签收兑付
func (app *MideaBillApplication) billTransferForcePay(req *Request) (*Response, error) {
	billTransferForcePay := req.GetBillTransferForcePay()
	if billTransferForcePay == nil {
		return nil, ErrGetValueIsNull
	}

	err := app.checkBillTransferForcePay(req)
	if err != nil {
		return nil, err
	}

	//通过票号查询票信息
	_, value, exists := app.state.Get(KeyDraft(billTransferForcePay.MideaDraftId))
	if !exists {
		return nil, ErrBillIsNotExists
	}
	bill := &MideaBill{}
	err = UnmarshalMessage(value, bill)
	if err != nil {
		return nil, ErrStorage
	}

	// 收票企业赋值给原收票企业
	bill.LastBillEntName = bill.RecvBillEntName
	bill.LastBillEntCode = bill.RecvBillEntCode
	bill.LastBillPublicKey = bill.RecvBillPublicKey
	// 待收票企业赋值给收票企业
	bill.RecvBillEntName = bill.WaitRecvBillEntName
	bill.RecvBillEntCode = bill.WaitRecvBillEntCode
	bill.RecvBillPublicKey = bill.WaitRecvBillPublicKey
	// 待收票企业字段清空
	bill.WaitRecvBillEntName = ""
	bill.WaitRecvBillEntCode = ""
	bill.WaitRecvBillPublicKey = nil
	// 状态置为'已兑付'
	bill.BillState = BillState_BillPaid
	bill.PayDay = time.Unix(req.OperatorTime, 0).Format(formatDate)
	bill.UpdateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)

	save, err := MarshalMessage(bill)
	if err != nil {
		return nil, err
	}
	app.state.Set(KeyDraft(bill.MideaDraftId), save)

	return &Response{Value: &Response_BillTransferForcePay{&ResponseBillTransferForcePay{InstructionId: req.InstructionId}}}, nil
}

//15-检查--整融
func (app *MideaBillApplication) checkBillTotalFinancing(req *Request) error {
	billTotalFinancing := req.GetBillTotalFinancing()
	if billTotalFinancing == nil {
		return ErrGetValueIsNull
	}

	if billTotalFinancing.MideaDraftId == "" {
		return ErrMideaDraftIdIsNull
	}

	//检查票的'状态'、票的'收票企业'是不是本用户所在企业
	_, value, exists := app.state.Get(KeyUser(req.UserName))
	if !exists {
		return ErrUserNotExist
	}
	user := &UserInfo{}
	err := UnmarshalMessage(value, user)
	if err != nil {
		return ErrStorage
	}

	_, value, exists = app.state.Get(KeyDraft(billTotalFinancing.MideaDraftId))
	if !exists {
		return ErrBillIsNotExists
	}
	bill := &MideaBill{}
	err = UnmarshalMessage(value, bill)
	if err != nil {
		return ErrStorage
	}

	if !bytes.Equal(bill.RecvBillPublicKey, user.EntPublicKey) {
		return ErrRecvBillPublicKeyNotMatch
	}
    if bill.BillState != BillState_BillNormalOwn {
		return ErrBillStateNotMatch
	}

	_, value, exists = app.state.Get(KeyEnt(billTotalFinancing.WaitRecvBillEntCode))
	if !exists {
		return ErrEntNotExist
	}
	ent := &EntInfo{}
	err = UnmarshalMessage(value, ent)
	if err != nil {
		return ErrStorage
	}

	if billTotalFinancing.WaitRecvBillEntCode != ent.EntCode {
		return ErrWaitRecvBillEntCodeNotMatch
	}

	return nil
}

// 整融
func (app *MideaBillApplication) billTotalFinancing(req *Request) (*Response, error) {
	billTotalFinancing := req.GetBillTotalFinancing()
	if billTotalFinancing == nil {
		return nil, ErrGetValueIsNull
	}

	err := app.checkBillTotalFinancing(req)
	if err != nil {
		return nil, err
	}

	//通过票号查询票信息
	_, value, exists := app.state.Get(KeyDraft(billTotalFinancing.MideaDraftId))
	if !exists {
		return nil, ErrBillIsNotExists
	}
	bill := &MideaBill{}
	err = UnmarshalMessage(value, bill)
	if err != nil {
		return nil, ErrStorage
	}

	bill.WaitRecvBillEntCode = billTotalFinancing.WaitRecvBillEntCode

	//通过组织代码查询待收票企业公钥
	_, value, exists = app.state.Get(KeyEnt(billTotalFinancing.WaitRecvBillEntCode))
	if !exists {
		return nil, ErrEntNotExist
	}
	ent := &EntInfo{}
	err = UnmarshalMessage(value, ent)
	if err != nil {
		return nil, ErrStorage
	}

	bill.WaitRecvBillPublicKey = ent.EntPublicKey
	bill.BillState = BillState_BillFinancingWaitCheck
	bill.UpdateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)

	save, err := MarshalMessage(bill)
	if err != nil {
		return nil, err
	}
	app.state.Set(KeyDraft(bill.MideaDraftId), save)

	return &Response{Value: &Response_BillTotalFinancing{&ResponseBillTotalFinancing{InstructionId: req.InstructionId}}}, nil
}

// 检查--拆融
func (app *MideaBillApplication) checkBillPartFinancing(req *Request) error {
	billPartFinancing := req.GetBillPartFinancing()
	if billPartFinancing == nil {
		return ErrGetValueIsNull
	}

	if billPartFinancing.MideaDraftId == "" {
		return ErrMideaDraftIdIsNull
	}

	billSubList := billPartFinancing.GetSub()
	if len(billSubList) <= 0 {
		return ErrBillSubListIsNull
	}

	//检查票的'状态'、票的'收票企业'是不是本用户所在企业
	_, value, exists := app.state.Get(KeyUser(req.UserName))
	if !exists {
		return ErrUserNotExist
	}
	user := &UserInfo{}
	err := UnmarshalMessage(value, user)
	if err != nil {
		return ErrStorage
	}

	_, value, exists = app.state.Get(KeyDraft(billPartFinancing.MideaDraftId))
	if !exists {
		return ErrBillIsNotExists
	}
	bill := &MideaBill{}
	err = UnmarshalMessage(value, bill)
	if err != nil {
		return ErrStorage
	}

	if !bytes.Equal(bill.RecvBillPublicKey, user.EntPublicKey) {
		return ErrRecvBillPublicKeyNotMatch
	}
    if bill.BillState != BillState_BillNormalOwn {
		return ErrBillStateNotMatch
	}

    var draftAmount int64

    for _, subBill := range billSubList {
    	_, _, exists := app.state.Get(KeyDraft(subBill.MideaDraftId))
    	if exists {
    		return ErrBillIsExists
    	}
        _, value, exists = app.state.Get(KeyEnt(subBill.WaitRecvBillEntCode))
        if !exists {
                return ErrEntNotExist
        }
        ent := &EntInfo{}
        err := UnmarshalMessage(value, ent)
        if err != nil {
            return ErrStorage
        }

        if subBill.MideaDraftId == "" {
        	return ErrMideaDraftIdIsNull
        }
        if subBill.MideaDraftAmount <= 0 {
        	return ErrMideaDraftAmountIsNull
        }
	    if subBill.WaitRecvBillEntCode != ent.EntCode {
	    	return ErrWaitRecvBillEntCodeNotMatch
        }
        draftAmount += subBill.MideaDraftAmount
    }

    if bill.MideaDraftAmount != draftAmount {
	    return ErrMideaDraftAmountNotMatch
    }

	return nil
}

//16-拆融
func (app *MideaBillApplication) billPartFinancing(req *Request) (*Response, error) {
	billPartFinancing := req.GetBillPartFinancing()
	if billPartFinancing == nil {
		return nil, ErrGetValueIsNull
	}

	err := app.checkBillPartFinancing(req)
	if err != nil {
		return nil, err
	}

	//通过票号查询母票信息
	_, value, exists := app.state.Get(KeyDraft(billPartFinancing.MideaDraftId))
	if !exists {
		return nil, ErrBillIsNotExists
	}
	bill := &MideaBill{}
	err = UnmarshalMessage(value, bill)
	if err != nil {
		return nil, ErrStorage
	}

	for _, w := range billPartFinancing.GetSub() {
		//通过票号查询子票组织结构代码
		_, value, exists := app.state.Get(KeyEnt(w.WaitRecvBillEntCode))
		if !exists {
			return nil, ErrEntNotExist
		}
		ent := &EntInfo{}
		err := UnmarshalMessage(value, ent)
		if err != nil {
			return nil, ErrStorage
		}

		//记录母票的每一个子票票号
		bill.NextMideaDraftId = append(bill.NextMideaDraftId, w.MideaDraftId)

		//每次循环创建一张新票
		newBill := &MideaBill{}
		newBill.MideaDraftId = w.MideaDraftId
		newBill.MideaDraftAmount = w.MideaDraftAmount
		newBill.IssueBillDay			= bill.IssueBillDay
		newBill.ExpireDay				= bill.ExpireDay
		newBill.IssueBillEntName		= bill.IssueBillEntName
		newBill.IssueBillEntCode		= bill.IssueBillEntCode
		newBill.IssueBillPublicKey		= bill.IssueBillPublicKey
		newBill.PayNum					= bill.PayNum
		newBill.RecvBillEntName			= bill.RecvBillEntName		//收票企业暂时和母票一样
		newBill.RecvBillEntCode			= bill.RecvBillEntCode
		newBill.RecvBillPublicKey		= bill.RecvBillPublicKey

		// 判断待收票企业等于原收票企业，则是自己持有，状态为 正常持有
		if bytes.Equal(bill.RecvBillPublicKey, ent.EntPublicKey) {
			newBill.SignDay = time.Unix(req.OperatorTime, 0).Format(formatDate)
			newBill.BillState = BillState_BillNormalOwn
		} else {
			newBill.WaitRecvBillEntName = ent.EntName
			newBill.WaitRecvBillEntCode = ent.EntCode
			newBill.WaitRecvBillPublicKey = ent.EntPublicKey
			newBill.BillState = BillState_BillFinancingWaitCheck
		}
		
		newBill.PreMideaDraftId = bill.MideaDraftId //上一级美汇票号--母票票号
		newBill.CreateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)
		newBill.UpdateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)

		//存入子票
		save, err := MarshalMessage(newBill)
		if err != nil {
			return nil, err
		}
		app.state.Set(KeyDraft(w.MideaDraftId), save)
	}

	//对母票进行操作   状态置为--已拆融
	bill.BillState = BillState_BillFinancingOk
	//修改原母票信息并存入数据库
	save, err := MarshalMessage(bill)
	if err != nil {
		return nil, err
	}
	app.state.Set(KeyDraft(bill.MideaDraftId), save)

	return &Response{Value: &Response_BillPartFinancing{&ResponseBillPartFinancing{InstructionId: req.InstructionId}}}, nil
}

//17-检查--融资审核通过
func (app *MideaBillApplication) checkBillFinancingCheckOk(req *Request) error {
	billFinancingCheckOk := req.GetBillFinancingCheckOk()
	if billFinancingCheckOk == nil {
		return ErrGetValueIsNull
	}

	if billFinancingCheckOk.MideaDraftId == "" {
		return ErrMideaDraftIdIsNull
	}
	if req.UserName != SdkUser {
		return ErrUserNameIsNotSDKUser
	}

	_, value, exists := app.state.Get(KeyDraft(billFinancingCheckOk.MideaDraftId))
	if !exists {
		return ErrBillIsNotExists
	}
	bill := &MideaBill{}
	err := UnmarshalMessage(value, bill)
	if err != nil {
		return ErrStorage
	}

	if bill.BillState != BillState_BillFinancingWaitCheck {
		return ErrBillStateNotMatch
	}

	return nil
}

//17-融资审核通过
func (app *MideaBillApplication) billFinancingCheckOk(req *Request) (*Response, error) {
	billFinancingCheckOk := req.GetBillFinancingCheckOk()
	if billFinancingCheckOk == nil {
		return nil, ErrGetValueIsNull
	}

	err := app.checkBillFinancingCheckOk(req)
	if err != nil {
		return nil, err
	}

	//通过票号查询票信息
	_, value, exists := app.state.Get(KeyDraft(billFinancingCheckOk.MideaDraftId))
	if !exists {
		return nil, ErrBillIsNotExists
	}
	bill := &MideaBill{}
	err = UnmarshalMessage(value, bill)
	if err != nil {
		return nil, ErrStorage
	}

	bill.SignDay = time.Unix(req.OperatorTime, 0).Format(formatDate)
	//'收票企业'赋值给'原收票企业'
	bill.LastBillEntName = bill.RecvBillEntName
	bill.LastBillEntCode = bill.RecvBillEntCode
	bill.LastBillPublicKey = bill.RecvBillPublicKey
	//'待收票企业'赋值给‘收票企业’
	bill.RecvBillEntName = bill.WaitRecvBillEntName
	bill.RecvBillEntCode = bill.WaitRecvBillEntCode
	bill.RecvBillPublicKey = bill.WaitRecvBillPublicKey
	//'待收票企业'字段清空
	bill.WaitRecvBillEntName = ""
	bill.WaitRecvBillEntCode = ""
	bill.WaitRecvBillPublicKey = nil

	bill.BillState = BillState_BillNormalOwn
	bill.UpdateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)

	save, err := MarshalMessage(bill)
	if err != nil {
		return nil, err
	}
	app.state.Set(KeyDraft(bill.MideaDraftId), save)

	return &Response{Value: &Response_BillFinancingCheckOk{&ResponseBillFinancingCheckOk{InstructionId: req.InstructionId}}}, nil
}

//18-检查--融资审核拒绝
func (app *MideaBillApplication) checkBillFinancingCheckFail(req *Request) error {
	billFinancingCheckFail := req.GetBillFinancingCheckFail()
	if billFinancingCheckFail == nil {
		return ErrGetValueIsNull
	}

	if billFinancingCheckFail.MideaDraftId == "" {
		return ErrMideaDraftIdIsNull
	}
	if req.UserName != SdkUser {
		return ErrUserNameIsNotSDKUser
	}

	_, value, exists := app.state.Get(KeyDraft(billFinancingCheckFail.MideaDraftId))
	if !exists {
		return ErrBillIsNotExists
	}
	bill := &MideaBill{}
	err := UnmarshalMessage(value, bill)
	if err != nil {
		return ErrStorage
	}

	if bill.BillState != BillState_BillFinancingWaitCheck {
		return ErrBillStateNotMatch
	}

	return nil
}

//18-融资审核拒绝
func (app *MideaBillApplication) billFinancingCheckFail(req *Request) (*Response, error) {
	billFinancingCheckFail := req.GetBillFinancingCheckFail()
	if billFinancingCheckFail == nil {
		return nil, ErrGetValueIsNull
	}

	err := app.checkBillFinancingCheckFail(req)
	if err != nil {
		return nil, err
	}

	//通过票号查询票信息
	_, value, exists := app.state.Get(KeyDraft(billFinancingCheckFail.MideaDraftId))
	if !exists {
		return nil, ErrBillIsNotExists
	}
	bill := &MideaBill{}
	err = UnmarshalMessage(value, bill)
	if err != nil {
		return nil, ErrStorage
	}

	//'待收票企业'字段清空
	bill.WaitRecvBillEntName = ""
	bill.WaitRecvBillEntCode = ""
	bill.WaitRecvBillPublicKey = nil
	bill.BillState = BillState_BillNormalOwn
	bill.UpdateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)

	save, err := MarshalMessage(bill)
	if err != nil {
		return nil, err
	}
	app.state.Set(KeyDraft(bill.MideaDraftId), save)

	return &Response{Value: &Response_BillFinancingCheckFail{&ResponseBillFinancingCheckFail{InstructionId: req.InstructionId}}}, nil
}

// 检查--融资冲销
func (app *MideaBillApplication) checkBillFinancingFail(req *Request) error {
	billFinancingFail := req.GetBillFinancingFail()
	if billFinancingFail == nil {
		return ErrGetValueIsNull
	}

	if billFinancingFail.MideaDraftId == "" {
		return ErrMideaDraftIdIsNull
	}

	if req.UserName != SdkUser {
		return ErrUserNameIsNotSDKUser
	}

	_, value, exists := app.state.Get(KeyDraft(billFinancingFail.MideaDraftId))
	if !exists {
		return ErrBillIsNotExists
	}
	bill := &MideaBill{}
	err := UnmarshalMessage(value, bill)
	if err != nil {
		return ErrStorage
	}

    if bill.BillState != BillState_BillNormalOwn {
		return ErrBillStateNotMatch
	}

	// 如果上一持票人状态为空，则已撤销，不能再发起
	if bill.LastBillPublicKey == nil {
		return ErrBillStateNotMatch
	}

	return nil
}

// 融资冲销
func (app *MideaBillApplication) billFinancingFail(req *Request) (*Response, error) {
	billFinancingFail := req.GetBillFinancingFail()
	if billFinancingFail == nil {
		return nil, ErrGetValueIsNull
	}

	err := app.checkBillFinancingFail(req)
	if err != nil {
		return nil, err
	}
	//通过票号查询票信息
	_, value, exists := app.state.Get(KeyDraft(billFinancingFail.MideaDraftId))
	if !exists {
		return nil, ErrBillIsNotExists
	}
	bill := &MideaBill{}
	err = UnmarshalMessage(value, bill)
	if err != nil {
		return nil, ErrStorage
	}

	//'原收票企业'赋值给'收票企业'，'原收票企业'字段清空
	bill.RecvBillEntName = bill.LastBillEntName
	bill.RecvBillEntCode = bill.LastBillEntCode
	bill.RecvBillPublicKey = bill.LastBillPublicKey
	bill.LastBillEntName = ""
	bill.LastBillEntCode = ""
	bill.LastBillPublicKey = nil
	bill.UpdateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)

	save, err := MarshalMessage(bill)
	if err != nil {
		return nil, err
	}
	app.state.Set(KeyDraft(bill.MideaDraftId), save)

	return &Response{Value: &Response_BillFinancingFail{&ResponseBillFinancingFail{InstructionId: req.InstructionId}}}, nil
}

// 检查--兑付
func (app *MideaBillApplication) checkBillPay(req *Request) error {
	billPay := req.GetBillPay()
	if billPay == nil {
		return ErrGetValueIsNull
	}

	if billPay.MideaDraftId == "" {
		return ErrMideaDraftIdIsNull
	}

	if req.UserName != SdkUser {
		return ErrUserNameIsNotSDKUser
	}

	//检查票是否存在
	_, value, exists := app.state.Get(KeyDraft(billPay.MideaDraftId))
	if !exists {
		return ErrBillIsNotExists
	}

	bill := &MideaBill{}
	err := UnmarshalMessage(value, bill)
	if err != nil {
		return ErrStorage
	}

	if bill.BillState != BillState_BillNormalOwn {
		return ErrBillStateNotMatch
	}
	if strings.Compare(bill.ExpireDay, time.Now().Format(formatDate)) < 0 {
		return ErrExpireDayNotMatch
	}

	return nil
}

//20-兑付
func (app *MideaBillApplication) billPay(req *Request) (*Response, error) {
	billPay := req.GetBillPay()
	if billPay == nil {
		return nil, ErrGetValueIsNull
	}

	err := app.checkBillPay(req)
	if err != nil {
		return nil, err
	}

	//通过票号查询票信息
	_, value, exists := app.state.Get(KeyDraft(billPay.MideaDraftId))
	if !exists {
		return nil, ErrBillIsNotExists
	}
	bill := &MideaBill{}
	err = UnmarshalMessage(value, bill)
	if err != nil {
		return nil, ErrStorage
	}

	bill.BillState = BillState_BillPaid
	bill.PayDay = time.Unix(req.OperatorTime, 0).Format(formatDate)
	bill.UpdateTime = time.Unix(req.OperatorTime, 0).Format(formatDateTime)

	save, err := MarshalMessage(bill)
	if err != nil {
		return nil, err
	}
	app.state.Set(KeyDraft(bill.MideaDraftId), save)

	return &Response{Value: &Response_BillPay{&ResponseBillPay{InstructionId: req.InstructionId}}}, nil
}


func (app *MideaBillApplication) saveReceipt(instructionId int64, receipt *Receipt) error {
	save, err := MarshalMessage(receipt)
	if err != nil {
		return err
	}
	app.state.Set(KeyReq(instructionId), save)
	return err
}

func (app *MideaBillApplication) checkInstructionId(instructionId int64) error {
	_, _, exists := app.state.Get(KeyReq(instructionId))
	if !exists {
		return nil
	}
	return ErrDupInstructionId
}
