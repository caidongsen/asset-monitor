package smsc

import (
	"bytes"
	//"log"
)

func (app *SmscApplication) checkSetAdmin(pubkey []byte) error {
	if !app.isAdmin(pubkey) {
		return ErrNotAdmin
	}

	return nil
}

func (app *SmscApplication) setAdmin(pubkey []byte, adminPubkey []byte, instructionId int64) (*Response, error) {
	err := app.checkSetAdmin(pubkey)
	if err != nil {
		return nil, err
	}
	admin := &Admin{}
	admin.Pubkey = adminPubkey[:]
	save, err := MarshalMessage(admin)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyAdmin()), save)
	event := &EventSetAdmin{}
	event.Pubkey = adminPubkey[:]
	return &Response{Value: &Response_SetAdmin{&ResponseSetAdmin{InstructionId: instructionId, Event: &Event{Value: &Event_SetAdmin{event}}}}}, nil
}

func (app *SmscApplication) checkCreateAccount(pubkey []byte, role Role) error {
	if !app.isAdmin(pubkey) {
		return ErrNotAdmin
	}
	if err := checkRole(role); err != nil {
		return err
	}

	return nil
}

func (app *SmscApplication) createAccount(pubkey []byte, id int64, userPubkey []byte, account, name string, role Role, instructionId int64) (*Response, error) {
	err := app.checkCreateAccount(pubkey, role)
	if err != nil {
		return nil, err
	}
	switch role {
	case Role_RPlanner:
		planner := &Planner{}
		planner.Pubkey = userPubkey[:]
		planner.Id = id
		planner.Account = account
		planner.Name = name
		save, err := MarshalMessage(planner)
		if err != nil {
			return nil, err
		}
		app.state.Set([]byte(KeyPlanner(id)), save)
	case Role_RSupplier:
		supplier := &Supplier{}
		supplier.Pubkey = userPubkey[:]
		supplier.Id = id
		supplier.Account = account
		supplier.Name = name
		save, err := MarshalMessage(supplier)
		if err != nil {
			return nil, err
		}
		app.state.Set([]byte(KeySupplier(id)), save)
	case Role_RCarrier:
		carrier := &Carrier{}
		carrier.Pubkey = userPubkey[:]
		carrier.Id = id
		carrier.Account = account
		carrier.Name = name
		save, err := MarshalMessage(carrier)
		if err != nil {
			return nil, err
		}
		app.state.Set([]byte(KeyCarrier(id)), save)
	case Role_RChecker:
		checker := &Checker{}
		checker.Pubkey = userPubkey[:]
		checker.Id = id
		checker.Account = account
		checker.Name = name
		save, err := MarshalMessage(checker)
		if err != nil {
			return nil, err
		}
		app.state.Set([]byte(KeyChecker(id)), save)
	}

	event := &EventCreateAccount{}
	event.Id = id
	return &Response{Value: &Response_CreateAccount{&ResponseCreateAccount{InstructionId: instructionId, Event: &Event{Value: &Event_CreateAccount{event}}}}}, nil
}

func (app *SmscApplication) checkEditAccount(pubkey []byte, id int64, role Role) error {
	user, err := app.getUser(id, role)
	if err != nil {
		return err
	}
	var userPubkey []byte
	switch role {
	case Role_RPlanner:
		userPubkey = user.(Planner).Pubkey[:]
	case Role_RSupplier:
		userPubkey = user.(Supplier).Pubkey[:]
	case Role_RCarrier:
		userPubkey = user.(Carrier).Pubkey[:]
	case Role_RChecker:
		userPubkey = user.(Checker).Pubkey[:]
	}
	if !app.isAdmin(pubkey) && !bytes.Equal(pubkey, userPubkey) {
		return ErrNoRight
	}

	return nil
}

func (app *SmscApplication) editAccount(pubkey []byte, id int64, userPubkey []byte, account, name string, role Role, instructionId int64) (*Response, error) {
	err := app.checkEditAccount(pubkey, id, role)
	if err != nil {
		return nil, err
	}
	switch role {
	case Role_RPlanner:
		planner := &Planner{}
		planner.Pubkey = userPubkey[:]
		planner.Id = id
		planner.Account = account
		planner.Name = name
		save, err := MarshalMessage(planner)
		if err != nil {
			return nil, err
		}
		app.state.Set([]byte(KeyPlanner(id)), save)
	case Role_RSupplier:
		supplier := &Supplier{}
		supplier.Pubkey = userPubkey[:]
		supplier.Id = id
		supplier.Account = account
		supplier.Name = name
		save, err := MarshalMessage(supplier)
		if err != nil {
			return nil, err
		}
		app.state.Set([]byte(KeySupplier(id)), save)
	case Role_RCarrier:
		carrier := &Carrier{}
		carrier.Pubkey = userPubkey[:]
		carrier.Id = id
		carrier.Account = account
		carrier.Name = name
		save, err := MarshalMessage(carrier)
		if err != nil {
			return nil, err
		}
		app.state.Set([]byte(KeyCarrier(id)), save)
	case Role_RChecker:
		checker := &Checker{}
		checker.Pubkey = userPubkey[:]
		checker.Id = id
		checker.Account = account
		checker.Name = name
		save, err := MarshalMessage(checker)
		if err != nil {
			return nil, err
		}
		app.state.Set([]byte(KeyChecker(id)), save)
	}

	event := &EventEditAccount{}
	event.NewPubkey = userPubkey[:]
	return &Response{Value: &Response_EditAccount{&ResponseEditAccount{InstructionId: instructionId, Event: &Event{Value: &Event_EditAccount{event}}}}}, nil
}

func (app *SmscApplication) checkDeleteAccount(pubkey []byte, roles []Role) error {
	if !app.isAdmin(pubkey) {
		return ErrNotAdmin
	}
	for _, v := range roles {
		if err := checkRole(v); err != nil {
			return err
		}
	}

	return nil
}

func (app *SmscApplication) deleteAccount(pubkey []byte, ids []int64, roles []Role, accounts, names []string, instructionId int64) (*Response, error) {
	err := app.checkDeleteAccount(pubkey, roles)
	if err != nil {
		return nil, err
	}
	for i, v := range roles {
		switch v {
		case Role_RPlanner:
			userP, err := app.getUser(ids[i], Role_RPlanner)
			if err != nil {
				return nil, err
			}
			planner := userP.(Planner)
			for _, vv := range planner.Supplier {
				userS, err := app.getUser(vv, Role_RSupplier)
				if err != nil {
					continue
				}
				supplier := userS.(Supplier)
				supplier.Planner = 0
				save, err := MarshalMessage(&supplier)
				if err != nil {
					continue
				}
				app.state.Set([]byte(KeySupplier(vv)), save)
			}
			app.state.Remove([]byte(KeyPlanner(ids[i])))
		case Role_RSupplier:
			userS, err := app.getUser(ids[i], Role_RSupplier)
			if err != nil {
				return nil, err
			}
			supplier := userS.(Supplier)
			userP, err := app.getUser(supplier.Planner, Role_RPlanner)
			if err == nil {
				planner := userP.(Planner)
				for ii, vv := range planner.Supplier {
					if vv == ids[i] {
						tmpSupplier := planner.Supplier[:]
						planner.Supplier = planner.Supplier[:ii]
						planner.Supplier = append(planner.Supplier, tmpSupplier[ii+1:]...)
						save, err := MarshalMessage(&planner)
						if err != nil {
							break
						}
						app.state.Set([]byte(KeyPlanner(planner.Id)), save)
						break
					}
				}
			}
			app.state.Remove([]byte(KeySupplier(ids[i])))
		case Role_RCarrier:
			app.state.Remove([]byte(KeyCarrier(ids[i])))
		case Role_RChecker:
			app.state.Remove([]byte(KeyChecker(ids[i])))
		}
	}

	event := &EventDeleteAccount{}
	event.Id = ids[:]
	return &Response{Value: &Response_DeleteAccount{&ResponseDeleteAccount{InstructionId: instructionId, Event: &Event{Value: &Event_DeleteAccount{event}}}}}, nil
}

func (app *SmscApplication) checkSetSupplier(pubkey []byte, plannerId int64, supplierIdsAdd, supplierIdsDel []int64) error {
	if !app.isAdmin(pubkey) {
		return ErrNotAdmin
	}
	if _, err := app.getUser(plannerId, Role_RPlanner); err != nil {
		return err
	}
	for _, v := range supplierIdsAdd {
		if _, err := app.getUser(v, Role_RSupplier); err != nil {
			return err
		}
	}
	for _, v := range supplierIdsDel {
		if _, err := app.getUser(v, Role_RSupplier); err != nil {
			return err
		}
	}

	return nil
}

func (app *SmscApplication) setSupplier(pubkey []byte, plannerId int64, supplierIdsAdd, supplierIdsDel []int64, supplierAccountsAdd, supplierAccountsDel, supplierNamesAdd, supplierNamesDel []string, instructionId int64) (*Response, error) {
	err := app.checkSetSupplier(pubkey, plannerId, supplierIdsAdd, supplierIdsDel)
	if err != nil {
		return nil, err
	}
	user, err := app.getUser(plannerId, Role_RPlanner)
	if err != nil {
		return nil, err
	}
	planner := user.(Planner)
	for _, supplierId := range supplierIdsAdd {
		user, err = app.getUser(supplierId, Role_RSupplier)
		if err != nil {
			return nil, err
		}
		supplier := user.(Supplier)
		bExist := false
		for _, v := range planner.Supplier {
			if v == supplierId {
				bExist = true
				break
			}
		}
		if !bExist {
			planner.Supplier = append(planner.Supplier, supplierId)
			supplier.Planner = plannerId
			save, err := MarshalMessage(&planner)
			if err != nil {
				return nil, err
			}
			app.state.Set([]byte(KeyPlanner(plannerId)), save)
			save, err = MarshalMessage(&supplier)
			if err != nil {
				return nil, err
			}
			app.state.Set([]byte(KeySupplier(supplierId)), save)
		}
	}
	for _, supplierId := range supplierIdsDel {
		user, err = app.getUser(supplierId, Role_RSupplier)
		if err != nil {
			return nil, err
		}
		supplier := user.(Supplier)
		for i, v := range planner.Supplier {
			if v == supplierId {
				tmpSupplier := planner.Supplier[:]
				planner.Supplier = planner.Supplier[:i]
				planner.Supplier = append(planner.Supplier, tmpSupplier[i+1:]...)
				break
			}
		}
		supplier.Planner = 0
		save, err := MarshalMessage(&planner)
		if err != nil {
			return nil, err
		}
		app.state.Set([]byte(KeyPlanner(plannerId)), save)
		save, err = MarshalMessage(&supplier)
		if err != nil {
			return nil, err
		}
		app.state.Set([]byte(KeySupplier(supplierId)), save)
	}

	event := &EventSetSupplier{}
	event.SupplierIdAdd = supplierIdsAdd[:]
	event.SupplierIdDel = supplierIdsDel[:]
	return &Response{Value: &Response_SetSupplier{&ResponseSetSupplier{InstructionId: instructionId, Event: &Event{Value: &Event_SetSupplier{event}}}}}, nil
}

func (app *SmscApplication) checkCreateOrder(plannerId, supplierId int64) error {
	if _, err := app.getUser(plannerId, Role_RPlanner); err != nil {
		return err
	}
	if _, err := app.getUser(supplierId, Role_RSupplier); err != nil {
		return err
	}

	return nil
}

func (app *SmscApplication) createOrder(plannerId, supplierId, partNum int64, orderId, partId, boxId, requiredDate, plannerAccount, plannerName, supplierAccount, supplierName string, instructionId int64) (*Response, error) {
	err := app.checkCreateOrder(plannerId, supplierId)
	if err != nil {
		return nil, err
	}
	order := &Order{}
	order.Id = orderId
	order.PartId = partId
	order.BoxId = boxId
	order.PartNum = partNum
	order.RequiredDate = requiredDate
	order.Planner = plannerId
	order.Supplier = supplierId
	order.State = OrderState_OSTodelivered
	save, err := MarshalMessage(order)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyOrder(orderId)), save)

	event := &EventCreateOrder{}
	event.Id = orderId
	return &Response{Value: &Response_CreateOrder{&ResponseCreateOrder{InstructionId: instructionId, Event: &Event{Value: &Event_CreateOrder{event}}}}}, nil
}

func (app *SmscApplication) checkDelivery(supplierId, carrierId int64, orderId string) error {
	if _, err := app.getUser(supplierId, Role_RSupplier); err != nil {
		return err
	}
	if _, err := app.getUser(carrierId, Role_RCarrier); err != nil {
		return err
	}
	order, err := app.getOrder(orderId)
	if err != nil {
		return err
	}
	if order.Supplier != supplierId {
		return ErrWrongSupplier
	}

	return nil
}

func (app *SmscApplication) delivery(supplierId, carrierId, partNum int64, orderId, partId, boxId, deliveryDate, supplierAccount, supplierName, carrierAccount, carrierName string, instructionId int64) (*Response, error) {
	err := app.checkDelivery(supplierId, carrierId, orderId)
	if err != nil {
		return nil, err
	}
	order, err := app.getOrder(orderId)
	if err != nil {
		return nil, err
	}
	order.Id = orderId
	order.PartId = partId
	order.BoxId = boxId
	order.PartNum = partNum
	order.DeliveryDate = deliveryDate
	order.Carrier = carrierId
	order.State = OrderState_OSToCarried
	save, err := MarshalMessage(&order)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyOrder(orderId)), save)

	event := &EventDelivery{}
	event.Carrier = carrierId
	return &Response{Value: &Response_Delivery{&ResponseDelivery{InstructionId: instructionId, Event: &Event{Value: &Event_Delivery{event}}}}}, nil
}

func (app *SmscApplication) checkCarry(carrierId int64, orderId string) error {
	if _, err := app.getUser(carrierId, Role_RCarrier); err != nil {
		return err
	}
	order, err := app.getOrder(orderId)
	if err != nil {
		return err
	}
	if order.Carrier != carrierId {
		return ErrWrongCarrier
	}

	return nil
}

func (app *SmscApplication) carry(carrierId int64, orderId, boxId, carId, carryDate, carrierAccount, carrierName string, boxNum, instructionId int64) (*Response, error) {
	err := app.checkCarry(carrierId, orderId)
	if err != nil {
		return nil, err
	}
	order, err := app.getOrder(orderId)
	if err != nil {
		return nil, err
	}
	order.Id = orderId
	order.BoxId = boxId
	order.CarryDate = carryDate
	order.CarId = carId
	order.State = OrderState_OSCarrying
	save, err := MarshalMessage(&order)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyOrder(orderId)), save)

	event := &EventCarry{}
	event.CarId = carId
	return &Response{Value: &Response_Carry{&ResponseCarry{InstructionId: instructionId, Event: &Event{Value: &Event_Carry{event}}}}}, nil
}

func (app *SmscApplication) checkCheck(checkerId int64, orderId string) error {
	if _, err := app.getUser(checkerId, Role_RChecker); err != nil {
		return err
	}
	_, err := app.getOrder(orderId)
	if err != nil {
		return err
	}

	return nil
}

func (app *SmscApplication) check(checkerId int64, orderId, checkDate string, op Operate, checkerAccount, checkerName, partId, boxId, carId string, boxNum, instructionId int64) (*Response, error) {
	err := app.checkCheck(checkerId, orderId)
	if err != nil {
		return nil, err
	}
	order, err := app.getOrder(orderId)
	if err != nil {
		return nil, err
	}
	order.Checker = checkerId
	order.CheckDate = checkDate
	if op == Operate_ORefuse {
		order.State = OrderState_OSRefused
	}
	if op == Operate_OPass {
		order.State = OrderState_OSChecked
	}
	save, err := MarshalMessage(&order)
	if err != nil {
		return nil, err
	}
	app.state.Set([]byte(KeyOrder(orderId)), save)

	event := &EventCheck{}
	event.OrderId = orderId
	return &Response{Value: &Response_Check{&ResponseCheck{InstructionId: instructionId, Event: &Event{Value: &Event_Check{event}}}}}, nil
}
