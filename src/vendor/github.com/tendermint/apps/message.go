package apps

import (
	"encoding/hex"
	"errors"
	"fmt"

	proto "github.com/golang/protobuf/proto"
	"github.com/tendermint/abci/types"
	"github.com/tendermint/ed25519"
	gocrypto "github.com/tendermint/go-crypto"
	godb "github.com/tendermint/tendermint/apps/kvdb"
)

var fakeCrypto = false

var (
	ErrNotFound      = errors.New("ErrNotFound")
	ErrPubKeyFormat  = errors.New("ErrPubKeyFormat")
	ErrTooManySender = errors.New("ErrTooManySender")
	ErrNotSender     = errors.New("ErrNotSender")
	ErrExist         = errors.New("ErrExist")
	ErrVersion       = errors.New("ErrVersion")
	ErrSignsNil      = errors.New("ErrSignsNil")
	ErrSign          = errors.New("ErrSign")
)

type Transaction struct {
	MainTx
	Hash []byte
}

type IRequest interface {
	proto.Message
	GetAccount() []byte
	GetSigns() [][]byte
	Copy() IRequest
	GetMsgType() MessageType
	SetSigns([][]byte)
	Signdatas([]*[64]byte)
}

func Signdata(priv *[64]byte, data []byte) []byte {
	if fakeCrypto {
		sign := make([]byte, 64)
		return sign[:]
	}
	sign := ed25519.Sign(priv, data)
	return sign[:]
}

func GetPub(key *[64]byte) string {
	uid := hex.EncodeToString(key[32:])
	return uid
}

func GenKey() (priv *[64]byte, err error) {
	key := gocrypto.GenPrivKeyEd25519()
	privkey := [64]byte(key)
	return &privkey, nil
}

func CheckSign(data []byte, uid []byte, sig []byte) error {
	if fakeCrypto {
		return nil
	}
	var pubkey [32]byte
	copy(pubkey[:], uid)
	var sign [64]byte
	copy(sign[:], sig)
	ok := ed25519.Verify(&pubkey, data, &sign)
	if !ok {
		return ErrSign
	}
	return nil
}

// Write proto message, length delimited
func Encode(msg proto.Message) ([]byte, error) {
	return proto.Marshal(msg)
}

// Read proto message, length delimited
func Decode(bz []byte, msg proto.Message) error {
	return proto.Unmarshal(bz, msg)
}

func (req *WriteRequest) CheckSign() error {
	return CheckReqSigns(req)
}

func (req *WriteRequest) Signdatas(key []*[64]byte) {
	SignReqdatas(req, key)
}

func (req *WriteRequest) Copy() IRequest {
	tmp := *req
	return &tmp
}

func (req *WriteRequest) SetSigns(signs [][]byte) {
	req.Signs = signs
}

func SignReqdatas(req IRequest, keys []*[64]byte) {
	reqcopy := req.Copy()
	reqcopy.SetSigns(nil)
	data, err := Encode(reqcopy)
	if err != nil {
		panic(err)
	}

	var signs [][]byte
	for _, k := range keys {
		s := Signdata(k, data)
		var sign [96]byte
		copy(sign[:32], k[32:])
		copy(sign[32:], s)
		signs = append(signs, sign[:])
	}
	req.SetSigns(signs)
}

func CheckReqSigns(req IRequest) error {
	reqcopy := req.Copy()
	reqcopy.SetSigns(nil)
	data, err := Encode(reqcopy)
	if err != nil {
		return err
	}

	signs := req.GetSigns()
	num := len(signs)

	if num == 0 {
		return ErrSignsNil
	}

	for i := 0; i < num; i++ {
		err := verifys(data, signs[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (req *WriteRequest) GetMsgType() MessageType {
	return req.ActionId
}

func (req *ReadRequest) GetMsgType() MessageType { // value
	return req.ActionId
}

func (req *ReadRequest) Copy() IRequest { // value
	tmp := *req
	return &tmp
}

func (req *ReadRequest) SetSigns(signs [][]byte) {
	req.Signs = signs
}

func (req *ReadRequest) Signdatas(key []*[64]byte) {
	SignReqdatas(req, key)
}

func (req *ReadRequest) CheckSign() error { // value
	return CheckReqSigns(req)
}

func HashSha256(data []byte) []byte {
	return gocrypto.Sha256(data)
}

func ParseTx(tx []byte) (*Transaction, error) {
	var maintx MainTx
	err := Decode(tx, &maintx)
	if err != nil {
		return nil, err
	}
	hash := HashSha256(maintx.Data)
	return &Transaction{maintx, hash}, nil
}

func KeyAccount(uid []byte, coinId int32) string {
	key := "apps-acc:" + hex.EncodeToString(uid) + ":" + pad32(coinId)
	return key
}

func LoadAccount(ldb godb.DB, uid []byte, coinId int32) (*Account, error) {
	value := ldb.Get([]byte(KeyAccount(uid, coinId)))
	if value == nil {
		return nil, ErrNotFound
	}
	var acc Account
	err := Decode(value, &acc)
	if err != nil {
		panic(err)
	}
	return &acc, nil
}

func SaveAccounts(ldb godb.DB, accounts []*Account) error {
	batch := ldb.NewBatch(true)
	for i := 0; i < len(accounts); i++ {
		key := []byte(KeyAccount(accounts[i].Account, accounts[i].CoinId))
		value, err := Encode(accounts[i])
		if err != nil {
			panic(err)
		}
		err = batch.Put(key, value)
		if err != nil {
			return err
		}
	}
	return batch.Commit()
}

func pad32(i int32) string {
	return fmt.Sprintf("%010d", i)
}

func pad64(i int64) string {
	return fmt.Sprintf("%020d", i)
}

func KeyTx(hash []byte) string {
	return "apps-tx:" + string(hash)
}

func KeyConf(appId int32) string {
	return "apps-conf:" + pad32(appId)
}

func KeyReceipt(appVersion int32, hash []byte) string {
	return "apps-receipt:" + pad32(appVersion) + ":" + string(hash)
}

func NewError(code types.CodeType, err string) (resQuery types.ResponseQuery) {
	resQuery.Code = code
	resQuery.Log = err
	return
}

func NewData(code types.CodeType, data []byte) (resQuery types.ResponseQuery) {
	resQuery.Code = code
	resQuery.Value = data
	return
}

func verifys(data []byte, sign []byte) error {
	if len(sign) != 96 {
		return errors.New("errSignFormat")
	}

	_pubkey := sign[0:32]
	_sign := sign[32:]
	return CheckSign(data, _pubkey, _sign)
}

func KeySubAccount(account []byte, subaccount []byte, coinId int32) string {
	key := "apps-subacc:" + hex.EncodeToString(account) + ":" + hex.EncodeToString(subaccount) + ":" + pad32(coinId)
	return key
}

func LoadSubAccount(ldb godb.DB, account []byte, subaccount []byte, coinId int32) (*SubAccount, error) {
	key := KeySubAccount(account, subaccount, coinId)

	value := ldb.Get([]byte(key))
	if value == nil {
		return nil, ErrNotFound
	}

	var acc SubAccount
	err := Decode(value, &acc)
	if err != nil {
		panic(err)
	}

	return &acc, nil
}

func SaveSubAccount(ldb godb.DB, a *SubAccount) error {
	batch := ldb.NewBatch(true)
	key := []byte(KeySubAccount(a.GetAccount(), a.GetSubaccount(), a.GetCoinId()))
	value, err := Encode(a)
	if err != nil {
		panic(err)
	}

	err = batch.Put(key, value)
	if err != nil {
		return err
	}

	return batch.Commit()
}

func SaveAccount(ldb godb.DB, a *Account) error {
	batch := ldb.NewBatch(true)
	key := []byte(KeyAccount(a.GetAccount(), a.GetCoinId()))
	value, err := Encode(a)
	if err != nil {
		panic(err)
	}

	err = batch.Put(key, value)
	if err != nil {
		return err
	}

	return batch.Commit()
}
