package smsc

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"

	"dev.33.cn/33/crypto"
	_ "dev.33.cn/33/crypto/ed25519"
)

var ErrDupInstructionId = errors.New("ErrDupInstructionId")
var ErrSign = errors.New("error sign")
var ErrWrongMessageType = errors.New("wrong message type")
var ErrSignWrongLength = errors.New("wrong length")
var ErrUidTooShort = errors.New("ErrUidTooShort")
var ErrStorage = errors.New("storage error")
var ErrNoRight = errors.New("no right")
var ErrNotAdmin = errors.New("not admin")
var ErrWrongPubkey = errors.New("wrong pubkey")
var ErrInvalidRole = errors.New("invalid role")
var ErrUserExist = errors.New("user exist")
var ErrUserNotExist = errors.New("user not exist")
var ErrOrderNotExist = errors.New("order not exist")
var ErrWrongSupplier = errors.New("wrong supplier id")
var ErrWrongCarrier = errors.New("wrong carrier id")

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

func (app *SmscApplication) saveReceipt(instructionId int64, receipt *Receipt) error {
	save, err := MarshalMessage(receipt)
	if err != nil {
		return err
	}
	app.state.Set(reqkey(instructionId), save)
	return err
}

func (app *SmscApplication) checkInstructionId(instructionId int64) error {
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

func KeyPlanner(id int64) string {
	return "planner:" + fmt.Sprintf("%d", id)
}

func KeySupplier(id int64) string {
	return "supplier:" + fmt.Sprintf("%d", id)
}

func KeyCarrier(id int64) string {
	return "carrier:" + fmt.Sprintf("%d", id)
}

func KeyChecker(id int64) string {
	return "checker:" + fmt.Sprintf("%d", id)
}

func KeyOrder(id string) string {
	return "order:" + fmt.Sprintf("%s", id)
}

func KeyAdmin() string {
	return "admin"
}

func isOriginalAdmin(str []byte) bool {
	_, ok := adminList[hex.EncodeToString(str)]
	if ok {
		return ok
	}
	return false
}

func (app *SmscApplication) isAdmin(pubkey []byte) bool {
	if isOriginalAdmin(pubkey) {
		return true
	}
	_, value, exists := app.state.Get([]byte(KeyAdmin()))
	if !exists {
		return false
	}
	var admin Admin
	err := UnmarshalMessage(value, &admin)
	if err != nil {
		return false
	}
	if !bytes.Equal(admin.Pubkey, pubkey) {
		return false
	}
	return true
}

func checkRole(role Role) error {
	if role != Role_RPlanner && role != Role_RSupplier && role != Role_RCarrier && role != Role_RChecker {
		return ErrInvalidRole
	}
	return nil
}

func (app *SmscApplication) getUser(id int64, role Role) (interface{}, error) {
	if err := checkRole(role); err != nil {
		return nil, err
	}
	switch role {
	case Role_RPlanner:
		_, value, exists := app.state.Get([]byte(KeyPlanner(id)))
		if !exists {
			return nil, ErrUserNotExist
		}
		var planner Planner
		err := UnmarshalMessage(value, &planner)
		if err != nil {
			return nil, ErrStorage
		}
		return planner, nil
	case Role_RSupplier:
		_, value, exists := app.state.Get([]byte(KeySupplier(id)))
		if !exists {
			return nil, ErrUserNotExist
		}
		var supplier Supplier
		err := UnmarshalMessage(value, &supplier)
		if err != nil {
			return nil, ErrStorage
		}
		return supplier, nil
	case Role_RCarrier:
		_, value, exists := app.state.Get([]byte(KeyCarrier(id)))
		if !exists {
			return nil, ErrUserNotExist
		}
		var carrier Carrier
		err := UnmarshalMessage(value, &carrier)
		if err != nil {
			return nil, ErrStorage
		}
		return carrier, nil
	case Role_RChecker:
		_, value, exists := app.state.Get([]byte(KeyChecker(id)))
		if !exists {
			return nil, ErrUserNotExist
		}
		var checker Checker
		err := UnmarshalMessage(value, &checker)
		if err != nil {
			return nil, ErrStorage
		}
		return checker, nil
	}
	return nil, ErrInvalidRole
}

func (app *SmscApplication) getOrder(id string) (Order, error) {
	var order Order
	_, value, exists := app.state.Get([]byte(KeyOrder(id)))
	if !exists {
		return order, ErrOrderNotExist
	}
	err := UnmarshalMessage(value, &order)
	if err != nil {
		return order, ErrStorage
	}
	return order, nil
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
