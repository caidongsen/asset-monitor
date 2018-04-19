package supplychain

import (
	"bytes"
)

func (app *SupplychainApplication) checkUserCreate(pubkey []byte, userUid string) error {
	if len(userUid) < 6 {
		return ErrUidTooShort
	}
	if app.getAdminType(pubkey) != AdminType_A_NORMAL {
		return ErrNotAdmin
	}
	_, _, exists := app.state.Get([]byte(KeyUser(userUid)))
	if exists {
		return ErrUserExist
	}
	return nil
}

func (app *SupplychainApplication) userCreate(pubkey []byte, userUid string, userPubkey []byte, userType UserType, userCredit int64, userInfo []byte, instructionId int64) (*Response, error) {
	err := app.checkUserCreate(pubkey, userUid)
	if err != nil {
		return nil, err
	}
	user := &User{}
	user.Pubkey = userPubkey[:]
	user.Type = userType
	user.Credit = userCredit
	user.Info = userInfo[:]
	user.State = AccountState_ACC_CREATED
	save, err := MarshalMessage(user)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyUser(userUid)), save)
	event := &EventUserCreate{}
	event.Uid = userUid
	return &Response{Value: &Response_UserCreate{&ResponseUserCreate{InstructionId: instructionId, Event: &Event{Value: &Event_UserCreate{event}}}}}, nil
}

func (app *SupplychainApplication) checkUserPaid(pubkey []byte, uid string, userPubkey []byte) error {
	if app.getAdminType(pubkey) != AdminType_A_NORMAL {
		return ErrNotAdmin
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
	if !bytes.Equal(user.Pubkey, userPubkey) {
		return ErrWrongPubkey
	}
	return nil
}

func (app *SupplychainApplication) userPaid(pubkey []byte, uid string, userPubkey []byte, instructionId int64) (*Response, error) {
	if app.getAdminType(pubkey) != AdminType_A_NORMAL {
		return nil, ErrNotAdmin
	}
	_, value, exists := app.state.Get([]byte(KeyUser(uid)))
	if !exists {
		return nil, ErrUserNotExist
	}
	var user User
	err := UnmarshalMessage(value, &user)
	if err != nil {
		return nil, ErrStorage
	}
	if !bytes.Equal(user.Pubkey, userPubkey) {
		return nil, ErrWrongPubkey
	}
	user.State = AccountState_ACC_PAID
	save, err := MarshalMessage(&user)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyUser(uid)), save)
	event := &EventUserPaid{}
	event.State = user.State
	return &Response{Value: &Response_UserPaid{&ResponseUserPaid{InstructionId: instructionId, Event: &Event{Value: &Event_UserPaid{event}}}}}, nil
}

func (app *SupplychainApplication) checkChangePubkey(pubkey []byte, uid string, newPubkey []byte) error {
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
	return nil
}

func (app *SupplychainApplication) changePubkey(pubkey []byte, uid string, newPubkey []byte, instructionId int64) (*Response, error) {
	_, value, exists := app.state.Get([]byte(KeyUser(uid)))
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
	copy(user.Pubkey[:], newPubkey[:])
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

func (app *SupplychainApplication) checkChangeCredit(pubkey []byte, uid string, credit int64) error {
	if app.getAdminType(pubkey) != AdminType_A_NORMAL {
		return ErrNotAdmin
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
	return nil
}

func (app *SupplychainApplication) changeCredit(pubkey []byte, uid string, credit int64, instructionId int64) (*Response, error) {
	if app.getAdminType(pubkey) != AdminType_A_NORMAL {
		return nil, ErrNotAdmin
	}
	_, value, exists := app.state.Get([]byte(KeyUser(uid)))
	if !exists {
		return nil, ErrUserNotExist
	}
	var user User
	err := UnmarshalMessage(value, &user)
	if err != nil {
		return nil, ErrStorage
	}
	user.Credit = credit
	save, err := MarshalMessage(&user)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyUser(uid)), save)
	event := &EventChangeCredit{}
	event.Credit = credit
	return &Response{Value: &Response_ChangeCredit{&ResponseChangeCredit{InstructionId: instructionId, Event: &Event{Value: &Event_ChangeCredit{event}}}}}, nil
}

func (app *SupplychainApplication) checkCreateLoan(pubkey []byte, userUid string, loanId int64, amount int64) error {
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
	if user.Credit < amount {
		return ErrCreditNotEnough
	}
	_, _, exists = app.state.Get([]byte(KeyLoan(loanId)))
	if exists {
		return ErrLoanExist
	}
	return nil
}

func (app *SupplychainApplication) createLoan(pubkey []byte, userUid string, loanId int64, amount int64, rate int32, expiration int64, info []byte, instructionId int64) (*Response, error) {
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
	_, _, exists = app.state.Get([]byte(KeyLoan(loanId)))
	if exists {
		return nil, ErrLoanExist
	}
	if user.Credit < amount {
		return nil, ErrCreditNotEnough
	}
	loan := &Loan{}
	loan.LoanId = userUid
	loan.Info = info[:]
	loan.State = LoanState_L_CREATED
	loan.Amount = amount
	loan.Rate = rate
	loan.Expiration = expiration
	save, err := MarshalMessage(loan)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyLoan(loanId)), save)
	user.LoanIds = append(user.LoanIds, loanId)
	save, err = MarshalMessage(&user)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyUser(userUid)), save)
	event := &EventCreateLoan{}
	event.LoanId = loanId
	event.State = loan.State
	return &Response{Value: &Response_CreateLoan{&ResponseCreateLoan{InstructionId: instructionId, Event: &Event{Value: &Event_CreateLoan{event}}}}}, nil
}

func (app *SupplychainApplication) checkApplyCredit(pubkey []byte, userUid string, loanId int64) error {
	_, value, exists := app.state.Get([]byte(KeyLoan(loanId)))
	if !exists {
		return ErrLoanNotExist
	}
	var loan Loan
	err := UnmarshalMessage(value, &loan)
	if err != nil {
		return ErrStorage
	}
	if userUid != loan.LoanId {
		return ErrNotCreator
	}
	_, value, exists = app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return ErrUserNotExist
	}
	var user User
	err = UnmarshalMessage(value, &user)
	if err != nil {
		return ErrStorage
	}
	if !bytes.Equal(pubkey, user.Pubkey) {
		return ErrNoRight
	}
	return nil
}

func (app *SupplychainApplication) applyCredit(pubkey []byte, userUid string, loanId int64, instructionId int64) (*Response, error) {
	_, value, exists := app.state.Get([]byte(KeyLoan(loanId)))
	if !exists {
		return nil, ErrLoanNotExist
	}
	var loan Loan
	err := UnmarshalMessage(value, &loan)
	if err != nil {
		return nil, ErrStorage
	}
	if userUid != loan.LoanId {
		return nil, ErrNotCreator
	}
	_, value, exists = app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return nil, ErrUserNotExist
	}
	var user User
	err = UnmarshalMessage(value, &user)
	if err != nil {
		return nil, ErrStorage
	}
	if !bytes.Equal(pubkey, user.Pubkey) {
		return nil, ErrNoRight
	}
	loan.Type = LoanType_LT_CREDIT
	loan.State = LoanState_L_APPLY_CREDIT
	save, err := MarshalMessage(&loan)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyLoan(loanId)), save)
	event := &EventApplyCredit{}
	event.State = loan.State
	return &Response{Value: &Response_ApplyCredit{&ResponseApplyCredit{InstructionId: instructionId, Event: &Event{Value: &Event_ApplyCredit{event}}}}}, nil
}

func (app *SupplychainApplication) checkApplyGuarantee(pubkey []byte, userUid, guaranteeId string, loanId int64) error {
	_, value, exists := app.state.Get([]byte(KeyLoan(loanId)))
	if !exists {
		return ErrLoanNotExist
	}
	var loan Loan
	err := UnmarshalMessage(value, &loan)
	if err != nil {
		return ErrStorage
	}
	if userUid != loan.LoanId {
		return ErrNotCreator
	}
	_, value, exists = app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return ErrUserNotExist
	}
	var user User
	err = UnmarshalMessage(value, &user)
	if err != nil {
		return ErrStorage
	}
	if !bytes.Equal(pubkey, user.Pubkey) {
		return ErrNoRight
	}
	_, value, exists = app.state.Get([]byte(KeyUser(guaranteeId)))
	if !exists {
		return ErrGuaranteeNotExist
	}
	var guarantee User
	err = UnmarshalMessage(value, &guarantee)
	if err != nil {
		return ErrStorage
	}
	if guarantee.Guarantee < loan.Amount {
		return ErrGuaranteeNotEnough
	}
	return nil
}

func (app *SupplychainApplication) applyGuarantee(pubkey []byte, userUid string, guaranteeId string, loanId int64, instructionId int64) (*Response, error) {
	_, value, exists := app.state.Get([]byte(KeyLoan(loanId)))
	if !exists {
		return nil, ErrLoanNotExist
	}
	var loan Loan
	err := UnmarshalMessage(value, &loan)
	if err != nil {
		return nil, ErrStorage
	}
	if userUid != loan.LoanId {
		return nil, ErrNotCreator
	}
	_, value, exists = app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return nil, ErrUserNotExist
	}
	var user User
	err = UnmarshalMessage(value, &user)
	if err != nil {
		return nil, ErrStorage
	}
	if !bytes.Equal(pubkey, user.Pubkey) {
		return nil, ErrNoRight
	}
	_, value, exists = app.state.Get([]byte(KeyUser(guaranteeId)))
	if !exists {
		return nil, ErrGuaranteeNotExist
	}
	var guarantee User
	err = UnmarshalMessage(value, &guarantee)
	if err != nil {
		return nil, ErrStorage
	}
	if guarantee.Guarantee < loan.Amount {
		return nil, ErrGuaranteeNotEnough
	}
	loan.Type = LoanType_LT_GUARANTEE
	loan.State = LoanState_L_APPLY_GUARANTEE
	loan.GuaranteeId = guaranteeId
	save, err := MarshalMessage(&loan)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyLoan(loanId)), save)
	event := &EventApplyGuarantee{}
	event.State = loan.State
	event.GuaranteeId = guaranteeId
	return &Response{Value: &Response_ApplyGuarantee{&ResponseApplyGuarantee{InstructionId: instructionId, Event: &Event{Value: &Event_ApplyGuarantee{event}}}}}, nil
}

func (app *SupplychainApplication) checkGuaranteeFeedback(pubkey []byte, userUid string, loanId int64) error {
	_, value, exists := app.state.Get([]byte(KeyLoan(loanId)))
	if !exists {
		return ErrLoanNotExist
	}
	var loan Loan
	err := UnmarshalMessage(value, &loan)
	if err != nil {
		return ErrStorage
	}
	if userUid != loan.GuaranteeId {
		return ErrNotGuarantee
	}
	_, value, exists = app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return ErrUserNotExist
	}
	var user User
	err = UnmarshalMessage(value, &user)
	if err != nil {
		return ErrStorage
	}
	if !bytes.Equal(pubkey, user.Pubkey) {
		return ErrNoRight
	}
	if user.Guarantee < loan.Amount {
		return ErrGuaranteeNotEnough
	}
	return nil
}

func (app *SupplychainApplication) guaranteeFeedback(pubkey []byte, userUid string, loanId int64, instructionId int64) (*Response, error) {
	_, value, exists := app.state.Get([]byte(KeyLoan(loanId)))
	if !exists {
		return nil, ErrLoanNotExist
	}
	var loan Loan
	err := UnmarshalMessage(value, &loan)
	if err != nil {
		return nil, ErrStorage
	}
	if userUid != loan.GuaranteeId {
		return nil, ErrNotGuarantee
	}
	_, value, exists = app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return nil, ErrUserNotExist
	}
	var user User
	err = UnmarshalMessage(value, &user)
	if err != nil {
		return nil, ErrStorage
	}
	if !bytes.Equal(pubkey, user.Pubkey) {
		return nil, ErrNoRight
	}
	loan.State = LoanState_L_GUARANTEE_AGREE
	save, err := MarshalMessage(&loan)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyLoan(loanId)), save)
	if user.Guarantee < loan.Amount {
		return nil, ErrGuaranteeNotEnough
	}
	user.Guarantee -= loan.Amount
	user.GuaranteeIds = append(user.GuaranteeIds, loanId)
	save, err = MarshalMessage(&user)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyUser(userUid)), save)
	event := &EventGuaranteeFeedback{}
	event.State = loan.State
	event.Guarantee = loan.Amount
	return &Response{Value: &Response_GuaranteeFeedback{&ResponseGuaranteeFeedback{InstructionId: instructionId, Event: &Event{Value: &Event_GuaranteeFeedback{event}}}}}, nil
}

func (app *SupplychainApplication) checkCreditFeedback(pubkey []byte, loanId int64) error {
	_, value, exists := app.state.Get([]byte(KeyLoan(loanId)))
	if !exists {
		return ErrLoanNotExist
	}
	var loan Loan
	err := UnmarshalMessage(value, &loan)
	if err != nil {
		return ErrStorage
	}
	if app.getAdminType(pubkey) != AdminType_A_NORMAL {
		return ErrNotAdmin
	}
	return nil
}

func (app *SupplychainApplication) creditFeedback(pubkey []byte, loanId int64, instructionId int64) (*Response, error) {
	_, value, exists := app.state.Get([]byte(KeyLoan(loanId)))
	if !exists {
		return nil, ErrLoanNotExist
	}
	var loan Loan
	err := UnmarshalMessage(value, &loan)
	if err != nil {
		return nil, ErrStorage
	}
	if app.getAdminType(pubkey) != AdminType_A_NORMAL {
		return nil, ErrNotAdmin
	}
	loan.State = LoanState_L_ADMIN_AGREE
	save, err := MarshalMessage(&loan)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyLoan(loanId)), save)
	event := &EventCreditFeedback{}
	event.State = loan.State
	return &Response{Value: &Response_CreditFeedback{&ResponseCreditFeedback{InstructionId: instructionId, Event: &Event{Value: &Event_CreditFeedback{event}}}}}, nil
}

func (app *SupplychainApplication) checkEditCredit(pubkey []byte, userUid string, loanId int64) error {
	_, value, exists := app.state.Get([]byte(KeyLoan(loanId)))
	if !exists {
		return ErrLoanNotExist
	}
	var loan Loan
	err := UnmarshalMessage(value, &loan)
	if err != nil {
		return ErrStorage
	}
	if loan.LoanId != userUid {
		return ErrNotCreator
	}
	_, value, exists = app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return ErrUserNotExist
	}
	var user User
	err = UnmarshalMessage(value, &user)
	if err != nil {
		return ErrStorage
	}
	if !bytes.Equal(pubkey, user.Pubkey) {
		return ErrNoRight
	}
	return nil
}

func (app *SupplychainApplication) editCredit(pubkey []byte, userUid string, loanId int64, info []byte, instructionId int64) (*Response, error) {
	_, value, exists := app.state.Get([]byte(KeyLoan(loanId)))
	if !exists {
		return nil, ErrLoanNotExist
	}
	var loan Loan
	err := UnmarshalMessage(value, &loan)
	if err != nil {
		return nil, ErrStorage
	}
	if loan.LoanId != userUid {
		return nil, ErrNotCreator
	}
	_, value, exists = app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return nil, ErrUserNotExist
	}
	var user User
	err = UnmarshalMessage(value, &user)
	if err != nil {
		return nil, ErrStorage
	}
	if !bytes.Equal(pubkey, user.Pubkey) {
		return nil, ErrNoRight
	}
	copy(loan.Info[:], info[:])
	loan.State = LoanState_L_EDIT
	save, err := MarshalMessage(&loan)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyLoan(loanId)), save)
	event := &EventEditCredit{}
	event.State = loan.State
	return &Response{Value: &Response_EditCredit{&ResponseEditCredit{InstructionId: instructionId, Event: &Event{Value: &Event_EditCredit{event}}}}}, nil
}

func (app *SupplychainApplication) checkIssueLoan(pubkey []byte, userUid string, loanId int64) error {
	_, value, exists := app.state.Get([]byte(KeyLoan(loanId)))
	if !exists {
		return ErrLoanNotExist
	}
	var loan Loan
	err := UnmarshalMessage(value, &loan)
	if err != nil {
		return ErrStorage
	}
	if loan.LoanId != userUid {
		return ErrNotCreator
	}
	_, value, exists = app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return ErrUserNotExist
	}
	var user User
	err = UnmarshalMessage(value, &user)
	if err != nil {
		return ErrStorage
	}
	if !bytes.Equal(pubkey, user.Pubkey) {
		return ErrNoRight
	}
	if user.Credit < loan.Amount {
		return ErrCreditNotEnough
	}

	return nil
}

func (app *SupplychainApplication) issueLoan(pubkey []byte, userUid string, loanId int64, instructionId int64) (*Response, error) {
	_, value, exists := app.state.Get([]byte(KeyLoan(loanId)))
	if !exists {
		return nil, ErrLoanNotExist
	}
	var loan Loan
	err := UnmarshalMessage(value, &loan)
	if err != nil {
		return nil, ErrStorage
	}
	if loan.LoanId != userUid {
		return nil, ErrNotCreator
	}
	_, value, exists = app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return nil, ErrUserNotExist
	}
	var user User
	err = UnmarshalMessage(value, &user)
	if err != nil {
		return nil, ErrStorage
	}
	if !bytes.Equal(pubkey, user.Pubkey) {
		return nil, ErrNoRight
	}
	loan.State = LoanState_L_ISSUE
	save, err := MarshalMessage(&loan)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyLoan(loanId)), save)
	if user.Credit < loan.Amount {
		return nil, ErrCreditNotEnough
	}
	user.Credit -= loan.Amount
	save, err = MarshalMessage(&user)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyUser(userUid)), save)
	event := &EventIssueLoan{}
	event.State = loan.State
	event.Credit = loan.Amount
	return &Response{Value: &Response_IssueLoan{&ResponseIssueLoan{InstructionId: instructionId, Event: &Event{Value: &Event_IssueLoan{event}}}}}, nil
}

func (app *SupplychainApplication) checkCancelLoan(pubkey []byte, userUid string, loanId int64) error {
	_, value, exists := app.state.Get([]byte(KeyLoan(loanId)))
	if !exists {
		return ErrLoanNotExist
	}
	var loan Loan
	err := UnmarshalMessage(value, &loan)
	if err != nil {
		return ErrStorage
	}
	if loan.LoanId != userUid {
		return ErrNotCreator
	}
	_, value, exists = app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return ErrUserNotExist
	}
	var user User
	err = UnmarshalMessage(value, &user)
	if err != nil {
		return ErrStorage
	}
	if !bytes.Equal(pubkey, user.Pubkey) {
		return ErrNoRight
	}
	if loan.State != LoanState_L_ISSUE {
		return ErrWrongState
	}
	if loan.Type == LoanType_LT_GUARANTEE {
		_, value, exists = app.state.Get([]byte(KeyUser(loan.GuaranteeId)))
		if !exists {
			return ErrGuaranteeNotExist
		}
		var guarantee User
		err = UnmarshalMessage(value, &guarantee)
		if err != nil {
			return ErrStorage
		}
	}

	return nil
}

func (app *SupplychainApplication) cancelLoan(pubkey []byte, userUid string, loanId int64, instructionId int64) (*Response, error) {
	_, value, exists := app.state.Get([]byte(KeyLoan(loanId)))
	if !exists {
		return nil, ErrLoanNotExist
	}
	var loan Loan
	err := UnmarshalMessage(value, &loan)
	if err != nil {
		return nil, ErrStorage
	}
	if loan.LoanId != userUid {
		return nil, ErrNotCreator
	}
	_, value, exists = app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return nil, ErrUserNotExist
	}
	var user User
	err = UnmarshalMessage(value, &user)
	if err != nil {
		return nil, ErrStorage
	}
	if !bytes.Equal(pubkey, user.Pubkey) {
		return nil, ErrNoRight
	}
	if loan.State != LoanState_L_ISSUE {
		return nil, ErrWrongState
	}
	loan.State = LoanState_L_CANCELED
	save, err := MarshalMessage(&loan)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyLoan(loanId)), save)
	user.Credit += loan.Amount
	for k, v := range user.LoanIds {
		if v == loanId {
			tmpLoanIds := user.LoanIds[:]
			user.LoanIds = tmpLoanIds[:k]
			user.LoanIds = append(user.LoanIds, tmpLoanIds[k+1:]...)
			break
		}
	}
	save, err = MarshalMessage(&user)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyUser(userUid)), save)
	if loan.Type == LoanType_LT_GUARANTEE {
		_, value, exists = app.state.Get([]byte(KeyUser(loan.GuaranteeId)))
		if !exists {
			return nil, ErrGuaranteeNotExist
		}
		var guarantee User
		err = UnmarshalMessage(value, &guarantee)
		if err != nil {
			return nil, ErrStorage
		}
		guarantee.Guarantee += loan.Amount
		for k, v := range guarantee.GuaranteeIds {
			if v == loanId {
				tmpGuaranteeIds := guarantee.GuaranteeIds[:]
				guarantee.GuaranteeIds = tmpGuaranteeIds[:k]
				guarantee.GuaranteeIds = append(guarantee.GuaranteeIds, tmpGuaranteeIds[k+1:]...)
				break
			}
		}
		save, err = MarshalMessage(&guarantee)
		if err != nil {
			return nil, err
		}
		app.state.Set([]byte(KeyUser(loan.GuaranteeId)), save)
	}
	event := &EventCancelLoan{}
	event.State = loan.State
	return &Response{Value: &Response_CancelLoan{&ResponseCancelLoan{InstructionId: instructionId, Event: &Event{Value: &Event_CancelLoan{event}}}}}, nil
}

func (app *SupplychainApplication) checkPrepareBuy(pubkey []byte, userUid string, loanId int64) error {
	_, value, exists := app.state.Get([]byte(KeyLoan(loanId)))
	if !exists {
		return ErrLoanNotExist
	}
	var loan Loan
	err := UnmarshalMessage(value, &loan)
	if err != nil {
		return ErrStorage
	}
	if loan.LoanId == userUid || loan.GuaranteeId == userUid {
		return ErrCannotBuy
	}
	_, value, exists = app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return ErrUserNotExist
	}
	var user User
	err = UnmarshalMessage(value, &user)
	if err != nil {
		return ErrStorage
	}
	if user.Rmb < loan.Amount {
		return ErrRmbNotEnough
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return ErrNoRight
	}
	if loan.State != LoanState_L_ISSUE {
		return ErrWrongState
	}
	return nil
}

func (app *SupplychainApplication) prepareBuy(pubkey []byte, userUid string, loanId int64, instructionId int64) (*Response, error) {
	_, value, exists := app.state.Get([]byte(KeyLoan(loanId)))
	if !exists {
		return nil, ErrLoanNotExist
	}
	var loan Loan
	err := UnmarshalMessage(value, &loan)
	if err != nil {
		return nil, ErrStorage
	}
	if loan.LoanId == userUid || loan.GuaranteeId == userUid {
		return nil, ErrCannotBuy
	}
	_, value, exists = app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return nil, ErrUserNotExist
	}
	var user User
	err = UnmarshalMessage(value, &user)
	if err != nil {
		return nil, ErrStorage
	}
	if user.Rmb < loan.Amount {
		return nil, ErrRmbNotEnough
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return nil, ErrNoRight
	}
	if loan.State != LoanState_L_ISSUE {
		return nil, ErrWrongState
	}
	loan.BuyId = userUid
	save, err := MarshalMessage(&loan)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyLoan(loanId)), save)
	user.BuyIds = append(user.BuyIds, loanId)
	save, err = MarshalMessage(&user)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyUser(userUid)), save)
	return &Response{Value: &Response_PrepareBuy{&ResponsePrepareBuy{InstructionId: instructionId}}}, nil
}

func (app *SupplychainApplication) checkPay(pubkey []byte, userUid string, loanId int64) error {
	_, value, exists := app.state.Get([]byte(KeyLoan(loanId)))
	if !exists {
		return ErrLoanNotExist
	}
	var loan Loan
	err := UnmarshalMessage(value, &loan)
	if err != nil {
		return ErrStorage
	}
	if loan.BuyId != userUid {
		return ErrNotBuyer
	}
	if loan.State != LoanState_L_ISSUE {
		return ErrWrongState
	}
	_, value, exists = app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return ErrUserNotExist
	}
	var user User
	err = UnmarshalMessage(value, &user)
	if err != nil {
		return ErrStorage
	}
	if user.Rmb < loan.Amount {
		return ErrRmbNotEnough
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return ErrNoRight
	}
	return nil
}

func (app *SupplychainApplication) pay(pubkey []byte, userUid string, loanId int64, instructionId int64) (*Response, error) {
	_, value, exists := app.state.Get([]byte(KeyLoan(loanId)))
	if !exists {
		return nil, ErrLoanNotExist
	}
	var loan Loan
	err := UnmarshalMessage(value, &loan)
	if err != nil {
		return nil, ErrStorage
	}
	if loan.BuyId != userUid {
		return nil, ErrNotBuyer
	}
	if loan.State != LoanState_L_ISSUE {
		return nil, ErrWrongState
	}
	_, value, exists = app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return nil, ErrUserNotExist
	}
	var user User
	err = UnmarshalMessage(value, &user)
	if err != nil {
		return nil, ErrStorage
	}
	if user.Rmb < loan.Amount {
		return nil, ErrRmbNotEnough
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return nil, ErrNoRight
	}
	loan.State = LoanState_L_PAYED
	save, err := MarshalMessage(&loan)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyLoan(loanId)), save)
	user.Rmb -= loan.Amount
	save, err = MarshalMessage(&user)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyUser(userUid)), save)
	event := &EventPay{}
	event.State = loan.State
	event.Rmb = loan.Amount
	return &Response{Value: &Response_Pay{&ResponsePay{InstructionId: instructionId, Event: &Event{Value: &Event_Pay{event}}}}}, nil
}

func (app *SupplychainApplication) checkConfirmReceive(pubkey []byte, userUid string, loanId int64) error {
	_, value, exists := app.state.Get([]byte(KeyLoan(loanId)))
	if !exists {
		return ErrLoanNotExist
	}
	var loan Loan
	err := UnmarshalMessage(value, &loan)
	if err != nil {
		return ErrStorage
	}
	if loan.LoanId != userUid {
		return ErrNotCreator
	}
	_, value, exists = app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return ErrUserNotExist
	}
	var user User
	err = UnmarshalMessage(value, &user)
	if err != nil {
		return ErrStorage
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return ErrNoRight
	}
	return nil
}

func (app *SupplychainApplication) confirmReceive(pubkey []byte, userUid string, loanId int64, instructionId int64) (*Response, error) {
	_, value, exists := app.state.Get([]byte(KeyLoan(loanId)))
	if !exists {
		return nil, ErrLoanNotExist
	}
	var loan Loan
	err := UnmarshalMessage(value, &loan)
	if err != nil {
		return nil, ErrStorage
	}
	if loan.LoanId != userUid {
		return nil, ErrNotCreator
	}
	_, value, exists = app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return nil, ErrUserNotExist
	}
	var user User
	err = UnmarshalMessage(value, &user)
	if err != nil {
		return nil, ErrStorage
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return nil, ErrNoRight
	}
	loan.State = LoanState_L_CONFIRM_RECEIVE
	save, err := MarshalMessage(&loan)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyLoan(loanId)), save)
	user.Rmb += loan.Amount
	save, err = MarshalMessage(&user)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyUser(userUid)), save)
	event := &EventConfirmReceive{}
	event.State = loan.State
	event.Rmb = loan.Amount
	return &Response{Value: &Response_ConfirmReceive{&ResponseConfirmReceive{InstructionId: instructionId, Event: &Event{Value: &Event_ConfirmReceive{event}}}}}, nil
}

func (app *SupplychainApplication) checkRepayAdvance(pubkey []byte, userUid string, loanId, amount int64) error {
	_, value, exists := app.state.Get([]byte(KeyLoan(loanId)))
	if !exists {
		return ErrLoanNotExist
	}
	var loan Loan
	err := UnmarshalMessage(value, &loan)
	if err != nil {
		return ErrStorage
	}
	if loan.LoanId != userUid {
		return ErrNotCreator
	}
	_, value, exists = app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return ErrUserNotExist
	}
	var user User
	err = UnmarshalMessage(value, &user)
	if err != nil {
		return ErrStorage
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return ErrNoRight
	}
	if user.Rmb < amount {
		return ErrRmbNotEnough
	}
	return nil
}

func (app *SupplychainApplication) repayAdvance(pubkey []byte, userUid string, loanId, amount int64, instructionId int64) (*Response, error) {
	_, value, exists := app.state.Get([]byte(KeyLoan(loanId)))
	if !exists {
		return nil, ErrLoanNotExist
	}
	var loan Loan
	err := UnmarshalMessage(value, &loan)
	if err != nil {
		return nil, ErrStorage
	}
	if loan.LoanId != userUid {
		return nil, ErrNotCreator
	}
	_, value, exists = app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return nil, ErrUserNotExist
	}
	var user User
	err = UnmarshalMessage(value, &user)
	if err != nil {
		return nil, ErrStorage
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return nil, ErrNoRight
	}
	if user.Rmb < amount {
		return nil, ErrRmbNotEnough
	}
	loan.State = LoanState_L_REPAY_ADVANCE
	loan.RepayAmount = amount
	save, err := MarshalMessage(&loan)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyLoan(loanId)), save)
	user.Rmb -= amount
	save, err = MarshalMessage(&user)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyUser(userUid)), save)
	event := &EventRepayAdvance{}
	event.State = loan.State
	event.Rmb = amount
	return &Response{Value: &Response_RepayAdvance{&ResponseRepayAdvance{InstructionId: instructionId, Event: &Event{Value: &Event_RepayAdvance{event}}}}}, nil
}

func (app *SupplychainApplication) checkRepay(pubkey []byte, userUid string, loanId, amount int64) error {
	_, value, exists := app.state.Get([]byte(KeyLoan(loanId)))
	if !exists {
		return ErrLoanNotExist
	}
	var loan Loan
	err := UnmarshalMessage(value, &loan)
	if err != nil {
		return ErrStorage
	}
	if loan.LoanId != userUid {
		return ErrNotCreator
	}
	_, value, exists = app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return ErrUserNotExist
	}
	var user User
	err = UnmarshalMessage(value, &user)
	if err != nil {
		return ErrStorage
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return ErrNoRight
	}
	if user.Rmb < amount {
		return ErrRmbNotEnough
	}
	return nil
}

func (app *SupplychainApplication) repay(pubkey []byte, userUid string, loanId, amount int64, instructionId int64) (*Response, error) {
	_, value, exists := app.state.Get([]byte(KeyLoan(loanId)))
	if !exists {
		return nil, ErrLoanNotExist
	}
	var loan Loan
	err := UnmarshalMessage(value, &loan)
	if err != nil {
		return nil, ErrStorage
	}
	if loan.LoanId != userUid {
		return nil, ErrNotCreator
	}
	_, value, exists = app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return nil, ErrUserNotExist
	}
	var user User
	err = UnmarshalMessage(value, &user)
	if err != nil {
		return nil, ErrStorage
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return nil, ErrNoRight
	}
	if user.Rmb < amount {
		return nil, ErrRmbNotEnough
	}
	loan.State = LoanState_L_REPAY
	loan.RepayAmount = amount
	save, err := MarshalMessage(&loan)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyLoan(loanId)), save)
	user.Rmb -= amount
	save, err = MarshalMessage(&user)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyUser(userUid)), save)
	event := &EventRepay{}
	event.State = loan.State
	event.Rmb = amount
	return &Response{Value: &Response_Repay{&ResponseRepay{InstructionId: instructionId, Event: &Event{Value: &Event_Repay{event}}}}}, nil
}

func (app *SupplychainApplication) checkConfirmRepay(pubkey []byte, userUid string, loanId int64) error {
	_, value, exists := app.state.Get([]byte(KeyLoan(loanId)))
	if !exists {
		return ErrLoanNotExist
	}
	var loan Loan
	err := UnmarshalMessage(value, &loan)
	if err != nil {
		return ErrStorage
	}
	if loan.BuyId != userUid {
		return ErrNotBuyer
	}
	_, value, exists = app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return ErrUserNotExist
	}
	var user User
	err = UnmarshalMessage(value, &user)
	if err != nil {
		return ErrStorage
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return ErrNoRight
	}
	_, value, exists = app.state.Get([]byte(KeyUser(loan.LoanId)))
	if !exists {
		return ErrSellerNotExist
	}
	var seller User
	err = UnmarshalMessage(value, &seller)
	if err != nil {
		return ErrStorage
	}
	_, value, exists = app.state.Get([]byte(KeyUser(loan.GuaranteeId)))
	if !exists {
		return ErrGuaranteeNotExist
	}
	var guarantee User
	err = UnmarshalMessage(value, &guarantee)
	if err != nil {
		return ErrStorage
	}
	return nil
}

func (app *SupplychainApplication) confirmRepay(pubkey []byte, userUid string, loanId int64, instructionId int64) (*Response, error) {
	_, value, exists := app.state.Get([]byte(KeyLoan(loanId)))
	if !exists {
		return nil, ErrLoanNotExist
	}
	var loan Loan
	err := UnmarshalMessage(value, &loan)
	if err != nil {
		return nil, ErrStorage
	}
	if loan.BuyId != userUid {
		return nil, ErrNotBuyer
	}
	_, value, exists = app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return nil, ErrUserNotExist
	}
	var user User
	err = UnmarshalMessage(value, &user)
	if err != nil {
		return nil, ErrStorage
	}
	if !bytes.Equal(user.Pubkey, pubkey) {
		return nil, ErrNoRight
	}
	_, value, exists = app.state.Get([]byte(KeyUser(loan.LoanId)))
	if !exists {
		return nil, ErrSellerNotExist
	}
	var seller User
	err = UnmarshalMessage(value, &seller)
	if err != nil {
		return nil, ErrStorage
	}
	loan.State = LoanState_L_CONFIRM_REPAY
	save, err := MarshalMessage(&loan)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyLoan(loanId)), save)
	user.Rmb += loan.RepayAmount
	for k, v := range user.BuyIds {
		if v == loanId {
			tmpBuyIds := user.BuyIds[:]
			user.BuyIds = tmpBuyIds[:k]
			user.BuyIds = append(user.BuyIds, tmpBuyIds[k+1:]...)
			break
		}
	}
	save, err = MarshalMessage(&user)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyUser(userUid)), save)
	seller.Credit += loan.Amount
	for k, v := range seller.LoanIds {
		if v == loanId {
			tmpLoanIds := seller.LoanIds[:]
			seller.LoanIds = tmpLoanIds[:k]
			seller.LoanIds = append(seller.LoanIds, tmpLoanIds[k+1:]...)
			break
		}
	}
	save, err = MarshalMessage(&seller)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyUser(loan.LoanId)), save)
	if loan.Type == LoanType_LT_GUARANTEE {
		_, value, exists = app.state.Get([]byte(KeyUser(loan.GuaranteeId)))
		if !exists {
			return nil, ErrUserNotExist
		}
		var guarantee User
		err = UnmarshalMessage(value, &guarantee)
		if err != nil {
			return nil, ErrStorage
		}
		guarantee.Guarantee += loan.Amount
		for k, v := range guarantee.GuaranteeIds {
			if v == loanId {
				tmpGuaranteeIds := guarantee.GuaranteeIds[:]
				guarantee.GuaranteeIds = tmpGuaranteeIds[:k]
				guarantee.GuaranteeIds = append(guarantee.GuaranteeIds, tmpGuaranteeIds[k+1:]...)
				break
			}
		}
		save, err = MarshalMessage(&guarantee)
		if err != nil {
			return nil, err
		}
		app.state.Set([]byte(KeyUser(loan.GuaranteeId)), save)
	}
	event := &EventConfirmRepay{}
	event.State = loan.State
	event.Rmb = loan.RepayAmount
	return &Response{Value: &Response_ConfirmRepay{&ResponseConfirmRepay{InstructionId: instructionId, Event: &Event{Value: &Event_ConfirmRepay{event}}}}}, nil
}

func (app *SupplychainApplication) checkCreateAdmin(pubkey []byte, adminPubkey []byte) error {
	if app.getAdminType(pubkey) != AdminType_A_NORMAL {
		return ErrNotAdmin
	}
	if app.getAdminType(adminPubkey) == AdminType_A_NORMAL {
		return ErrAdminExist
	}

	return nil
}

func (app *SupplychainApplication) createAdmin(pubkey []byte, adminPubkey []byte, adminType AdminType, instructionId int64) (*Response, error) {
	err := app.checkCreateAdmin(pubkey, adminPubkey)
	if err != nil {
		return nil, err
	}
	admin := &Admin{}
	admin.AdminAddr = adminPubkey[:]
	admin.AdminType = adminType
	admins := app.loadAdmins()
	if admins == nil {
		admins = &Admins{}
	}
	admins.Admins = append(admins.Admins, admin)
	save, err := MarshalMessage(admins)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyAdmins()), save)
	event := &EventCreateAdmin{}
	event.AdminPubkey = adminPubkey[:]
	return &Response{Value: &Response_CreateAdmin{&ResponseCreateAdmin{InstructionId: instructionId, Event: &Event{Value: &Event_CreateAdmin{event}}}}}, nil
}

func (app *SupplychainApplication) checkSetGuarantee(pubkey []byte, uid string, guarantee int64) error {
	if app.getAdminType(pubkey) != AdminType_A_NORMAL {
		return ErrNotAdmin
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
	if user.Type != UserType_U_CORE {
		return ErrWrongUserType
	}
	return nil
}

func (app *SupplychainApplication) setGuarantee(pubkey []byte, uid string, guarantee int64, instructionId int64) (*Response, error) {
	if app.getAdminType(pubkey) != AdminType_A_NORMAL {
		return nil, ErrNotAdmin
	}
	_, value, exists := app.state.Get([]byte(KeyUser(uid)))
	if !exists {
		return nil, ErrUserNotExist
	}
	var user User
	err := UnmarshalMessage(value, &user)
	if err != nil {
		return nil, ErrStorage
	}
	if user.Type != UserType_U_CORE {
		return nil, ErrWrongUserType
	}
	user.Guarantee = guarantee
	save, err := MarshalMessage(&user)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyUser(uid)), save)
	event := &EventSetGuarantee{}
	event.Guarantee = guarantee
	return &Response{Value: &Response_SetGuarantee{&ResponseSetGuarantee{InstructionId: instructionId, Event: &Event{Value: &Event_SetGuarantee{event}}}}}, nil
}

func (app *SupplychainApplication) checkDeposit(pubkey []byte, userUid string, userPubkey []byte, cash int64) error {
	if !isBank(pubkey) {
		return ErrNotBank
	}
	if cash <= 0 {
		return ErrWrongCash
	}
	_, value, exists := app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return ErrUserNotExist
	}
	var user User
	err := UnmarshalMessage(value, &user)
	if err != nil {
		return ErrStorage
	}
	if !bytes.Equal(user.Pubkey, userPubkey) {
		return ErrWrongPubkey
	}
	_, value, exists = app.state.Get([]byte(KeyBank(pubkey)))
	if !exists {
		return ErrBankNotExist
	}
	var bank Bank
	err = UnmarshalMessage(value, &bank)
	if err != nil {
		return ErrStorage
	}
	if bank.Rmb < cash {
		return ErrBankRmbNotEnough
	}
	return nil
}

func (app *SupplychainApplication) deposit(pubkey []byte, userUid string, userPubkey []byte, cash, instructionId int64) (*Response, error) {
	if !isBank(pubkey) {
		return nil, ErrNotBank
	}
	if cash <= 0 {
		return nil, ErrWrongCash
	}
	_, value, exists := app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return nil, ErrUserNotExist
	}
	var user User
	err := UnmarshalMessage(value, &user)
	if err != nil {
		return nil, ErrStorage
	}
	if !bytes.Equal(user.Pubkey, userPubkey) {
		return nil, ErrWrongPubkey
	}
	_, value, exists = app.state.Get([]byte(KeyBank(pubkey)))
	if !exists {
		return nil, ErrBankNotExist
	}
	var bank Bank
	err = UnmarshalMessage(value, &bank)
	if err != nil {
		return nil, ErrStorage
	}
	if bank.Rmb < cash {
		return nil, ErrBankRmbNotEnough
	}
	user.Rmb += cash
	bank.Rmb -= cash
	save, err := MarshalMessage(&user)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyUser(userUid)), save)
	save, err = MarshalMessage(&bank)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyBank(pubkey)), save)
	event := &EventDeposit{}
	event.Balance = user.Rmb
	return &Response{Value: &Response_Deposit{&ResponseDeposit{InstructionId: instructionId, Event: &Event{Value: &Event_Deposit{event}}}}}, nil
}

func (app *SupplychainApplication) checkWithdraw(pubkey []byte, userUid string, bankAddr []byte, cash int64) error {
	if !isBank(bankAddr) {
		return ErrNotBank
	}
	if cash <= 0 {
		return ErrWrongCash
	}
	_, value, exists := app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return ErrUserNotExist
	}
	var user User
	err := UnmarshalMessage(value, &user)
	if err != nil {
		return ErrStorage
	}
	if !bytes.Equal(pubkey, user.Pubkey) {
		return ErrNoRight
	}
	if user.Rmb < cash {
		return ErrUserCashNotEnough
	}

	return nil
}

func (app *SupplychainApplication) withdraw(pubkey []byte, userUid string, bankAddr []byte, cash, instructionId int64) (*Response, error) {
	if !isBank(bankAddr) {
		return nil, ErrNotBank
	}
	if cash <= 0 {
		return nil, ErrWrongCash
	}
	_, value, exists := app.state.Get([]byte(KeyUser(userUid)))
	if !exists {
		return nil, ErrUserNotExist
	}
	var user User
	err := UnmarshalMessage(value, &user)
	if err != nil {
		return nil, ErrStorage
	}
	if !bytes.Equal(pubkey, user.Pubkey) {
		return nil, ErrNoRight
	}
	if user.Rmb < cash {
		return nil, ErrUserCashNotEnough
	}
	_, value, exists = app.state.Get([]byte(KeyBank(bankAddr)))
	if !exists {
		return nil, ErrBankNotExist
	}
	var bank Bank
	err = UnmarshalMessage(value, &bank)
	if err != nil {
		return nil, ErrStorage
	}
	bank.Rmb += cash
	user.Rmb -= cash
	save, err := MarshalMessage(&user)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyUser(userUid)), save)
	save, err = MarshalMessage(&bank)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyBank(bankAddr)), save)
	event := &EventWithdraw{}
	event.Balance = user.Rmb
	return &Response{Value: &Response_Withdraw{&ResponseWithdraw{InstructionId: instructionId, Event: &Event{Value: &Event_Withdraw{event}}}}}, nil
}

func (app *SupplychainApplication) checkIncreaseBankRmb(pubkey []byte, bankAddr []byte, cash int64) error {
	if app.getAdminType(pubkey) != AdminType_A_NORMAL {
		return ErrNotAdmin
	}
	if !isBank(bankAddr) {
		return ErrNotBank
	}
	if cash <= 0 {
		return ErrWrongCash
	}

	return nil
}

func (app *SupplychainApplication) increaseBankRmb(pubkey []byte, bankAddr []byte, cash, instructionId int64) (*Response, error) {
	if app.getAdminType(pubkey) != AdminType_A_NORMAL {
		return nil, ErrNotAdmin
	}
	if !isBank(bankAddr) {
		return nil, ErrNotBank
	}
	if cash <= 0 {
		return nil, ErrWrongCash
	}
	var bank Bank
	_, value, exists := app.state.Get([]byte(KeyBank(bankAddr)))
	if exists {
		err := UnmarshalMessage(value, &bank)
		if err != nil {
			return nil, ErrStorage
		}
	}
	bank.Rmb += cash
	save, err := MarshalMessage(&bank)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyBank(bankAddr)), save)
	event := &EventIncreaseBankRmb{}
	event.Balance = bank.Rmb
	return &Response{Value: &Response_IncreaseBankRmb{&ResponseIncreaseBankRmb{InstructionId: instructionId, Event: &Event{Value: &Event_IncreaseBankRmb{event}}}}}, nil
}
