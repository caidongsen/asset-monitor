package supplychain2

import (
	//"bytes"
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
var ErrAdminExist = errors.New("admin exist")
var ErrAdminNotExist = errors.New("admin not exist")
var ErrUserExist = errors.New("user exist")
var ErrUserNotExist = errors.New("user not exist")
var ErrUserNameExist = errors.New("name exist")
var ErrUserNameNotExist = errors.New("name not exist")
var ErrLoanExist = errors.New("loan exist")
var ErrLoanNotExist = errors.New("loan not exist")
var ErrStorage = errors.New("storage error")
var ErrNoRight = errors.New("no right")
var ErrNotAdmin = errors.New("not admin")
var ErrWrongPubkey = errors.New("wrong pubkey")
var ErrCannotBuy = errors.New("can not buy")
var ErrNotCreator = errors.New("not creator")
var ErrNotGuarantee = errors.New("not guarantee")
var ErrCreditNotEnough = errors.New("credit not enough")
var ErrGuaranteeNotExist = errors.New("guarantee not exist")
var ErrGuaranteeNotEnough = errors.New("guarantee not enough")
var ErrWrongState = errors.New("wrong state")
var ErrWrongUploadId = errors.New("wrong uploadid")
var ErrRmbNotEnough = errors.New("rmb not enough")
var ErrNotBuyer = errors.New("not buyer")
var ErrSellerNotExist = errors.New("seller not exist")
var ErrWrongUserType = errors.New("wrong user type")
var ErrNotBank = errors.New("not bank")
var ErrWrongCash = errors.New("wrong cash value")
var ErrBankRmbNotEnough = errors.New("bank rmb not enough")
var ErrBankNotExist = errors.New("bank not exist")
var ErrUserCashNotEnough = errors.New("user cash not enough")
var ErrIdCardExists = errors.New("idcard exists")
var ErrPersonalInfoNotExist = errors.New("no personal info")

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

func (app *Supplychain2Application) saveReceipt(instructionId int64, receipt *Receipt) error {
	save, err := MarshalMessage(receipt)
	if err != nil {
		return err
	}
	app.state.Set(reqkey(instructionId), save)
	return err
}

func (app *Supplychain2Application) checkInstructionId(instructionId int64) error {
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

func KeyAdmins() string {
	return "administrators"
}

func KeyUser(id string) string {
	return "user:" + fmt.Sprintf("%s", id)
}

func KeyUserName(name string) string {
	return "username:" + fmt.Sprintf("%s", name)
}

func KeyUserIdCard(id string) string {
	return "idcard:" + fmt.Sprintf("%s", id)
}

func KeyEnterpriseBaseInfo(id string) string {
	return "enterprise:" + fmt.Sprintf("%s", id)
}

func KeyPersonalInfo(id string) string {
	return "personal:" + fmt.Sprintf("%s", id)
}

func KeyUploadFile(id int64) string {
	return "uploadfile:" + fmt.Sprintf("%d", id)
}

func KeyLoan(id int64) string {
	return "loan:" + fmt.Sprintf("%d", id)
}

func KeyRecord(id int64) string {
	return "record:" + fmt.Sprintf("%d", id)
}

func KeyBank(bank []byte) string {
	return "bank:" + fmt.Sprintf("%s", string(bank))
}

func isOriginalAdmin(str []byte) bool {
	_, ok := adminList[hex.EncodeToString(str)]
	return ok
}

/*func (app *Supplychain2Application) loadAdmins() *Admins {
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

func (app *Supplychain2Application) getAdminType(str []byte) AdminType {
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
