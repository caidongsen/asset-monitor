package mideaSupply

import (
	"bytes"
	"fmt"
	"strconv"
)

func (app *MideaSupplyApplication) checkInitPlatform(req *Request) error {
	_, _, exists := app.state.Get([]byte(KeyPlatform()))
	if exists {
		return ErrPlatformExists
	}
	fmt.Println(req.GetPubkey(), req.GetActionId(), req.GetInitPlatform().Info)
	if req.GetInitPlatform() == nil {
		return ErrEmptyValue
	}
	if !isOriginalAdmin(req.GetPubkey()) {
		return ErrNoRight
	}
	if len(req.GetInitPlatform().PlatformKey) != 32 {
		return ErrWrongPubkey
	}
	return nil
}

func (app *MideaSupplyApplication) initPlatform(req *Request) (*Response, error) {
	err1 := app.checkInitPlatform(req)
	if err1 != nil {
		return nil, err1
	}

	platform := &Platform{}
	platform.Pubkey = req.GetInitPlatform().PlatformKey
	platform.Info = req.GetInitPlatform().Info
	save, err := MarshalMessage(platform)
	if err != nil {
		return nil, ErrStorage
	}
	app.state.Set([]byte(KeyPlatform()), save)
	instructionId := req.GetInstructionId()
	event := &EventInitPlatform{}
	event.PlatformKey = req.GetInitPlatform().PlatformKey
	return &Response{Value: &Response_InitPlatform{&ResponseInitPlatform{InstructionId: instructionId, Event: &Event{Value: &Event_InitPlatform{event}}}}}, nil
}

func (app *MideaSupplyApplication) checkRegisterSupplier(req *Request) error {
	if err := app.isPlatformAdmin(req.GetPubkey()); err != nil {
		return err
	}
	_, _, exists := app.state.Get([]byte(KeySupplier(req.GetRegisterSupplier().Supplier.VendorId)))
	if exists {
		return ErrSupplierExists
	}
	if req.GetRegisterSupplier() == nil {
		return ErrEmptyValue
	}
	return nil
}

func (app *MideaSupplyApplication) registerSupplier(req *Request) (*Response, error) {
	err1 := app.checkRegisterSupplier(req)
	if err1 != nil {
		return nil, err1
	}
	supplier := &Supplier{}
	supplier.VendorId = req.GetRegisterSupplier().Supplier.VendorId
	supplier.VendorCode = req.GetRegisterSupplier().Supplier.VendorCode
	supplier.UserPubkey = req.GetRegisterSupplier().Supplier.UserPubkey
	supplier.VendorName = req.GetRegisterSupplier().Supplier.VendorName
	save, err := MarshalMessage(supplier)
	if err != nil {
		return nil, ErrStorage
	}
	app.state.Set([]byte(KeySupplier(req.GetRegisterSupplier().Supplier.VendorId)), save)
	instructionId := req.GetInstructionId()
	event := &EventRegisterSupplier{}
	event.Pubkey = req.GetRegisterSupplier().Supplier.UserPubkey
	return &Response{Value: &Response_RegisterSupplier{&ResponseRegisterSupplier{InstructionId: instructionId, Event: &Event{Value: &Event_RegisterSupplier{event}}}}}, nil
}

func (app *MideaSupplyApplication) checkRegisterSupplierList(req *Request) error {
	if err := app.isPlatformAdmin(req.GetPubkey()); err != nil {
		return err
	}
	if len(req.GetRegisterSupplierList().SupplierList) > 1000 {
		return ErrOverMaxLimit
	}
	list := make(map[int64]int64, 0)
	for _, v := range req.GetRegisterSupplierList().SupplierList {
		_, ok := list[v.VendorId]
		if ok {
			return ErrDupVendorId
		}

		_, _, exists := app.state.Get([]byte(KeySupplier(v.VendorId)))
		if exists {
			return ErrSupplierExists
		}
		if len(v.UserPubkey) != 32 {
			return ErrWrongPubkey
		}
		list[v.VendorId] = v.VendorId
	}

	if req.GetRegisterSupplierList() == nil {
		return ErrEmptyValue
	}
	return nil
}

func (app *MideaSupplyApplication) registerSupplierList(req *Request) (*Response, error) {
	err1 := app.checkRegisterSupplierList(req)
	if err1 != nil {
		return nil, err1
	}
	for _, v := range req.GetRegisterSupplierList().SupplierList {
		supplier := &Supplier{}
		supplier.VendorId = v.VendorId
		supplier.VendorCode = v.VendorCode
		supplier.UserPubkey = v.UserPubkey
		supplier.VendorName = v.VendorName
		save, err := MarshalMessage(supplier)
		if err != nil {
			return nil, ErrStorage
		}
		app.state.Set([]byte(KeySupplier(v.VendorId)), save)
	}
	instructionId := req.GetInstructionId()
	event := &EventRegisterSupplierList{}
	event.InstructionId = instructionId
	return &Response{Value: &Response_RegisterSupplierList{&ResponseRegisterSupplierList{InstructionId: instructionId, Event: &Event{Value: &Event_RegisterSupplierList{event}}}}}, nil
}

func (app *MideaSupplyApplication) checkWarehouseEntry(req *Request) error {
	uid, err := strconv.ParseInt(req.GetUid(), 10, 64)
	_, value, exists := app.state.Get([]byte(KeySupplier(uid)))
	if !exists {
		return ErrSupplierNotRegister
	}
	var supplier Supplier
	err = UnmarshalMessage(value, &supplier)
	if err != nil {
		return ErrStorage
	}
	if !bytes.Equal(req.GetPubkey(), supplier.UserPubkey) {
		return ErrNoRight
	}
	if req.GetWarehouseEntry() == nil {
		return ErrEmptyValue
	}
	if req.GetWarehouseEntry().ReceiveHeader.RecHeaderId <= int64(0) {
		return ErrWrongRecHeaderId
	}
	mlist := make(map[int64]int64, 0)

	for _, v := range req.GetWarehouseEntry().ReceiveHeader.ReceiveLines {
		if v.ShipmentHeaderId <= int64(0) {
			return ErrWrongHeaderId
		}
		if v.ShipmentHeaderId != req.GetWarehouseEntry().ReceiveHeader.ShipmentHeaderId {
			return ErrWrongHeaderId
		}
		if v.RcvTranId <= int64(0) {
			return ErrEmptyTranId
		}

		_, _, exists := app.state.Get([]byte(KeyTranId(v.RcvTranId)))
		if exists {
			return ErrTranIdExists
		}
		_, ok := mlist[v.RcvTranId]
		if ok {
			return ErrDupTranId
		}
		mlist[v.RcvTranId] = v.RcvTranId
	}

	_, _, exists = app.state.Get([]byte(KeyEnrty(req.GetWarehouseEntry().ReceiveHeader.RecHeaderId)))
	if exists {
		return ErrRecDupHeaderId
	}

	return nil
}

func (app *MideaSupplyApplication) warehouseEntry(req *Request) (*Response, error) {
	err1 := app.checkWarehouseEntry(req)
	if err1 != nil {
		return nil, err1
	}
	var linenum, lineamount int64
	receiveHeder := &ReceiveHeader{}
	receiveHeder.RecHeaderId = req.GetWarehouseEntry().ReceiveHeader.RecHeaderId
	receiveHeder.ShipmentHeaderId = req.GetWarehouseEntry().ReceiveHeader.ShipmentHeaderId
	receiveHeder.ReceiptNum = req.GetWarehouseEntry().ReceiveHeader.ReceiptNum
	receiveHeder.ReceiveDate = req.GetWarehouseEntry().ReceiveHeader.ReceiveDate
	receiveHeder.ReceivePerson = req.GetWarehouseEntry().ReceiveHeader.ReceivePerson
	receiveHeder.PurchansePerson = req.GetWarehouseEntry().ReceiveHeader.PurchansePerson
	receiveHeder.OrganizationId = req.GetWarehouseEntry().ReceiveHeader.OrganizationId
	receiveHeder.OrganizationCode = req.GetWarehouseEntry().ReceiveHeader.OrganizationCode
	receiveHeder.VendorId = req.GetWarehouseEntry().ReceiveHeader.VendorId
	receiveHeder.VendorCode = req.GetWarehouseEntry().ReceiveHeader.VendorCode
	receiveHeder.VendorName = req.GetWarehouseEntry().ReceiveHeader.VendorName
	receiveHeder.VendorSiteId = req.GetWarehouseEntry().ReceiveHeader.VendorSiteId
	receiveHeder.VendorSiteCode = req.GetWarehouseEntry().ReceiveHeader.VendorSiteCode
	receiveHeder.AttributeCategory = req.GetWarehouseEntry().ReceiveHeader.AttributeCategory
	receiveHeder.Attribute1 = req.GetWarehouseEntry().ReceiveHeader.Attribute1
	receiveHeder.Attribute2 = req.GetWarehouseEntry().ReceiveHeader.Attribute2
	receiveHeder.Attribute3 = req.GetWarehouseEntry().ReceiveHeader.Attribute3
	receiveHeder.Attribute4 = req.GetWarehouseEntry().ReceiveHeader.Attribute4
	receiveHeder.Attribute5 = req.GetWarehouseEntry().ReceiveHeader.Attribute5
	receiveHeder.Attribute6 = req.GetWarehouseEntry().ReceiveHeader.Attribute6
	receiveHeder.Attribute7 = req.GetWarehouseEntry().ReceiveHeader.Attribute7
	receiveHeder.Attribute8 = req.GetWarehouseEntry().ReceiveHeader.Attribute8
	receiveHeder.Attribute9 = req.GetWarehouseEntry().ReceiveHeader.Attribute9
	receiveHeder.Attribute10 = req.GetWarehouseEntry().ReceiveHeader.Attribute10
	receiveHeder.Attribute11 = req.GetWarehouseEntry().ReceiveHeader.Attribute11
	receiveHeder.Attribute12 = req.GetWarehouseEntry().ReceiveHeader.Attribute12
	receiveHeder.Attribute13 = req.GetWarehouseEntry().ReceiveHeader.Attribute13
	receiveHeder.Attribute14 = req.GetWarehouseEntry().ReceiveHeader.Attribute14
	receiveHeder.Attribute15 = req.GetWarehouseEntry().ReceiveHeader.Attribute15
	receiveHeder.ReceiveLines = req.GetWarehouseEntry().ReceiveHeader.ReceiveLines

	_, value, exists := app.state.Get([]byte(KeySupplier(req.GetWarehouseEntry().ReceiveHeader.VendorId)))
	if !exists {
		return nil, ErrSupplierNotRegister
	}
	var supplier Supplier
	err := UnmarshalMessage(value, &supplier)
	if err != nil {
		return nil, ErrStorage
	}
	check := false
	for _, v := range supplier.SupplierSite {
		if v.SiteId == req.GetWarehouseEntry().ReceiveHeader.VendorSiteId {
			check = true
			break
		}
	}
	if !check {
		var supplierSite SupplierSite
		//supplierSite.Site = req.GetWarehouseEntry().ReceiveHeader.VendorSite
		supplierSite.SiteCode = req.GetWarehouseEntry().ReceiveHeader.VendorSiteCode
		supplierSite.SiteId = req.GetWarehouseEntry().ReceiveHeader.VendorSiteId
		supplier.SupplierSite = append(supplier.SupplierSite, &supplierSite)
	}

	rightLines := make([]*ReceiveLine, 0)
	returnLines := make([]*ReceiveLine, 0)
	for _, v := range req.GetWarehouseEntry().ReceiveHeader.ReceiveLines {
		if v.RcvTranType == "RETURN TO VENDOR" {
			// for _, vv := range rightLines {
			// 	if v.TopTranId == vv.TopTranId {
			// 		vv.Quantity = vv.Quantity - v.ReturnQuantity
			// 		break
			// 	}
			// }
			lineamount = lineamount - (v.Quantity*v.PoUnitPrice)/QUANTITY_ZERO_LIMIT
			returnLines = append(returnLines, v)
		} else {
			rightLines = append(rightLines, v)
			lineamount = lineamount + (v.Quantity*v.PoUnitPrice)/QUANTITY_ZERO_LIMIT
		}

	}
	linenum = int64(len(rightLines))
	tempLine := make([]int64, 0)
	for _, v := range receiveHeder.ReceiveLines {
		receiveLine := &ReceiveLine{}
		receiveLine.RcvTranId = v.RcvTranId
		receiveLine.ShipmentHeaderId = v.ShipmentHeaderId
		receiveLine.ItemId = v.ItemId
		receiveLine.ItemCode = v.ItemCode
		//receiveLine.ItemName = v.ItemName
		receiveLine.ItemDesc = v.ItemDesc
		receiveLine.PrimaryUnit = v.PrimaryUnit
		receiveLine.PoHeaderId = v.PoHeaderId
		receiveLine.PoLineId = v.PoLineId
		//receiveLine.LineLocationId = v.LineLocationId
		receiveLine.PoNum = v.PoNum
		receiveLine.LineNum = v.LineNum
		receiveLine.CurrencyCode = v.CurrencyCode
		receiveLine.CurrencyConversionRate = v.CurrencyConversionRate
		receiveLine.CurrencyConversionDate = v.CurrencyConversionDate
		receiveLine.Quantity = v.Quantity
		receiveLine.PoUnitPrice = v.PoUnitPrice
		receiveLine.PriceMatched = v.PriceMatched
		receiveLine.QuantityMatched = v.QuantityMatched
		receiveLine.AmountMatched = v.AmountMatched
		receiveLine.Remark = v.Remark
		receiveLine.RcvTranType = v.RcvTranType
		receiveLine.RcvTranDate = v.RcvTranDate
		receiveLine.ReturnQuantity = v.ReturnQuantity
		receiveLine.TopTranId = v.TopTranId
		receiveLine.AttributeCategory = v.AttributeCategory
		receiveLine.Attribute1 = v.Attribute1
		receiveLine.Attribute2 = v.Attribute2
		receiveLine.Attribute3 = v.Attribute3
		receiveLine.Attribute4 = v.Attribute4
		receiveLine.Attribute5 = v.Attribute5
		receiveLine.Attribute6 = v.Attribute6
		receiveLine.Attribute7 = v.Attribute7
		receiveLine.Attribute8 = v.Attribute8
		receiveLine.Attribute9 = v.Attribute9
		receiveLine.Attribute10 = v.Attribute10
		receiveLine.Attribute11 = v.Attribute11
		receiveLine.Attribute12 = v.Attribute12
		receiveLine.Attribute13 = v.Attribute13
		receiveLine.Attribute14 = v.Attribute14
		receiveLine.Attribute15 = v.Attribute15
		save, err := MarshalMessage(receiveLine)
		if err != nil {
			return nil, ErrStorage
		}

		app.state.Set([]byte(KeyTranId(v.RcvTranId)), save)
	}

	for _, v := range rightLines {

		supplier.Rcvlines = append(supplier.Rcvlines, v.RcvTranId)
		tempLine = append(tempLine, v.PoLineId)
		uninvline := &UnInvoiceLine{}
		uninvline.VendorCode = req.GetWarehouseEntry().ReceiveHeader.VendorCode
		uninvline.VendorName = req.GetWarehouseEntry().ReceiveHeader.VendorName
		uninvline.VendorSiteCode = req.GetWarehouseEntry().ReceiveHeader.VendorSiteCode
		uninvline.ItemCode = v.ItemCode
		uninvline.ItemDesc = v.ItemDesc
		uninvline.CurrencyCode = v.CurrencyCode
		uninvline.PoNum = v.PoNum
		uninvline.ReceiptNum = req.GetWarehouseEntry().ReceiveHeader.ReceiptNum
		uninvline.ReceiveDate = v.RcvTranDate
		uninvline.PrimayUnit = v.PrimaryUnit
		uninvline.QuantityReceived = v.Quantity
		uninvline.PoUnitPrice = v.PoUnitPrice
		uninvline.OrganizationId = req.GetWarehouseEntry().ReceiveHeader.OrganizationId
		uninvline.TranId = v.RcvTranId
		supplier.UninvLines = append(supplier.UninvLines, uninvline)
		//app.state.Set([]byte(KeyTranId(v.RcvTranId)), []byte(strconv.FormatInt(v.RcvTranId, 10)))
	}
	for _, v := range returnLines {
		_, value, exists := app.state.Get([]byte(KeyTranId(v.RcvTranId)))
		if !exists {
			return nil, ErrEmptyTranId
		}
		var revline ReceiveLine
		err := UnmarshalMessage(value, &revline)
		if err != nil {
			return nil, ErrStorage
		}
		//revline.Quantity = revline.Quantity - v.ReturnQuantity
		revline.ReturnQuantity = v.ReturnQuantity
		save, err := MarshalMessage(&revline)
		if err != nil {
			return nil, ErrStorage
		}
		app.state.Set([]byte(KeyTranId(v.RcvTranId)), save)
	}

	supplier.Unlinenum = supplier.Unlinenum + linenum
	supplier.Unlineamount = supplier.Unlineamount + lineamount
	save, err := MarshalMessage(receiveHeder)
	if err != nil {
		return nil, ErrStorage
	}
	app.state.Set([]byte(KeyEnrty(req.GetWarehouseEntry().ReceiveHeader.RecHeaderId)), save)
	save, err = MarshalMessage(&supplier)
	if err != nil {
		return nil, ErrStorage
	}
	app.state.Set([]byte(KeyEnrty(req.GetWarehouseEntry().ReceiveHeader.VendorId)), save)
	instructionId := req.GetInstructionId()
	event := &EventWarehouseEntry{}
	event.HeaderId = req.GetWarehouseEntry().ReceiveHeader.RecHeaderId
	event.Line = tempLine
	return &Response{Value: &Response_WarehouseEntry{&ResponseWarehouseEntry{InstructionId: instructionId, Event: &Event{Value: &Event_WarehouseEntry{event}}}}}, nil
}

func (app *MideaSupplyApplication) checkOpenInvoice(req *Request) error {
	uid, err := strconv.ParseInt(req.GetUid(), 10, 64)
	_, value, exists := app.state.Get([]byte(KeySupplier(uid)))
	if !exists {
		return ErrSupplierNotRegister
	}
	var supplier Supplier
	err = UnmarshalMessage(value, &supplier)
	if err != nil {
		return ErrStorage
	}
	if !bytes.Equal(req.GetPubkey(), supplier.UserPubkey) {
		return ErrNoRight
	}
	_, _, exists = app.state.Get([]byte(KeyInvoice(req.GetOpenInvoice().InvoiceHeader.IspInvoiceId)))
	if exists {
		return ErrInvoiceExists
	}
	if req.GetOpenInvoice() == nil {
		return ErrEmptyValue
	}
	if req.GetOpenInvoice().InvoiceHeader.IspInvoiceId <= int64(0) {
		return ErrEmptyInvoiceHeaderId
	}
	if len(req.GetOpenInvoice().InvoiceHeader.InvoiceLine) > 5000 {
		return ErrOverMaxLimit
	}
	check := false
	lineTranList := make(map[int64]int64, 0)
	for _, v := range req.GetOpenInvoice().InvoiceHeader.InvoiceLine {
		if v.IspInvoiceLineId <= int64(0) {
			check = true
			break
		}
		if v.IspTranId <= int64(0) {
			return ErrEmptyTranId
		}
		_, value, exists := app.state.Get([]byte(KeyTranId(v.IspTranId)))
		if !exists {
			return ErrTranIdNotExists
		}
		var rcvLine ReceiveLine
		err := UnmarshalMessage(value, &rcvLine)
		if err != nil {
			return ErrStorage
		}
		if rcvLine.Quantity < v.Quantity+rcvLine.QuantityMatched {
			return ErrQuantityException
		}
		_, ok := lineTranList[v.IspTranId]
		if ok {
			return ErrDupTranId
		}
		lineTranList[v.IspTranId] = v.IspTranId
	}
	if check {
		return ErrEmptyInvoiceLineId
	}
	return nil
}

func (app *MideaSupplyApplication) openInvoice(req *Request) (*Response, error) {
	err1 := app.checkOpenInvoice(req)
	if err1 != nil {
		return nil, err1
	}
	uid, err := strconv.ParseInt(req.GetUid(), 10, 64)
	_, value, exists := app.state.Get([]byte(KeySupplier(uid)))
	if !exists {
		return nil, ErrSupplierNotRegister
	}
	var supplier Supplier
	err = UnmarshalMessage(value, &supplier)
	if err != nil {
		return nil, ErrStorage
	}
	invoice := &InvoiceHeader{}
	invoice.IspInvoiceId = req.GetOpenInvoice().InvoiceHeader.IspInvoiceId
	invoice.IspInvoiceCode = req.GetOpenInvoice().InvoiceHeader.IspInvoiceCode
	invoice.OrgId = req.GetOpenInvoice().InvoiceHeader.OrgId
	invoice.OrgName = req.GetOpenInvoice().InvoiceHeader.OrgName
	invoice.SourceCode = req.GetOpenInvoice().InvoiceHeader.SourceCode
	invoice.VendorId = req.GetOpenInvoice().InvoiceHeader.VendorId
	invoice.VendorCode = req.GetOpenInvoice().InvoiceHeader.VendorCode
	invoice.VendorName = req.GetOpenInvoice().InvoiceHeader.VendorName
	invoice.VendorSiteId = req.GetOpenInvoice().InvoiceHeader.VendorSiteId
	invoice.VendorSiteCode = req.GetOpenInvoice().InvoiceHeader.VendorSiteCode
	//invoice.VendorSite = req.GetOpenInvoice().InvoiceHeader.VendorSite
	invoice.TcmNoTaxAmount = req.GetOpenInvoice().InvoiceHeader.TcmNoTaxAmount
	invoice.TaxRate = req.GetOpenInvoice().InvoiceHeader.TaxRate
	invoice.TaxAmount = req.GetOpenInvoice().InvoiceHeader.TaxAmount
	invoice.CurrencyCode = req.GetOpenInvoice().InvoiceHeader.CurrencyCode
	invoice.CurrencyConversionRate = req.GetOpenInvoice().InvoiceHeader.CurrencyConversionRate
	invoice.CurrencyConversionDate = req.GetOpenInvoice().InvoiceHeader.CurrencyConversionDate
	invoice.Comments = req.GetOpenInvoice().InvoiceHeader.Comments
	invoice.ApInvoiceNumber = req.GetOpenInvoice().InvoiceHeader.ApInvoiceNumber
	invoice.GlDate = req.GetOpenInvoice().InvoiceHeader.GlDate
	invoice.InvoiceState = InvoiceState_IS_UNCHECKED
	invoice.InvoiceStatus = req.GetOpenInvoice().InvoiceHeader.InvoiceStatus
	invoice.AttributeCategory = req.GetOpenInvoice().InvoiceHeader.AttributeCategory
	invoice.Attribute1 = req.GetOpenInvoice().InvoiceHeader.Attribute1
	invoice.Attribute2 = req.GetOpenInvoice().InvoiceHeader.Attribute2
	invoice.Attribute3 = req.GetOpenInvoice().InvoiceHeader.Attribute3
	invoice.Attribute4 = req.GetOpenInvoice().InvoiceHeader.Attribute4
	invoice.Attribute5 = req.GetOpenInvoice().InvoiceHeader.Attribute5
	invoice.Attribute6 = req.GetOpenInvoice().InvoiceHeader.Attribute6
	invoice.Attribute7 = req.GetOpenInvoice().InvoiceHeader.Attribute7
	invoice.Attribute8 = req.GetOpenInvoice().InvoiceHeader.Attribute8
	invoice.Attribute9 = req.GetOpenInvoice().InvoiceHeader.Attribute9
	invoice.Attribute10 = req.GetOpenInvoice().InvoiceHeader.Attribute10
	invoice.Attribute11 = req.GetOpenInvoice().InvoiceHeader.Attribute11
	invoice.Attribute12 = req.GetOpenInvoice().InvoiceHeader.Attribute12
	invoice.Attribute13 = req.GetOpenInvoice().InvoiceHeader.Attribute13
	invoice.Attribute14 = req.GetOpenInvoice().InvoiceHeader.Attribute14
	invoice.Attribute15 = req.GetOpenInvoice().InvoiceHeader.Attribute15

	tempLine := make([]int64, 0)
	for _, v := range req.GetOpenInvoice().InvoiceHeader.InvoiceLine {
		invoiceLine := &InvoiceLine{}
		invoiceLine.IspInvoiceId = v.IspInvoiceId
		invoiceLine.IspInvoiceLineId = v.IspInvoiceLineId
		invoiceLine.IspTranId = v.IspTranId
		invoiceLine.OrganizationId = v.OrganizationId
		invoiceLine.OrganizationCode = v.OrganizationCode
		invoiceLine.LineNum = v.LineNum
		invoiceLine.InventoryItemId = v.InventoryItemId
		invoiceLine.ItemCode = v.ItemCode
		//invoiceLine.ItemName = v.ItemName
		invoiceLine.ItemDesc = v.ItemDesc
		invoiceLine.ItemUom = v.ItemUom
		invoiceLine.Quantity = v.Quantity
		invoiceLine.Price = v.Price
		invoiceLine.ShareFlag = v.ShareFlag
		invoiceLine.LineType = v.LineType
		invoiceLine.PenaltyId = v.PenaltyId
		invoiceLine.OverduePenaltyRate = v.OverduePenaltyRate
		invoiceLine.LineAmount = v.LineAmount
		invoiceLine.PoPrice = v.PoPrice
		invoiceLine.AttributeCategory = v.AttributeCategory
		invoiceLine.Attribute1 = v.Attribute1
		invoiceLine.Attribute2 = v.Attribute2
		invoiceLine.Attribute3 = v.Attribute3
		invoiceLine.Attribute4 = v.Attribute4
		invoiceLine.Attribute5 = v.Attribute5
		invoiceLine.Attribute6 = v.Attribute6
		invoiceLine.Attribute7 = v.Attribute7
		invoiceLine.Attribute8 = v.Attribute8
		invoiceLine.Attribute9 = v.Attribute9
		invoiceLine.Attribute10 = v.Attribute10
		invoiceLine.Attribute11 = v.Attribute11
		invoiceLine.Attribute12 = v.Attribute12
		invoiceLine.Attribute13 = v.Attribute13
		invoiceLine.Attribute14 = v.Attribute14
		invoiceLine.Attribute15 = v.Attribute15

		_, value, exists := app.state.Get([]byte(KeyTranId(v.IspTranId)))
		if !exists {
			return nil, ErrTranIdNotExists
		}
		var revline ReceiveLine
		err := UnmarshalMessage(value, &revline)
		if err != nil {
			return nil, ErrStorage
		}
		revline.QuantityMatched += v.Quantity
		revline.AmountMatched += v.LineAmount
		revline.PriceMatched = v.Price
		save, err := MarshalMessage(&revline)
		if err != nil {
			return nil, ErrStorage
		}
		app.state.Set([]byte(KeyTranId(v.IspTranId)), save)
		invoice.InvoiceLine = append(invoice.InvoiceLine, invoiceLine)
		if (revline.Quantity - revline.ReturnQuantity) == revline.QuantityMatched {
			tempLine = append(tempLine, v.IspTranId)
		}

	}

	restLine := make([]*UnInvoiceLine, 0)
	for _, v := range supplier.UninvLines {
		for _, vv := range tempLine {
			if v.TranId != vv {
				restLine = append(restLine, v)
			}
		}
	}
	supplier.UninvLines = restLine
	save, err := MarshalMessage(invoice)
	if err != nil {
		return nil, ErrStorage
	}
	app.state.Set([]byte(KeyInvoice(req.GetOpenInvoice().InvoiceHeader.IspInvoiceId)), save)

	save, err = MarshalMessage(&supplier)
	if err != nil {
		return nil, ErrStorage
	}
	app.state.Set([]byte(KeySupplier(invoice.VendorId)), save)
	instructionId := req.GetInstructionId()
	event := &EventOpenInvoice{}
	event.InvoiceHeader = req.GetOpenInvoice().InvoiceHeader.IspInvoiceId
	event.Line = tempLine
	return &Response{Value: &Response_OpenInvoice{&ResponseOpenInvoice{InstructionId: instructionId, Event: &Event{Value: &Event_OpenInvoice{event}}}}}, nil
}

func (app *MideaSupplyApplication) checkCheckInvoice(req *Request) error {
	if err := app.isPlatformAdmin(req.GetPubkey()); err != nil {
		return err
	}
	if req.GetCheckInvoice() == nil {
		return ErrEmptyValue
	}
	if req.GetCheckInvoice().IsInvoiceId <= int64(0) {
		return ErrEmptyInvoiceHeaderId
	}
	_, value, exists := app.state.Get([]byte(KeyInvoice(req.GetCheckInvoice().IsInvoiceId)))
	if !exists {
		return ErrEmptyInvoiceHeaderId
	}
	var invoice InvoiceHeader
	err := UnmarshalMessage(value, &invoice)
	if err != nil {
		return ErrStorage
	}
	if invoice.InvoiceState == InvoiceState_IS_CHECKED {
		return ErrInvoiceChecked
	} else {
		if !req.GetCheckInvoice().IsPass && invoice.InvoiceState == InvoiceState_IS_REJECTED {
			return ErrWrongState
		}
	}

	return nil
}

func (app *MideaSupplyApplication) checkInvoice(req *Request) (*Response, error) {
	err1 := app.checkCheckInvoice(req)
	if err1 != nil {
		return nil, err1
	}
	_, value, exists := app.state.Get([]byte(KeyInvoice(req.GetCheckInvoice().IsInvoiceId)))
	if !exists {
		return nil, ErrEmptyInvoiceHeaderId
	}
	var invoice InvoiceHeader
	err := UnmarshalMessage(value, &invoice)
	if err != nil {
		return nil, ErrStorage
	}
	if req.GetCheckInvoice().IsPass {
		invoice.InvoiceState = InvoiceState_IS_CHECKED
	} else {
		invoice.InvoiceState = InvoiceState_IS_REJECTED
	}

	save, err := MarshalMessage(&invoice)
	if err != nil {
		return nil, ErrStorage
	}

	app.state.Set([]byte(KeyInvoice(req.GetCheckInvoice().IsInvoiceId)), save)

	event := &EventCheckInvoice{}
	instructionId := req.GetInstructionId()
	event.InstructionId = instructionId
	return &Response{Value: &Response_CheckInvoice{&ResponseCheckInvoice{InstructionId: instructionId, Event: &Event{Value: &Event_CheckInvoice{event}}}}}, nil
}

func (app *MideaSupplyApplication) checkChangePubkey(req *Request) error {
	if req.GetChangePubkey() == nil {
		return ErrEmptyValue
	}
	uid, err := strconv.ParseInt(req.GetUid(), 10, 64)
	_, value, exists := app.state.Get([]byte(KeySupplier(uid)))
	if !exists {
		return ErrSupplierNotRegister
	}
	var supplier Supplier
	err = UnmarshalMessage(value, &supplier)
	if err != nil {
		return ErrStorage
	}
	if !bytes.Equal(req.GetPubkey(), supplier.UserPubkey) {
		return ErrNoRight
	}
	if len(req.GetChangePubkey().NewPubkey) != 32 {
		return ErrWrongPubkey
	}
	return nil
}

func (app *MideaSupplyApplication) changePubkey(req *Request) (*Response, error) {
	err1 := app.checkChangePubkey(req)
	if err1 != nil {
		return nil, err1
	}
	uid, err := strconv.ParseInt(req.GetUid(), 10, 64)
	_, value, exists := app.state.Get([]byte(KeySupplier(uid)))
	if !exists {
		return nil, ErrSupplierNotRegister
	}
	var supplier Supplier
	err = UnmarshalMessage(value, &supplier)
	if err != nil {
		return nil, ErrStorage
	}
	supplier.UserPubkey = req.GetChangePubkey().NewPubkey

	save, err := MarshalMessage(&supplier)
	if err != nil {
		return nil, ErrStorage
	}
	app.state.Set([]byte(KeySupplier(uid)), save)
	instructionId := req.GetInstructionId()
	event := &EventChangePubkey{}
	event.InstructionId = instructionId
	return &Response{Value: &Response_ChangePubkey{&ResponseChangePubkey{InstructionId: instructionId, Event: &Event{Value: &Event_ChangePubkey{event}}}}}, nil
}

func (app *MideaSupplyApplication) checkResetPubkey(req *Request) error {
	if err := app.isPlatformAdmin(req.GetPubkey()); err != nil {
		return err
	}
	if req.GetResetPubkey() == nil {
		return ErrEmptyValue
	}
	_, _, exists := app.state.Get([]byte(KeySupplier(req.GetResetPubkey().VendorId)))
	if !exists {
		return ErrSupplierNotRegister
	}
	if len(req.GetResetPubkey().NewPubkey) != 32 {
		return ErrWrongPubkey
	}
	return nil
}

func (app *MideaSupplyApplication) resetPubkey(req *Request) (*Response, error) {
	err1 := app.checkResetPubkey(req)
	if err1 != nil {
		return nil, err1
	}
	_, value, exists := app.state.Get([]byte(KeySupplier(req.GetResetPubkey().VendorId)))
	if !exists {
		return nil, ErrSupplierNotRegister
	}
	var supplier Supplier
	err := UnmarshalMessage(value, &supplier)
	if err != nil {
		return nil, ErrStorage
	}
	supplier.UserPubkey = req.GetResetPubkey().NewPubkey

	save, err := MarshalMessage(&supplier)
	if err != nil {
		return nil, ErrStorage
	}
	app.state.Set([]byte(KeySupplier(req.GetResetPubkey().VendorId)), save)
	instructionId := req.GetInstructionId()
	event := &EventResetPubkey{}
	event.InstructionId = instructionId
	return &Response{Value: &Response_ResetPubkey{&ResponseResetPubkey{InstructionId: instructionId, Event: &Event{Value: &Event_ResetPubkey{event}}}}}, nil
}

func (app *MideaSupplyApplication) checkWarehouseEntryList(req *Request) error {
	if err := app.isPlatformAdmin(req.GetPubkey()); err != nil {
		return err
	}
	// if len(req.GetWarehouseEntryList().ReceiveHeaderList) > 1000 {
	// 	return ErrOverMaxLimit
	// }
	list := make(map[int64]int64, 0)
	for _, v := range req.GetWarehouseEntryList().ReceiveHeaderList {
		_, ok := list[v.RecHeaderId]
		if ok {
			return ErrRecDupHeaderId
		}

		_, _, exists := app.state.Get([]byte(KeyEnrty(v.RecHeaderId)))
		if exists {
			return ErrRecDupHeaderId
		}

		list[v.RecHeaderId] = v.RecHeaderId
		mlist := make(map[int64]int64, 0)
		for _, vv := range v.ReceiveLines {
			if vv.ShipmentHeaderId <= int64(0) {
				return ErrWrongHeaderId
			}
			if vv.RcvTranId <= int64(0) {
				return ErrEmptyTranId
			}
			_, _, exists := app.state.Get([]byte(KeyTranId(vv.RcvTranId)))
			if exists {
				return ErrTranIdExists
			}
			_, ok := mlist[vv.RcvTranId]
			if ok {
				return ErrDupTranId
			}
			mlist[vv.RcvTranId] = vv.RcvTranId
		}
	}

	if req.GetWarehouseEntryList() == nil {
		return ErrEmptyValue
	}

	return nil
}

func (app *MideaSupplyApplication) warehouseEntryList(req *Request) (*Response, error) {
	err1 := app.checkWarehouseEntryList(req)
	if err1 != nil {
		return nil, err1
	}
	tempLine := make([]int64, 0)
	for _, h := range req.GetWarehouseEntryList().ReceiveHeaderList {
		tempLine = append(tempLine, h.RecHeaderId)
		var linenum, lineamount int64
		receiveHeder := &ReceiveHeader{}
		receiveHeder.RecHeaderId = h.RecHeaderId
		receiveHeder.ShipmentHeaderId = h.ShipmentHeaderId
		receiveHeder.ReceiptNum = h.ReceiptNum
		receiveHeder.ReceiveDate = h.ReceiveDate
		receiveHeder.ReceivePerson = h.ReceivePerson
		receiveHeder.PurchansePerson = h.PurchansePerson
		receiveHeder.OrganizationId = h.OrganizationId
		receiveHeder.OrganizationCode = h.OrganizationCode
		receiveHeder.VendorId = h.VendorId
		receiveHeder.VendorCode = h.VendorCode
		receiveHeder.VendorName = h.VendorName
		receiveHeder.VendorSiteId = h.VendorSiteId
		receiveHeder.VendorSiteCode = h.VendorSiteCode
		receiveHeder.AttributeCategory = h.AttributeCategory
		receiveHeder.Attribute1 = h.Attribute1
		receiveHeder.Attribute2 = h.Attribute2
		receiveHeder.Attribute3 = h.Attribute3
		receiveHeder.Attribute4 = h.Attribute4
		receiveHeder.Attribute5 = h.Attribute5
		receiveHeder.Attribute6 = h.Attribute6
		receiveHeder.Attribute7 = h.Attribute7
		receiveHeder.Attribute8 = h.Attribute8
		receiveHeder.Attribute9 = h.Attribute9
		receiveHeder.Attribute10 = h.Attribute10
		receiveHeder.Attribute11 = h.Attribute11
		receiveHeder.Attribute12 = h.Attribute12
		receiveHeder.Attribute13 = h.Attribute13
		receiveHeder.Attribute14 = h.Attribute14
		receiveHeder.Attribute15 = h.Attribute15
		receiveHeder.ReceiveLines = h.ReceiveLines

		_, value, exists := app.state.Get([]byte(KeySupplier(h.VendorId)))
		if !exists {
			return nil, ErrSupplierNotRegister
		}
		var supplier Supplier
		err := UnmarshalMessage(value, &supplier)
		if err != nil {
			return nil, ErrStorage
		}
		check := false
		for _, v := range supplier.SupplierSite {
			if v.SiteId == h.VendorSiteId {
				check = true
				break
			}
		}
		if !check {
			var supplierSite SupplierSite
			//supplierSite.Site = h.VendorSite
			supplierSite.SiteCode = h.VendorSiteCode
			supplierSite.SiteId = h.VendorSiteId
			supplier.SupplierSite = append(supplier.SupplierSite, &supplierSite)
		}

		rightLines := make([]*ReceiveLine, 0)
		returnLines := make([]*ReceiveLine, 0)
		for _, v := range h.ReceiveLines {
			if v.RcvTranType == "RETURN TO VENDOR" {
				// for _, vv := range rightLines {
				// 	if v.TopTranId == vv.TopTranId {
				// 		vv.Quantity = vv.Quantity - v.ReturnQuantity
				// 		break
				// 	}
				// }
				lineamount = lineamount - (v.Quantity*v.PoUnitPrice)/QUANTITY_ZERO_LIMIT
				returnLines = append(returnLines, v)
			} else {
				rightLines = append(rightLines, v)
				lineamount = lineamount + (v.Quantity*v.PoUnitPrice)/QUANTITY_ZERO_LIMIT
			}

		}
		linenum = int64(len(rightLines))

		for _, v := range receiveHeder.ReceiveLines {
			receiveLine := &ReceiveLine{}
			receiveLine.RcvTranId = v.RcvTranId
			receiveLine.ShipmentHeaderId = v.ShipmentHeaderId
			receiveLine.ItemId = v.ItemId
			receiveLine.ItemCode = v.ItemCode
			//receiveLine.ItemName = v.ItemName
			receiveLine.ItemDesc = v.ItemDesc
			receiveLine.PrimaryUnit = v.PrimaryUnit
			receiveLine.PoHeaderId = v.PoHeaderId
			receiveLine.PoLineId = v.PoLineId
			//receiveLine.LineLocationId = v.LineLocationId
			receiveLine.PoNum = v.PoNum
			receiveLine.LineNum = v.LineNum
			receiveLine.CurrencyCode = v.CurrencyCode
			receiveLine.Quantity = v.Quantity
			receiveLine.PoUnitPrice = v.PoUnitPrice
			receiveLine.PriceMatched = v.PriceMatched
			receiveLine.QuantityMatched = v.QuantityMatched
			receiveLine.AmountMatched = v.AmountMatched
			receiveLine.Remark = v.Remark
			receiveLine.RcvTranType = v.RcvTranType
			receiveLine.RcvTranDate = v.RcvTranDate
			receiveLine.ReturnQuantity = v.ReturnQuantity
			receiveLine.TopTranId = v.TopTranId
			receiveLine.AttributeCategory = v.AttributeCategory
			receiveLine.Attribute1 = v.Attribute1
			receiveLine.Attribute2 = v.Attribute2
			receiveLine.Attribute3 = v.Attribute3
			receiveLine.Attribute4 = v.Attribute4
			receiveLine.Attribute5 = v.Attribute5
			receiveLine.Attribute6 = v.Attribute6
			receiveLine.Attribute7 = v.Attribute7
			receiveLine.Attribute8 = v.Attribute8
			receiveLine.Attribute9 = v.Attribute9
			receiveLine.Attribute10 = v.Attribute10
			receiveLine.Attribute11 = v.Attribute11
			receiveLine.Attribute12 = v.Attribute12
			receiveLine.Attribute13 = v.Attribute13
			receiveLine.Attribute14 = v.Attribute14
			receiveLine.Attribute15 = v.Attribute15
			save, err := MarshalMessage(receiveLine)
			if err != nil {
				return nil, ErrStorage
			}

			app.state.Set([]byte(KeyTranId(v.RcvTranId)), save)
		}

		for _, v := range rightLines {

			supplier.Rcvlines = append(supplier.Rcvlines, v.RcvTranId)

			uninvline := &UnInvoiceLine{}
			uninvline.VendorCode = h.VendorCode
			uninvline.VendorName = h.VendorName
			uninvline.VendorSiteCode = h.VendorSiteCode
			uninvline.ItemCode = v.ItemCode
			uninvline.ItemDesc = v.ItemDesc
			uninvline.CurrencyCode = v.CurrencyCode
			uninvline.PoNum = v.PoNum
			uninvline.ReceiptNum = h.ReceiptNum
			uninvline.ReceiveDate = v.RcvTranDate
			uninvline.PrimayUnit = v.PrimaryUnit
			uninvline.QuantityReceived = v.Quantity
			uninvline.PoUnitPrice = v.PoUnitPrice
			uninvline.OrganizationId = h.OrganizationId
			uninvline.TranId = v.RcvTranId
			supplier.UninvLines = append(supplier.UninvLines, uninvline)
			//app.state.Set([]byte(KeyTranId(v.RcvTranId)), []byte(strconv.FormatInt(v.RcvTranId, 10)))
		}
		for _, v := range returnLines {
			_, value, exists := app.state.Get([]byte(KeyTranId(v.RcvTranId)))
			if !exists {
				return nil, ErrEmptyTranId
			}
			var revline ReceiveLine
			err := UnmarshalMessage(value, &revline)
			if err != nil {
				return nil, ErrStorage
			}
			//revline.Quantity = revline.Quantity - v.ReturnQuantity
			revline.ReturnQuantity = v.ReturnQuantity
			save, err := MarshalMessage(&revline)
			if err != nil {
				return nil, ErrStorage
			}
			app.state.Set([]byte(KeyTranId(v.RcvTranId)), save)
		}

		supplier.Unlinenum = supplier.Unlinenum + linenum
		supplier.Unlineamount = supplier.Unlineamount + lineamount
		save, err := MarshalMessage(receiveHeder)
		if err != nil {
			return nil, ErrStorage
		}
		app.state.Set([]byte(KeyEnrty(h.RecHeaderId)), save)
		save, err = MarshalMessage(&supplier)
		if err != nil {
			return nil, ErrStorage
		}
		app.state.Set([]byte(KeyEnrty(h.VendorId)), save)
	}
	instructionId := req.GetInstructionId()
	event := &EventWarehouseEntryList{}
	event.Ids = tempLine
	return &Response{Value: &Response_WarehouseEntryList{&ResponseWarehouseEntryList{InstructionId: instructionId, Event: &Event{Value: &Event_WarehouseEntryList{event}}}}}, nil
}
