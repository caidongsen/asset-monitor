package provinvest

import (
//"bytes"
)

func (app *ProvinvestApplication) checkOperationRecord(pubkey []byte, version string) error {

	// if app.getAdminType(pubkey) != AdminType_A_NORMAL {
	// 	return ErrNotAdmin
	// }
	return nil
}

func (app *ProvinvestApplication) operationRecord(pubkey []byte, version string, record string, instructionId int64) (*Response, error) {
	_, _, exist := app.state.Get([]byte(KeyRecord(instructionId)))
	if exist {
		return nil, ErrDupInstructionId
	}

	operationRecord := &OperationRecord{}
	operationRecord.RecordVersion = version
	operationRecord.Record = record
	save, err := MarshalMessage(operationRecord)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyRecord(instructionId)), save)
	event := &EventOperationRecord{}
	event.InstructionId = instructionId
	return &Response{Value: &Response_OperationRecord{&ResponseOperationRecord{InstructionId: instructionId, Event: &Event{Value: &Event_OperationRecord{event}}}}}, nil
}
