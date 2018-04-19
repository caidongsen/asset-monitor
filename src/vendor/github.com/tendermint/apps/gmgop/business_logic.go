package gmgop

import (
	"bytes"
	"encoding/hex"
	"log"
)

func (app *GmgopApplication) checkUserCreate(pubkey []byte, uid string) error {
	//println("checkuser:", uid)
	if len(uid) < 6 {
		return ErrUidTooShort
	}
	if !isAdmin(pubkey) {
		return ErrNotAdmin
	}
	_, _, exists := app.state.Get([]byte(KeyUser(uid)))
	if exists {
		return ErrUserExist
	}
	return nil
}

func (app *GmgopApplication) userCreate(uid string, pubkey []byte, userPubkey []byte, info []byte, role int32, instructionId int64) (*Response, error) {
	err := app.checkUserCreate(pubkey, uid)
	if err != nil {
		return nil, err
	}
	user := &User{}
	user.Pubkey = userPubkey[:]
	user.Role = role
	user.Info = info[:]
	save, err := MarshalMessage(user)
	if err != nil {
		return nil, err
	}
	log.Println("debug", uid, hex.EncodeToString(userPubkey), role)
	app.state.Set([]byte(KeyUser(uid)), save)
	event := &EventUserCreate{}
	event.Uid = uid
	event.Role = role
	return &Response{Value: &Response_UserCreate{&ResponseUserCreate{InstructionId: instructionId, Event: &Event{Value: &Event_UserCreate{event}}}}}, nil
}

func (app *GmgopApplication) checkGopAssertCreate(uid string, pubkey []byte, assertId string) error {
	_, _, exists := app.state.Get([]byte(KeyAssert(assertId)))
	if exists {
		return ErrAssertExist
	}
	_, value, exists := app.state.Get([]byte(KeyUser(uid)))
	if !exists {
		return ErrUserNotExist
	}
	var user User
	err := UnmarshalMessage(value, &user)
	if err != nil {
		return ErrStorage
	}
	if user.Role != int32(Role_RlGop) {
		return ErrNotGop
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		log.Println("debug", hex.EncodeToString(user.Pubkey), hex.EncodeToString(pubkey), uid)
		return ErrWrongPubkey
	}
	return nil
}

func (app *GmgopApplication) gopAssertCreate(uid string, pubkey []byte, assertId string, price int64, info []byte, instructionId int64) (*Response, error) {
	err := app.checkGopAssertCreate(uid, pubkey, assertId)
	if err != nil {
		return nil, err
	}
	assert := &Assert{}
	assert.Price = price
	assert.Info = info[:]
	assert.State = AssertState_StLeasehold
	assert.Isdel = false
	assert.CreatorUid = uid
	save, err := MarshalMessage(assert)
	if err != nil {
		return nil, err
	}
	log.Println("debug", assertId, price, uid)
	app.state.Set([]byte(KeyAssert(assertId)), save)
	event := &EventAssertCreate{}
	event.AssertId = assertId
	event.State = AssertState_StLeasehold
	return &Response{Value: &Response_AssertCreate{&ResponseAssertCreate{InstructionId: instructionId, Event: &Event{Value: &Event_AssertCreate{event}}}}}, nil
}

func (app *GmgopApplication) checkGopAssertDelete(uid string, pubkey []byte, assertId string) error {
	_, value, exists := app.state.Get([]byte(KeyAssert(assertId)))
	if !exists {
		return ErrAssertNotExist
	}
	var assert Assert
	err := UnmarshalMessage(value, &assert)
	if err != nil {
		return ErrStorage
	}
	if assert.CreatorUid != uid {
		return ErrNoRight
	}
	_, value, exists = app.state.Get([]byte(KeyUser(uid)))
	if !exists {
		return ErrUserNotExist
	}
	var user User
	err = UnmarshalMessage(value, &user)
	if err != nil {
		return ErrStorage
	}
	if user.Role != int32(Role_RlGop) {
		return ErrNotGop
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return ErrWrongPubkey
	}
	return nil
}

func (app *GmgopApplication) gopAssertDelete(uid string, pubkey []byte, assertId string, instructionId int64) (*Response, error) {
	_, value, exists := app.state.Get([]byte(KeyUser(uid)))
	if !exists {
		return nil, ErrUserNotExist
	}
	var user User
	err := UnmarshalMessage(value, &user)
	if err != nil {
		return nil, ErrStorage
	}
	if user.Role != int32(Role_RlGop) {
		return nil, ErrNotGop
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return nil, ErrWrongPubkey
	}
	_, value, exists = app.state.Get([]byte(KeyAssert(assertId)))
	if !exists {
		return nil, ErrAssertNotExist
	}
	var assert Assert
	err = UnmarshalMessage(value, &assert)
	if err != nil {
		return nil, ErrStorage
	}
	if assert.CreatorUid != uid {
		return nil, ErrNoRight
	}
	assert.Isdel = true
	save, err := MarshalMessage(&assert)
	if err != nil {
		return nil, err
	}
	log.Println("debug", assertId, uid)
	app.state.Set([]byte(KeyAssert(assertId)), save)
	event := &EventAssertDelete{}
	event.AssertId = assertId
	return &Response{Value: &Response_AssertDelete{&ResponseAssertDelete{InstructionId: instructionId, Event: &Event{Value: &Event_AssertDelete{event}}}}}, nil
}

func (app *GmgopApplication) checkBuContractCreate(uid string, pubkey []byte, assertId, contractId string, price int64) error {
	_, value, exists := app.state.Get([]byte(KeyAssert(assertId)))
	if !exists {
		return ErrAssertNotExist
	}
	var assert Assert
	err := UnmarshalMessage(value, &assert)
	if err != nil {
		return ErrStorage
	}
	if assert.Price != price {
		return ErrWrongPrice
	}
	if assert.Isdel {
		return ErrAssertIsDelete
	}
	if assert.State != AssertState_StLeasehold {
		return ErrAssertNotUsable
	}
	_, value, exists = app.state.Get([]byte(KeyUser(uid)))
	if !exists {
		log.Println("debug", uid)
		return ErrUserNotExist
	}
	var user User
	err = UnmarshalMessage(value, &user)
	if err != nil {
		return ErrStorage
	}
	if user.Role != int32(Role_RlBu) {
		return ErrNotBu
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return ErrWrongPubkey
	}
	_, _, exists = app.state.Get([]byte(KeyContract(contractId)))
	if exists {
		return ErrContractExist
	}
	return nil
}

func (app *GmgopApplication) buContractCreate(uid string, pubkey []byte, assertId, contractId string, price, start, end int64, info []byte, instructionId int64) (*Response, error) {
	err := app.checkBuContractCreate(uid, pubkey, assertId, contractId, price)
	if err != nil {
		return nil, err
	}
	_, value, exists := app.state.Get([]byte(KeyAssert(assertId)))
	if !exists {
		return nil, ErrAssertNotExist
	}
	var assert Assert
	err = UnmarshalMessage(value, &assert)
	if err != nil {
		return nil, ErrStorage
	}
	assert.Applicants = append(assert.Applicants, contractId)
	contract := &Contract{}
	contract.AssertId = assertId
	contract.Price = price
	contract.StartT = start
	contract.EndT = end
	contract.Info = info
	contract.State = ContractState_CSCreated
	contract.CreatorUid = uid
	save, err := MarshalMessage(contract)
	if err != nil {
		return nil, err
	}
	log.Println("debug", uid, assertId, contractId, price)
	app.state.Set([]byte(KeyContract(contractId)), save)
	save, err = MarshalMessage(&assert)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyAssert(assertId)), save)
	event := &EventContractCreate{}
	event.ContractId = contractId
	event.State = ContractState_CSCreated
	return &Response{Value: &Response_ContractCreate{&ResponseContractCreate{InstructionId: instructionId, Event: &Event{Value: &Event_ContractCreate{event}}}}}, nil
}

func (app *GmgopApplication) checkGopContractAgree(uid string, pubkey []byte, contractId string) error {
	_, value, exists := app.state.Get([]byte(KeyUser(uid)))
	if !exists {
		return ErrUserNotExist
	}
	var user User
	err := UnmarshalMessage(value, &user)
	if err != nil {
		return ErrStorage
	}
	if user.Role != int32(Role_RlGop) {
		return ErrNotGop
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return ErrWrongPubkey
	}
	_, value, exists = app.state.Get([]byte(KeyContract(contractId)))
	if !exists {
		return ErrContractNotExist
	}
	var contract Contract
	err = UnmarshalMessage(value, &contract)
	if err != nil {
		return ErrStorage
	}
	_, value, exists = app.state.Get([]byte(KeyAssert(contract.AssertId)))
	if !exists {
		return ErrAssertNotExist
	}
	var assert Assert
	err = UnmarshalMessage(value, &assert)
	if err != nil {
		return ErrStorage
	}
	if assert.Price != contract.Price {
		return ErrWrongPrice
	}
	if assert.Isdel {
		return ErrAssertIsDelete
	}
	if assert.State != AssertState_StLeasehold {
		return ErrAssertNotUsable
	}
	if assert.CreatorUid != uid {
		return ErrNoRight
	}
	return nil
}

func (app *GmgopApplication) gopContractAgree(uid string, pubkey []byte, contractId string, instructionId int64) (*Response, error) {
	_, value, exists := app.state.Get([]byte(KeyUser(uid)))
	if !exists {
		return nil, ErrUserNotExist
	}
	var user User
	err := UnmarshalMessage(value, &user)
	if err != nil {
		return nil, ErrStorage
	}
	if user.Role != int32(Role_RlGop) {
		return nil, ErrNotGop
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return nil, ErrWrongPubkey
	}
	_, value, exists = app.state.Get([]byte(KeyContract(contractId)))
	if !exists {
		return nil, ErrContractNotExist
	}
	var contract Contract
	err = UnmarshalMessage(value, &contract)
	if err != nil {
		return nil, ErrStorage
	}
	_, value, exists = app.state.Get([]byte(KeyAssert(contract.AssertId)))
	if !exists {
		return nil, ErrAssertNotExist
	}
	var assert Assert
	err = UnmarshalMessage(value, &assert)
	if err != nil {
		return nil, ErrStorage
	}
	if assert.Price != contract.Price {
		return nil, ErrWrongPrice
	}
	if assert.Isdel {
		return nil, ErrAssertIsDelete
	}
	if assert.State != AssertState_StLeasehold {
		return nil, ErrAssertNotUsable
	}
	if assert.CreatorUid != uid {
		return nil, ErrNoRight
	}
	contract.State = ContractState_CSAgreeed
	assert.State = AssertState_StLeased
	event := &EventContractAgree{}
	for _, v := range assert.Applicants {
		if v != contractId {
			_, value, exists = app.state.Get([]byte(KeyContract(v)))
			if !exists {
				log.Println("debug contract not found", v)
				continue
			}
			var tmpContract Contract
			err = UnmarshalMessage(value, &tmpContract)
			if err != nil {
				log.Println("debug contract bad", v)
				continue
			}
			tmpContract.State = ContractState_CSRejected
			save, err := MarshalMessage(&tmpContract)
			if err != nil {
				log.Println("debug contract not saved", v)
				continue
			}
			app.state.Set([]byte(KeyContract(v)), save)
			event.Rejected = append(event.Rejected, v)
		}
	}
	assert.Applicants = nil
	save, err := MarshalMessage(&contract)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyContract(contractId)), save)
	save, err = MarshalMessage(&assert)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyAssert(contract.AssertId)), save)
	event.ContractId = contractId
	event.Astate = assert.State
	event.Cstate = contract.State
	event.AssertId = contract.AssertId
	return &Response{Value: &Response_ContractAgree{&ResponseContractAgree{InstructionId: instructionId, Event: &Event{Value: &Event_ContractAgree{event}}}}}, nil
}

func (app *GmgopApplication) checkGopContractReject(uid string, pubkey []byte, contractId string) error {
	_, value, exists := app.state.Get([]byte(KeyUser(uid)))
	if !exists {
		return ErrUserNotExist
	}
	var user User
	err := UnmarshalMessage(value, &user)
	if err != nil {
		return ErrStorage
	}
	if user.Role != int32(Role_RlGop) {
		return ErrNotGop
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return ErrWrongPubkey
	}
	_, value, exists = app.state.Get([]byte(KeyContract(contractId)))
	if !exists {
		return ErrContractNotExist
	}
	var contract Contract
	err = UnmarshalMessage(value, &contract)
	if err != nil {
		return ErrStorage
	}
	_, value, exists = app.state.Get([]byte(KeyAssert(contract.AssertId)))
	if !exists {
		return ErrAssertNotExist
	}
	var assert Assert
	err = UnmarshalMessage(value, &assert)
	if err != nil {
		return ErrStorage
	}
	if assert.CreatorUid != uid {
		return ErrNoRight
	}
	return nil
}

func (app *GmgopApplication) gopContractReject(uid string, pubkey []byte, contractId string, instructionId int64) (*Response, error) {
	_, value, exists := app.state.Get([]byte(KeyUser(uid)))
	if !exists {
		return nil, ErrUserNotExist
	}
	var user User
	err := UnmarshalMessage(value, &user)
	if err != nil {
		return nil, ErrStorage
	}
	if user.Role != int32(Role_RlGop) {
		return nil, ErrNotGop
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return nil, ErrWrongPubkey
	}
	_, value, exists = app.state.Get([]byte(KeyContract(contractId)))
	if !exists {
		return nil, ErrContractNotExist
	}
	var contract Contract
	err = UnmarshalMessage(value, &contract)
	if err != nil {
		return nil, ErrStorage
	}
	_, value, exists = app.state.Get([]byte(KeyAssert(contract.AssertId)))
	if !exists {
		return nil, ErrAssertNotExist
	}
	var assert Assert
	err = UnmarshalMessage(value, &assert)
	if err != nil {
		return nil, ErrStorage
	}
	if assert.CreatorUid != uid {
		return nil, ErrNoRight
	}
	contract.State = ContractState_CSRejected
	for k, v := range assert.Applicants {
		if v == contractId {
			tmpApplicants := assert.Applicants[:]
			assert.Applicants = tmpApplicants[:k]
			assert.Applicants = append(assert.Applicants, tmpApplicants[k+1:]...)
			break
		}
	}
	save, err := MarshalMessage(&contract)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyContract(contractId)), save)
	save, err = MarshalMessage(&assert)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyAssert(contract.AssertId)), save)
	event := &EventContractReject{}
	event.ContractId = contractId
	event.State = contract.State
	return &Response{Value: &Response_ContractReject{&ResponseContractReject{InstructionId: instructionId, Event: &Event{Value: &Event_ContractReject{event}}}}}, nil
}

func (app *GmgopApplication) checkBuRecallContract(uid string, pubkey []byte, contractId string) error {
	_, value, exists := app.state.Get([]byte(KeyUser(uid)))
	if !exists {
		return ErrUserNotExist
	}
	var user User
	err := UnmarshalMessage(value, &user)
	if err != nil {
		return ErrStorage
	}
	if user.Role != int32(Role_RlBu) {
		return ErrNotBu
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return ErrWrongPubkey
	}
	_, value, exists = app.state.Get([]byte(KeyContract(contractId)))
	if !exists {
		return ErrContractNotExist
	}
	var contract Contract
	err = UnmarshalMessage(value, &contract)
	if err != nil {
		return ErrStorage
	}
	if contract.CreatorUid != uid {
		return ErrNoRight
	}
	if contract.State != ContractState_CSCreated {
		return ErrCannotRecall
	}
	_, value, exists = app.state.Get([]byte(KeyAssert(contract.AssertId)))
	if !exists {
		return ErrAssertNotExist
	}
	var assert Assert
	err = UnmarshalMessage(value, &assert)
	if err != nil {
		return ErrStorage
	}
	bHas := false
	for _, v := range assert.Applicants {
		if v == contractId {
			bHas = true
			break
		}
	}
	if !bHas {
		return ErrNotApplicant
	}
	return nil
}

func (app *GmgopApplication) buRecallContract(uid string, pubkey []byte, contractId string, instructionId int64) (*Response, error) {
	err := app.checkBuRecallContract(uid, pubkey, contractId)
	if err != nil {
		return nil, err
	}
	_, value, exists := app.state.Get([]byte(KeyContract(contractId)))
	if !exists {
		return nil, ErrContractNotExist
	}
	var contract Contract
	err = UnmarshalMessage(value, &contract)
	if err != nil {
		return nil, ErrStorage
	}
	contract.State = ContractState_CSRecalled
	_, value, exists = app.state.Get([]byte(KeyAssert(contract.AssertId)))
	if !exists {
		return nil, ErrAssertNotExist
	}
	var assert Assert
	err = UnmarshalMessage(value, &assert)
	if err != nil {
		return nil, ErrStorage
	}
	for k, v := range assert.Applicants {
		if v == contractId {
			tmpApplicants := assert.Applicants[:]
			assert.Applicants = tmpApplicants[:k]
			assert.Applicants = append(assert.Applicants, tmpApplicants[k+1:]...)
			break
		}
	}
	save, err := MarshalMessage(&contract)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyContract(contractId)), save)
	save, err = MarshalMessage(&assert)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyAssert(contract.AssertId)), save)
	event := &EventRecallContract{}
	event.ContractId = contractId
	event.Uid = uid
	return &Response{Value: &Response_BuRecallContract{&ResponseRecallContract{InstructionId: instructionId, Event: &Event{Value: &Event_RecallContract{event}}}}}, nil
}

func (app *GmgopApplication) checkGopDeliverAssert(uid string, pubkey []byte, contractId string) error {
	_, value, exists := app.state.Get([]byte(KeyUser(uid)))
	if !exists {
		return ErrUserNotExist
	}
	var user User
	err := UnmarshalMessage(value, &user)
	if err != nil {
		return ErrStorage
	}
	if user.Role != int32(Role_RlGop) {
		return ErrNotGop
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return ErrWrongPubkey
	}
	_, value, exists = app.state.Get([]byte(KeyContract(contractId)))
	if !exists {
		return ErrContractNotExist
	}
	var contract Contract
	err = UnmarshalMessage(value, &contract)
	if err != nil {
		return ErrStorage
	}
	_, value, exists = app.state.Get([]byte(KeyAssert(contract.AssertId)))
	if !exists {
		return ErrAssertNotExist
	}
	var assert Assert
	err = UnmarshalMessage(value, &assert)
	if err != nil {
		return ErrStorage
	}
	if assert.CreatorUid != uid {
		return ErrNoRight
	}
	return nil
}

func (app *GmgopApplication) gopDeliverAssert(uid string, pubkey []byte, contractId string, instructionId int64) (*Response, error) {
	_, value, exists := app.state.Get([]byte(KeyUser(uid)))
	if !exists {
		return nil, ErrUserNotExist
	}
	var user User
	err := UnmarshalMessage(value, &user)
	if err != nil {
		return nil, ErrStorage
	}
	if user.Role != int32(Role_RlGop) {
		return nil, ErrNotGop
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return nil, ErrWrongPubkey
	}
	_, value, exists = app.state.Get([]byte(KeyContract(contractId)))
	if !exists {
		return nil, ErrContractNotExist
	}
	var contract Contract
	err = UnmarshalMessage(value, &contract)
	if err != nil {
		return nil, ErrStorage
	}
	_, value, exists = app.state.Get([]byte(KeyAssert(contract.AssertId)))
	if !exists {
		return nil, ErrAssertNotExist
	}
	var assert Assert
	err = UnmarshalMessage(value, &assert)
	if err != nil {
		return nil, ErrStorage
	}
	if assert.CreatorUid != uid {
		return nil, ErrNoRight
	}
	assert.State = AssertState_StDelivered
	save, err := MarshalMessage(&assert)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyAssert(contract.AssertId)), save)
	event := &EventDeliverAssert{}
	event.AssertId = contract.AssertId
	event.State = assert.State
	return &Response{Value: &Response_DeliverAssert{&ResponseDeliverAssert{InstructionId: instructionId, Event: &Event{Value: &Event_DeliverAssert{event}}}}}, nil
}

func (app *GmgopApplication) checkBuDeliverAssertConfirm(uid string, pubkey []byte, contractId string) error {
	_, value, exists := app.state.Get([]byte(KeyUser(uid)))
	if !exists {
		return ErrUserNotExist
	}
	var user User
	err := UnmarshalMessage(value, &user)
	if err != nil {
		return ErrStorage
	}
	if user.Role != int32(Role_RlBu) {
		return ErrNotBu
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return ErrWrongPubkey
	}
	_, value, exists = app.state.Get([]byte(KeyContract(contractId)))
	if !exists {
		return ErrContractNotExist
	}
	var contract Contract
	err = UnmarshalMessage(value, &contract)
	if err != nil {
		return ErrStorage
	}
	if contract.CreatorUid != uid {
		return ErrNoRight
	}
	_, value, exists = app.state.Get([]byte(KeyAssert(contract.AssertId)))
	if !exists {
		return ErrAssertNotExist
	}
	var assert Assert
	err = UnmarshalMessage(value, &assert)
	if err != nil {
		return ErrStorage
	}
	return nil
}

func (app *GmgopApplication) buDeliverAssertConfirm(uid string, pubkey []byte, contractId string, instructionId int64) (*Response, error) {
	_, value, exists := app.state.Get([]byte(KeyUser(uid)))
	if !exists {
		return nil, ErrUserNotExist
	}
	var user User
	err := UnmarshalMessage(value, &user)
	if err != nil {
		return nil, ErrStorage
	}
	if user.Role != int32(Role_RlBu) {
		return nil, ErrNotBu
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return nil, ErrWrongPubkey
	}
	_, value, exists = app.state.Get([]byte(KeyContract(contractId)))
	if !exists {
		return nil, ErrContractNotExist
	}
	var contract Contract
	err = UnmarshalMessage(value, &contract)
	if err != nil {
		return nil, ErrStorage
	}
	if contract.CreatorUid != uid {
		return nil, ErrNoRight
	}
	_, value, exists = app.state.Get([]byte(KeyAssert(contract.AssertId)))
	if !exists {
		return nil, ErrAssertNotExist
	}
	var assert Assert
	err = UnmarshalMessage(value, &assert)
	if err != nil {
		return nil, ErrStorage
	}
	assert.State = AssertState_StConfirmed
	save, err := MarshalMessage(&assert)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyAssert(contract.AssertId)), save)
	event := &EventDeliverAssertConfirm{}
	event.AssertId = contract.AssertId
	event.State = assert.State
	return &Response{Value: &Response_DeliverAssertConfirm{&ResponseDeliverAssertConfirm{InstructionId: instructionId, Event: &Event{Value: &Event_DeliverAssertConfirm{event}}}}}, nil
}

func (app *GmgopApplication) checkGopFin(uid string, pubkey []byte, contractId string) error {
	_, value, exists := app.state.Get([]byte(KeyUser(uid)))
	if !exists {
		return ErrUserNotExist
	}
	var user User
	err := UnmarshalMessage(value, &user)
	if err != nil {
		return ErrStorage
	}
	if user.Role != int32(Role_RlGop) {
		return ErrNotGop
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return ErrWrongPubkey
	}
	_, value, exists = app.state.Get([]byte(KeyContract(contractId)))
	if !exists {
		return ErrContractNotExist
	}
	var contract Contract
	err = UnmarshalMessage(value, &contract)
	if err != nil {
		return ErrStorage
	}
	_, value, exists = app.state.Get([]byte(KeyAssert(contract.AssertId)))
	if !exists {
		return ErrAssertNotExist
	}
	var assert Assert
	err = UnmarshalMessage(value, &assert)
	if err != nil {
		return ErrStorage
	}
	if assert.CreatorUid != uid {
		return ErrNoRight
	}
	return nil
}

func (app *GmgopApplication) gopFin(uid string, pubkey []byte, contractId string, instructionId int64) (*Response, error) {
	_, value, exists := app.state.Get([]byte(KeyUser(uid)))
	if !exists {
		return nil, ErrUserNotExist
	}
	var user User
	err := UnmarshalMessage(value, &user)
	if err != nil {
		return nil, ErrStorage
	}
	if user.Role != int32(Role_RlGop) {
		return nil, ErrNotGop
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return nil, ErrWrongPubkey
	}
	_, value, exists = app.state.Get([]byte(KeyContract(contractId)))
	if !exists {
		return nil, ErrContractNotExist
	}
	var contract Contract
	err = UnmarshalMessage(value, &contract)
	if err != nil {
		return nil, ErrStorage
	}
	_, value, exists = app.state.Get([]byte(KeyAssert(contract.AssertId)))
	if !exists {
		return nil, ErrAssertNotExist
	}
	var assert Assert
	err = UnmarshalMessage(value, &assert)
	if err != nil {
		return nil, ErrStorage
	}
	if assert.CreatorUid != uid {
		return nil, ErrNoRight
	}
	assert.State = AssertState_StExpired
	contract.State = ContractState_CSExpired
	save, err := MarshalMessage(&assert)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyAssert(contract.AssertId)), save)
	save, err = MarshalMessage(&contract)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyContract(contractId)), save)
	event := &EventFin{}
	event.AssertId = contract.AssertId
	event.Cstate = contract.State
	event.Astate = assert.State
	event.ContractId = contractId
	return &Response{Value: &Response_Fin{&ResponseFin{InstructionId: instructionId, Event: &Event{Value: &Event_Fin{event}}}}}, nil
}

func (app *GmgopApplication) checkChangePubkey(uid string, oldPubkey, newPubkey []byte) error {
	_, value, exists := app.state.Get([]byte(KeyUser(uid)))
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

func (app *GmgopApplication) changePubkey(uid string, oldPubkey, newPubkey []byte, instructionId int64) (*Response, error) {
	_, value, exists := app.state.Get([]byte(KeyUser(uid)))
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
	user.Pubkey = newPubkey[:]
	save, err := MarshalMessage(&user)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyUser(uid)), save)
	event := &EventChangePubkey{}
	event.NewPubkey = newPubkey[:]
	event.Uid = uid
	return &Response{Value: &Response_ChangePubkey{&ResponseChangePubkey{InstructionId: instructionId, Event: &Event{Value: &Event_ChangePubkey{event}}}}}, nil
}
