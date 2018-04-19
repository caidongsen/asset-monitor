package mideaSupply

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
var ErrStorage = errors.New("storage error")
var ErrNoRight = errors.New("no right")
var ErrNotAdmin = errors.New("not admin")
var ErrWrongPubkey = errors.New("wrong pubkey")
var ErrGuaranteeNotEnough = errors.New("guarantee not enough")
var ErrWrongState = errors.New("wrong state")
var ErrPlatformExists = errors.New("platform exists")
var ErrPlatformNotExists = errors.New("platform not exists")
var ErrEmptyValue = errors.New("empty value")
var ErrSupplierExists = errors.New("supplier exists")
var ErrWrongHeaderId = errors.New("wrong header id")
var ErrWrongRecHeaderId = errors.New("wrong recheader id")
var ErrInvoiceExists = errors.New("inovice exists")
var ErrEmptyInvoiceHeaderId = errors.New("empty invoice header id")
var ErrEmptyInvoiceLineId = errors.New("empty invoice line id")
var ErrRecDupHeaderId = errors.New("RecHeaderId exists")
var ErrInvoiceChecked = errors.New(" Invoice Checked")
var ErrEmptyTranId = errors.New("ErrEmptyTranId")
var ErrTranIdNotExists = errors.New("ErrTranIdNotExists")
var ErrTranIdExists = errors.New("ErrTranIdExists")
var ErrDupTranId = errors.New("ErrDupTranId")
var ErrOverMaxLimit = errors.New("Over Max Limit")
var ErrDupVendorId = errors.New("ErrDupVendorId")
var ErrSupplierNotRegister = errors.New("ErrSupplierNotRegister")
var ErrQuantityException = errors.New("not enough quantity to do")

const (
	QUANTITY_ZERO_LIMIT = 1000000 //数量*1000000入区块链
	RATE_ZERO_LIMIT     = 1000000
	PRICE_ZERO_LIMIT    = 1000000
	AMOUNT_ZERO_LIMIT   = 1000
)

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

func (app *MideaSupplyApplication) saveReceipt(instructionId int64, receipt *Receipt) error {
	save, err := MarshalMessage(receipt)
	if err != nil {
		return err
	}
	app.state.Set(reqkey(instructionId), save)
	return err
}

func (app *MideaSupplyApplication) checkInstructionId(instructionId int64) error {
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

func KeyPlatform() []byte {
	return []byte("platform")
}

func KeySupplier(id int64) string {
	return "supplier:" + fmt.Sprintf("%d", id)
}

func KeyEnrty(id int64) string {
	return "entry:" + fmt.Sprintf("%d", id)
}

func KeyInvoice(id int64) string {
	return "invoice:" + fmt.Sprintf("%d", id)
}

func KeyTranId(id int64) string {
	return "tranid:" + fmt.Sprintf("%d", id)
}

func KeyAdmins() string {
	return "administrators"
}

func KeyUser(id string) string {
	return "user:" + fmt.Sprintf("%s", id)
}

func isOriginalAdmin(str []byte) bool {
	_, ok := adminList[hex.EncodeToString(str)]
	return ok
}

func (app *MideaSupplyApplication) isPlatformAdmin(str []byte) error {
	_, value, exists := app.state.Get([]byte(KeyPlatform()))
	if !exists {
		return ErrPlatformNotExists
	}
	var platform Platform
	err := UnmarshalMessage(value, &platform)
	if err != nil {
		return ErrStorage
	}
	if !bytes.Equal(str, platform.Pubkey) {
		return ErrNoRight
	}
	return nil
}

/*func (app *MideaSupplyApplication) loadAdmins() *Admins {
	_, value, exists := app.state.Get([]byte(KeyAdmins()))
	if !exists {
		return nil
	}
	var admins Admins
	err := UnmarshalMessage(value, &admins)
	if err != nil {
		return nil
	}
	return &admins
}

func (app *MideaSupplyApplication) getAdminType(str []byte) AdminType {
	if isOriginalAdmin(str) {
		return AdminType_A_NORMAL
	}
	admins := app.loadAdmins()
	if admins != nil {
		for i := 0; i < len(admins.Admins); i++ {
			if bytes.Equal(admins.Admins[i].AdminAddr, str) {
				return admins.Admins[i].AdminType
			}
		}
	}
	return AdminType_A_UNK
}*/

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
