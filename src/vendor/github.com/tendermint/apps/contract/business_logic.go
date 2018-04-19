package contract

import (
	"bytes"
)

func (app *ContractApplication) checkUserCreate(pubkey []byte, uid string) error {
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

func (app *ContractApplication) userCreate(uid string, pubkey []byte, userPubkey []byte, info []byte, instructionId int64) (*Response, error) {
	err := app.checkUserCreate(pubkey, uid)
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
	app.state.Set([]byte(KeyUser(uid)), save)
	event := &EventUserCreate{}
	event.Uid = uid
	return &Response{Value: &Response_UserCreate{&ResponseUserCreate{InstructionId: instructionId, Event: &Event{Value: &Event_UserCreate{event}}}}}, nil
}

func (app *ContractApplication) checkChangePubkey(uid string, oldPubkey, newPubkey []byte) error {
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

func (app *ContractApplication) changePubkey(uid string, oldPubkey, newPubkey []byte, instructionId int64) (*Response, error) {
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

func (app *ContractApplication) checkCreateContract(uid string, pubkey []byte, contractId string) error {
	_, value, exists := app.state.Get([]byte(KeyUser(uid)))
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
	_, _, exists = app.state.Get([]byte(KeyContract(contractId)))
	if exists {
		return ErrContractExist
	}
	return nil
}

func (app *ContractApplication) createContract(uid string, pubkey []byte, instructionId int64, contractId string, info []byte) (*Response, error) {
	err := app.checkCreateContract(uid, pubkey, contractId)
	if err != nil {
		return nil, err
	}
	contract := &Contract{}
	contract.Info = info[:]
	contract.CreatorUid = uid
	contract.State = ContractState_CSCreated
	save, err := MarshalMessage(contract)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyContract(contractId)), save)
	event := &EventCreateContract{}
	event.ContractId = contractId
	event.State = contract.State
	return &Response{Value: &Response_CreateContract{&ResponseCreateContract{InstructionId: instructionId, Event: &Event{Value: &Event_CreateContract{event}}}}}, nil
}

func (app *ContractApplication) checkLaunchSign(uid string, pubkey []byte, contractId string, signers []string, endT, opT int64) error {
	if opT >= endT {
		return ErrContractExpired
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
	if !bytes.Equal(user.Pubkey, pubkey) {
		return ErrNoRight
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
	if contract.State != ContractState_CSCreated && contract.State != ContractState_CSEdited && contract.State != ContractState_CSRepaired {
		return ErrCannotLaunch
	}
	if contract.CreatorUid != uid {
		return ErrNotCreator
	}
	for _, v := range signers {
		_, _, exists := app.state.Get([]byte(KeyUser(v)))
		if !exists {
			return ErrSignerNotExist
		}
	}
	return nil
}

func (app *ContractApplication) launchSign(uid string, pubkey []byte, instructionId int64, contractId string, endT, opT int64, signers []string) (*Response, error) {
	err := app.checkLaunchSign(uid, pubkey, contractId, signers[:], endT, opT)
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
	contract.EndT = endT
	contract.Signers = nil
	contract.Rejectors = nil
	for _, v := range signers {
		signer := &Signer{}
		signer.Uid = v
		signer.Signed = false
		contract.Signers = append(contract.Signers, signer)
	}
	contract.State = ContractState_CSLaunched
	save, err := MarshalMessage(&contract)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyContract(contractId)), save)
	event := &EventLaunchSign{}
	event.State = contract.State
	return &Response{Value: &Response_LaunchSign{&ResponseLaunchSign{InstructionId: instructionId, Event: &Event{Value: &Event_LaunchSign{event}}}}}, nil
}

func (app *ContractApplication) checkCreatorSign(uid string, pubkey []byte, contractId string, opT int64) error {
	_, value, exists := app.state.Get([]byte(KeyUser(uid)))
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
	_, value, exists = app.state.Get([]byte(KeyContract(contractId)))
	if !exists {
		return ErrContractNotExist
	}
	var contract Contract
	err = UnmarshalMessage(value, &contract)
	if err != nil {
		return ErrStorage
	}
	if contract.CreatorSignT != 0 {
		return ErrCreatorSigned
	}
	if contract.CreatorUid != uid {
		return ErrNotCreator
	}
	if contract.State != ContractState_CSLaunched {
		return ErrWrongState
	}
	if opT >= contract.EndT {
		return ErrContractExpired
	}
	return nil
}

func (app *ContractApplication) creatorSign(uid string, pubkey []byte, instructionId int64, contractId string, opT int64) (*Response, error) {
	err := app.checkCreatorSign(uid, pubkey, contractId, opT)
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
	contract.CreatorSignT = opT
	save, err := MarshalMessage(&contract)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyContract(contractId)), save)
	return &Response{Value: &Response_CreatorSign{&ResponseCreatorSign{InstructionId: instructionId}}}, nil
}

func (app *ContractApplication) checkEditContract(uid string, pubkey []byte, contractId string, opT int64) error {
	_, value, exists := app.state.Get([]byte(KeyUser(uid)))
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
	_, value, exists = app.state.Get([]byte(KeyContract(contractId)))
	if !exists {
		return ErrContractNotExist
	}
	var contract Contract
	err = UnmarshalMessage(value, &contract)
	if err != nil {
		return ErrStorage
	}
	if contract.CreatorSignT != 0 {
		return ErrCreatorSigned
	}
	if contract.CreatorUid != uid {
		return ErrNotCreator
	}
	if opT >= contract.EndT {
		return ErrContractExpired
	}
	return nil
}

func (app *ContractApplication) editContract(uid string, pubkey []byte, instructionId int64, contractId string, opT int64, info []byte) (*Response, error) {
	err := app.checkEditContract(uid, pubkey, contractId, opT)
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
	tmpInfo := make([]byte, len(contract.Info))
	copy(tmpInfo[:], contract.Info[:])
	contract.HistoryInfos = append(contract.HistoryInfos, tmpInfo)
	contract.Info = info[:]
	contract.State = ContractState_CSEdited
	save, err := MarshalMessage(&contract)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyContract(contractId)), save)
	return &Response{Value: &Response_EditContract{&ResponseEditContract{InstructionId: instructionId}}}, nil
}

func (app *ContractApplication) checkRejectContract(uid string, pubkey []byte, contractId string, opT int64) error {
	_, value, exists := app.state.Get([]byte(KeyUser(uid)))
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
	_, value, exists = app.state.Get([]byte(KeyContract(contractId)))
	if !exists {
		return ErrContractNotExist
	}
	var contract Contract
	err = UnmarshalMessage(value, &contract)
	if err != nil {
		return ErrStorage
	}
	bHas := false
	for _, v := range contract.Signers {
		if v.Uid == uid {
			bHas = true
			break
		}
	}
	if !bHas {
		return ErrNotSigner
	}
	if contract.State != ContractState_CSLaunched {
		return ErrWrongState
	}
	if opT >= contract.EndT {
		return ErrContractExpired
	}
	return nil
}

func (app *ContractApplication) rejectContract(uid string, pubkey []byte, instructionId int64, contractId string, opT int64) (*Response, error) {
	err := app.checkRejectContract(uid, pubkey, contractId, opT)
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
	contract.Rejectors = append(contract.Rejectors, uid)
	contract.State = ContractState_CSRejected
	save, err := MarshalMessage(&contract)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyContract(contractId)), save)
	event := &EventRejectContract{}
	event.State = contract.State
	return &Response{Value: &Response_RejectContract{&ResponseRejectContract{InstructionId: instructionId, Event: &Event{Value: &Event_RejectContract{event}}}}}, nil
}

func (app *ContractApplication) checkAcceptContract(uid string, pubkey []byte, contractId string, opT int64) error {
	_, value, exists := app.state.Get([]byte(KeyUser(uid)))
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
	_, value, exists = app.state.Get([]byte(KeyContract(contractId)))
	if !exists {
		return ErrContractNotExist
	}
	var contract Contract
	err = UnmarshalMessage(value, &contract)
	if err != nil {
		return ErrStorage
	}
	bHas := false
	for _, v := range contract.Signers {
		if v.Uid == uid {
			bHas = true
			break
		}
	}
	if !bHas {
		return ErrNotSigner
	}
	if contract.State != ContractState_CSLaunched {
		return ErrWrongState
	}
	if opT >= contract.EndT {
		return ErrContractExpired
	}
	return nil
}

func (app *ContractApplication) acceptContract(uid string, pubkey []byte, instructionId int64, contractId string, opT int64) (*Response, error) {
	err := app.checkAcceptContract(uid, pubkey, contractId, opT)
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
	for _, v := range contract.Signers {
		if v.Uid == uid {
			v.Signed = true
			break
		}
	}
	allSigned := true
	for _, v := range contract.Signers {
		if !v.Signed {
			allSigned = false
			break
		}
	}
	if allSigned {
		contract.State = ContractState_CSEffective
	}
	save, err := MarshalMessage(&contract)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyContract(contractId)), save)
	event := &EventAcceptContract{}
	event.State = contract.State
	return &Response{Value: &Response_AcceptContract{&ResponseAcceptContract{InstructionId: instructionId, Event: &Event{Value: &Event_AcceptContract{event}}}}}, nil
}

func (app *ContractApplication) checkRepairContract(uid string, pubkey []byte, contractId string, endT, opT int64) error {
	_, value, exists := app.state.Get([]byte(KeyUser(uid)))
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
	_, value, exists = app.state.Get([]byte(KeyContract(contractId)))
	if !exists {
		return ErrContractNotExist
	}
	var contract Contract
	err = UnmarshalMessage(value, &contract)
	if err != nil {
		return ErrStorage
	}
	if contract.State != ContractState_CSRejected {
		return ErrWrongState
	}
	if contract.CreatorUid != uid {
		return ErrNotCreator
	}
	if opT >= endT {
		return ErrContractExpired
	}
	return nil
}

func (app *ContractApplication) repairContract(uid string, pubkey []byte, instructionId int64, contractId string, endT, opT int64, info []byte) (*Response, error) {
	err := app.checkRepairContract(uid, pubkey, contractId, endT, opT)
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
	tmpInfo := make([]byte, len(contract.Info))
	copy(tmpInfo[:], contract.Info[:])
	contract.HistoryInfos = append(contract.HistoryInfos, tmpInfo)
	contract.Info = info[:]
	contract.EndT = endT
	contract.State = ContractState_CSRepaired
	save, err := MarshalMessage(&contract)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyContract(contractId)), save)
	event := &EventRepairContract{}
	event.State = contract.State
	return &Response{Value: &Response_RepairContract{&ResponseRepairContract{InstructionId: instructionId, Event: &Event{Value: &Event_RepairContract{event}}}}}, nil
}

func (app *ContractApplication) checkDisuseContract(uid string, pubkey []byte, contractId string, opT int64) error {
	_, value, exists := app.state.Get([]byte(KeyUser(uid)))
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
	_, value, exists = app.state.Get([]byte(KeyContract(contractId)))
	if !exists {
		return ErrContractNotExist
	}
	var contract Contract
	err = UnmarshalMessage(value, &contract)
	if err != nil {
		return ErrStorage
	}
	if contract.State == ContractState_CSEffective {
		return ErrWrongState
	}
	if contract.CreatorUid != uid {
		return ErrNotCreator
	}
	if opT >= contract.EndT {
		return ErrContractExpired
	}
	return nil
}

func (app *ContractApplication) disuseContract(uid string, pubkey []byte, instructionId int64, contractId string, opT int64) (*Response, error) {
	err := app.checkDisuseContract(uid, pubkey, contractId, opT)
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
	contract.State = ContractState_CSDisuse
	save, err := MarshalMessage(&contract)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyContract(contractId)), save)
	event := &EventDisuseContract{}
	event.State = contract.State
	return &Response{Value: &Response_DisuseContract{&ResponseDisuseContract{InstructionId: instructionId, Event: &Event{Value: &Event_DisuseContract{event}}}}}, nil
}
