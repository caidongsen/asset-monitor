package mideaBill_test

import (
	"testing"

	"time"
	"math/rand"
	"encoding/hex"
	"encoding/base64"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/tendermint/apps/mideaBill"

	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	rpc "github.com/tendermint/tendermint/rpc/lib/client"

	//"bytes"
	//"io/ioutil"
	//"sort"

	"github.com/stretchr/testify/require"
	//abcicli "github.com/tendermint/abci/client"
	//"github.com/tendermint/abci/server"
	//"github.com/tendermint/abci/types"
	//crypto "github.com/tendermint/go-crypto"
	//"github.com/tendermint/merkleeyes/iavl"
	//cmn "github.com/tendermint/tmlibs/common"
	//"github.com/tendermint/tmlibs/log"
)

var (
	sdkKey = ""
)

func genInitPlatform(t *testing.T) []byte {
	sdkPriv := GetPrivateKeyByPwd("sdkuser", "123456")
	sdkKey = hex.EncodeToString(sdkPriv)
t.Log("sdk pubkic key")
t.Log(hex.EncodeToString(sdkPriv[32:]))
t.Log("sdk private key")
t.Log(hex.EncodeToString(sdkPriv[:32]))

	req := &mideaBill.RequestInitPlatform{}
	req.UserName = "admin"
	req.UserPublicKey= GetPrivateKeyByPwd("admin", "123456")[32:]
	req.EntCode = "35992041-X"
	req.EntName = "美的商业保理有限公司"
	req.SdkUserName = "sdkuser"
	req.SdkPublicKey = sdkPriv[32:]

	request := &mideaBill.Request{}
	request.Value = &mideaBill.Request_InitPlatform{req}
	request.UserName = mideaBill.SdkUser
	request.InstructionId = time.Now().Unix()*1000 + int64(rand.Intn(1000))
	request.Pubkey = sdkPriv[32:]
	request.ActionId = mideaBill.MessageType_MsgInitPlatform
	request.OperatorTime = time.Now().Unix()
	data, err := mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	request.Sign = mideaBill.Signdata(sdkPriv, data)

	data, err = mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
        t.Log(base64.StdEncoding.EncodeToString(data))
	return data
}

func TestInitPlatform(t *testing.T) {
	data := genInitPlatform(t)
	testBroadcastTxCommit(t, data, clientJSON)

	data = genInitPlatform(t)
	testBroadcastTxCommitErr(t, data, clientURI)
	//send data2 to tendermint
}

func genRegisterUser(t *testing.T, userName string, entCode string) []byte {
	priv := GetPrivateKey(userName)

	req := &mideaBill.RequestRegisterUser{}
	req.EntCode = entCode
	request := &mideaBill.Request{}
	request.Value = &mideaBill.Request_RegisterUser{req}
	request.UserName = userName
	request.InstructionId = time.Now().Unix()*1000 + int64(rand.Intn(1000))
	request.Pubkey = priv[32:]
	request.ActionId = mideaBill.MessageType_MsgRegisterUser
	request.OperatorTime = time.Now().Unix()
	data, err := mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	request.Sign = mideaBill.Signdata(priv, data)

	data, err = mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestRegisterUser(t *testing.T) {
	data := genRegisterUser(t, "userName", "entCode")
	testBroadcastTxCommit(t, data, clientJSON)

	data = genRegisterUser(t, "userName", "entCode")
	testBroadcastTxCommitErr(t, data, clientURI)
	//send data2 to tendermint
}

func genAddUser(t *testing.T, userName string) []byte {
	priv := GetPrivateKey(userName)
	operPriv, err := hex.DecodeString(sdkKey)
	if err != nil {
		t.Fatal(err)
	}

	req := &mideaBill.RequestAddUser{}
	req.UserName = userName
	req.UserPublicKey = priv[32:]
	req.Operator = "userName"
	request := &mideaBill.Request{}
	request.Value = &mideaBill.Request_AddUser{req}
	request.UserName = mideaBill.SdkUser
	request.InstructionId = time.Now().Unix()*1000 + int64(rand.Intn(1000))
	request.Pubkey = operPriv[32:]
	request.ActionId = mideaBill.MessageType_MsgAddUser
	request.OperatorTime = time.Now().Unix()
	data, err := mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	request.Sign = mideaBill.Signdata(operPriv, data)

	data, err = mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestAddUser(t *testing.T) {
	data := genAddUser(t, "addUserName")
	testBroadcastTxCommit(t, data, clientJSON)

	data = genAddUser(t, "addUserName")
	testBroadcastTxCommitErr(t, data, clientURI)
	//send data2 to tendermint
}

func genUserPwdModify(t *testing.T, userName string) []byte {
	oldPriv := GetPrivateKey(userName)
	newPriv := GetPrivateKey("newPwd")

	req := &mideaBill.RequestUserPwdModify{}
	req.UserPublicKey = newPriv[32:]
	request := &mideaBill.Request{}
	request.Value = &mideaBill.Request_UserPwdModify{req}
	request.UserName = userName
	request.InstructionId = time.Now().Unix()*1000 + int64(rand.Intn(1000))
	request.Pubkey = oldPriv[32:]
	request.ActionId = mideaBill.MessageType_MsgUserPwdModify
	request.OperatorTime = time.Now().Unix()
	data, err := mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	request.Sign = mideaBill.Signdata(oldPriv, data)

	data, err = mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestUserPwdModify(t *testing.T) {
	data := genUserPwdModify(t, "addUserName")
	testBroadcastTxCommit(t, data, clientJSON)

	data = genUserPwdModify(t, "addUserName")
	testBroadcastTxCommitErr(t, data, clientURI)
	//send data2 to tendermint
}

func genUserPwdReset(t *testing.T, userName string) []byte {
	newPriv := GetPrivateKey(userName)
	priv, err := hex.DecodeString(sdkKey)
	if err != nil {
		t.Fatal(err)
	}

	req := &mideaBill.RequestUserPwdReset{}
	req.UserName = userName
	req.UserPublicKey = newPriv[32:]
	request := &mideaBill.Request{}
	request.Value = &mideaBill.Request_UserPwdReset{req}
	request.UserName = mideaBill.SdkUser
	request.InstructionId = time.Now().Unix()*1000 + int64(rand.Intn(1000))
	request.Pubkey = priv[32:]
	request.ActionId = mideaBill.MessageType_MsgUserPwdReset
	request.OperatorTime = time.Now().Unix()
	data, err := mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	request.Sign = mideaBill.Signdata(priv, data)

	data, err = mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestUserPwdReset(t *testing.T) {
	data := genUserPwdReset(t, "addUserName")
	testBroadcastTxCommit(t, data, clientJSON)

	//data = genUserPwdReset(t, "addUserName")
	//testBroadcastTxCommitErr(t, data, clientURI)
	//send data2 to tendermint
}

func genEntIdentifyCheck(t *testing.T, userName string, entCode, entName string) []byte {
	priv, err := hex.DecodeString(sdkKey)
	if err != nil {
		t.Fatal(err)
	}

	req := &mideaBill.RequestEntIdentifyCheck{}
	req.UserName = userName
	req.EntCode = entCode
	req.EntName = entName
	request := &mideaBill.Request{}
	request.Value = &mideaBill.Request_EntIdentifyCheck{req}
	request.UserName = mideaBill.SdkUser
	request.InstructionId = time.Now().Unix()*1000 + int64(rand.Intn(1000))
	request.Pubkey = priv[32:]
	request.ActionId = mideaBill.MessageType_MsgEntIdentifyCheck
	request.OperatorTime = time.Now().Unix()
	data, err := mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	request.Sign = mideaBill.Signdata(priv, data)

	data, err = mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestEntIdentifyCheck(t *testing.T) {
	data := genEntIdentifyCheck(t, "userName", "entCode", "entName")
	testBroadcastTxCommit(t, data, clientJSON)

	data = genEntIdentifyCheck(t, "userName", "entCode", "entName")
	testBroadcastTxCommitErr(t, data, clientURI)
}

func genApplyBill(t *testing.T, userName string, recvEntCode, recvEntName, mideaDraftId string) []byte {
	priv := GetPrivateKey(userName)

	req := &mideaBill.RequestApplyBill{}
	req.MideaDraftId = mideaDraftId
	req.MideaDraftAmount = 10000
	req.IssueBillDay = time.Now().Format("2006-01-02")
	req.ExpireDay = time.Now().Format("2006-01-02")
	req.PayNum = "1234"
	req.RecvBillEntCode = recvEntCode
	req.RecvBillEntName = recvEntName
	request := &mideaBill.Request{}
	request.Value = &mideaBill.Request_ApplyBill{req}
	request.UserName = userName
	request.InstructionId = time.Now().Unix()*1000 + int64(rand.Intn(1000))
	request.Pubkey = priv[32:]
	request.ActionId = mideaBill.MessageType_MsgApplyBill
	request.OperatorTime = time.Now().Unix()
	data, err := mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	request.Sign = mideaBill.Signdata(priv, data)

	data, err = mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestApplyBill(t *testing.T) {
	data := genRegisterUser(t, "recvUserName", "recvEntCode")
	testBroadcastTxCommit(t, data, clientJSON)

	data = genEntIdentifyCheck(t, "recvUserName", "recvEntCode", "recvEntName")
	testBroadcastTxCommit(t, data, clientJSON)

	data = genApplyBill(t, "userName", "recvEntCode", "recvEntName", "123456789")
	testBroadcastTxCommit(t, data, clientJSON)

	data = genApplyBill(t, "userName", "recvEntCode", "recvEntName", "123456789")
	testBroadcastTxCommitErr(t, data, clientURI)
	//send data2 to tendermint
}


func genApplyBillSign(t *testing.T, userName string, mideaDraftId string) []byte {
	priv := GetPrivateKey(userName)

	req := &mideaBill.RequestApplyBillSign{}
	req.MideaDraftId = mideaDraftId
	request := &mideaBill.Request{}
	request.Value = &mideaBill.Request_ApplyBillSign{req}
	request.UserName = userName
	request.InstructionId = time.Now().Unix()*1000 + int64(rand.Intn(1000))
	request.Pubkey = priv[32:]
	request.ActionId = mideaBill.MessageType_MsgApplyBillSign
	request.OperatorTime = time.Now().Unix()
	data, err := mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	request.Sign = mideaBill.Signdata(priv, data)

	data, err = mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestApplyBillSign(t *testing.T) {
	data := genApplyBillSign(t, "recvUserName", "123456789")
	testBroadcastTxCommit(t, data, clientJSON)

	data = genApplyBillSign(t, "recvUserName", "123456789")
	testBroadcastTxCommitErr(t, data, clientURI)
	//send data2 to tendermint
}

func genApplyBillSignRefuse(t *testing.T, userName string, mideaDraftId string) []byte {
	priv := GetPrivateKey(userName)

	req := &mideaBill.RequestApplyBillSignRefuse{}
	req.MideaDraftId = mideaDraftId
	request := &mideaBill.Request{}
	request.Value = &mideaBill.Request_ApplyBillSignRefuse{req}
	request.UserName = userName
	request.InstructionId = time.Now().Unix()*1000 + int64(rand.Intn(1000))
	request.Pubkey = priv[32:]
	request.ActionId = mideaBill.MessageType_MsgApplyBillSignRefuse
	request.OperatorTime = time.Now().Unix()
	data, err := mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	request.Sign = mideaBill.Signdata(priv, data)

	data, err = mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestApplyBillSignRefuse(t *testing.T) {
	data := genApplyBill(t, "userName", "recvEntCode", "recvEntName", "987654321")
	testBroadcastTxCommit(t, data, clientJSON)

	data = genApplyBillSignRefuse(t, "recvUserName", "987654321")
	testBroadcastTxCommit(t, data, clientJSON)

	data = genApplyBillSignRefuse(t, "recvUserName", "987654321")
	testBroadcastTxCommitErr(t, data, clientURI)
	//send data2 to tendermint
}


func genApplyBillSignCancle(t *testing.T, userName string, mideaDraftId string) []byte {
	priv := GetPrivateKey(userName)

	req := &mideaBill.RequestApplyBillSignCancle{}
	req.MideaDraftId = mideaDraftId
	request := &mideaBill.Request{}
	request.Value = &mideaBill.Request_ApplyBillSignCancle{req}
	request.UserName = userName
	request.InstructionId = time.Now().Unix()*1000 + int64(rand.Intn(1000))
	request.Pubkey = priv[32:]
	request.ActionId = mideaBill.MessageType_MsgApplyBillSignCancle
	request.OperatorTime = time.Now().Unix()
	data, err := mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	request.Sign = mideaBill.Signdata(priv, data)

	data, err = mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestApplyBillSignCancle(t *testing.T) {
	data := genApplyBill(t, "userName", "recvEntCode", "recvEntName", "11111111111")
	testBroadcastTxCommit(t, data, clientJSON)

	data = genApplyBillSignCancle(t, "userName", "11111111111")
	testBroadcastTxCommit(t, data, clientJSON)

	data = genApplyBillSignCancle(t, "userName", "11111111111")
	testBroadcastTxCommitErr(t, data, clientURI)
	//send data2 to tendermint
}


func genBillTotalTransfer(t *testing.T, userName string, mideaDraftId, entCode string) []byte {
	priv := GetPrivateKey(userName)

	req := &mideaBill.RequestBillTotalTransfer{}
	req.MideaDraftId = mideaDraftId
	req.WaitRecvBillEntCode = entCode
	request := &mideaBill.Request{}
	request.Value = &mideaBill.Request_BillTotalTransfer{req}
	request.UserName = userName
	request.InstructionId = time.Now().Unix()*1000 + int64(rand.Intn(1000))
	request.Pubkey = priv[32:]
	request.ActionId = mideaBill.MessageType_MsgBillTotalTransfer
	request.OperatorTime = time.Now().Unix()
	data, err := mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	request.Sign = mideaBill.Signdata(priv, data)

	data, err = mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestBillTotalTransfer(t *testing.T) {
	data := genBillTotalTransfer(t, "recvUserName", "123456789", "entCode")
	testBroadcastTxCommit(t, data, clientJSON)

	data = genBillTotalTransfer(t, "recvUserName", "123456789", "entCode")
	testBroadcastTxCommitErr(t, data, clientURI)
	//send data2 to tendermint
}

func genBillTransferCancle(t *testing.T, userName string, mideaDraftId string) []byte {
	priv := GetPrivateKey(userName)

	req := &mideaBill.RequestBillTransferCancle{}
	req.MideaDraftId = mideaDraftId
	request := &mideaBill.Request{}
	request.Value = &mideaBill.Request_BillTransferCancle{req}
	request.UserName = userName
	request.InstructionId = time.Now().Unix()*1000 + int64(rand.Intn(1000))
	request.Pubkey = priv[32:]
	request.ActionId = mideaBill.MessageType_MsgBillTransferCancle
	request.OperatorTime = time.Now().Unix()
	data, err := mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	request.Sign = mideaBill.Signdata(priv, data)

	data, err = mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestBillTransferCancle(t *testing.T) {
	data := genBillTransferCancle(t, "recvUserName", "123456789")
	testBroadcastTxCommit(t, data, clientJSON)

	data = genBillTransferCancle(t, "recvUserName", "123456789")
	testBroadcastTxCommitErr(t, data, clientURI)
	//send data2 to tendermint
}

func genBillTransferRefuse(t *testing.T, userName string, mideaDraftId string) []byte {
	priv := GetPrivateKey(userName)

	req := &mideaBill.RequestBillTransferRefuse{}
	req.MideaDraftId = mideaDraftId
	request := &mideaBill.Request{}
	request.Value = &mideaBill.Request_BillTransferRefuse{req}
	request.UserName = userName
	request.InstructionId = time.Now().Unix()*1000 + int64(rand.Intn(1000))
	request.Pubkey = priv[32:]
	request.ActionId = mideaBill.MessageType_MsgBillTransferRefuse
	request.OperatorTime = time.Now().Unix()
	data, err := mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	request.Sign = mideaBill.Signdata(priv, data)

	data, err = mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestBillTransferRefuse(t *testing.T) {
	data := genBillTotalTransfer(t, "recvUserName", "123456789", "entCode")
	testBroadcastTxCommit(t, data, clientJSON)

	data = genBillTransferRefuse(t, "userName", "123456789")
	testBroadcastTxCommit(t, data, clientJSON)

	data = genBillTransferRefuse(t, "userName", "123456789")
	testBroadcastTxCommitErr(t, data, clientURI)
	//send data2 to tendermint
}

func genBillTransferSign(t *testing.T, userName string, mideaDraftId string) []byte {
	priv := GetPrivateKey(userName)

	req := &mideaBill.RequestBillTransferSign{}
	req.MideaDraftId = mideaDraftId
	request := &mideaBill.Request{}
	request.Value = &mideaBill.Request_BillTransferSign{req}
	request.UserName = userName
	request.InstructionId = time.Now().Unix()*1000 + int64(rand.Intn(1000))
	request.Pubkey = priv[32:]
	request.ActionId = mideaBill.MessageType_MsgBillTransferSign
	request.OperatorTime = time.Now().Unix()
	data, err := mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	request.Sign = mideaBill.Signdata(priv, data)

	data, err = mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestBillTransferSign(t *testing.T) {
	data := genBillTotalTransfer(t, "recvUserName", "123456789", "entCode")
	testBroadcastTxCommit(t, data, clientJSON)

	data = genBillTransferSign(t, "userName", "123456789")
	testBroadcastTxCommit(t, data, clientJSON)

	data = genBillTransferSign(t, "userName", "123456789")
	testBroadcastTxCommitErr(t, data, clientURI)
	//send data2 to tendermint
}

func genBillPartTransfer(t *testing.T, userName string, mideaDraftId string) []byte {
	priv := GetPrivateKey(userName)

	req := &mideaBill.RequestBillPartTransfer{}
	req.MideaDraftId = mideaDraftId
	request := &mideaBill.Request{}
	request.Value = &mideaBill.Request_BillPartTransfer{req}
	request.UserName = userName
	request.InstructionId = time.Now().Unix()*1000 + int64(rand.Intn(1000))
	request.Pubkey = priv[32:]
	request.ActionId = mideaBill.MessageType_MsgBillPartTransfer
	request.OperatorTime = time.Now().Unix()
	data, err := mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	request.Sign = mideaBill.Signdata(priv, data)

	data, err = mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

/*
TODO:
func TestBillPartTransfer(t *testing.T) {
	data := genBillPartTransfer(t, "userName", "123456789")
	testBroadcastTxCommit(t, data, clientJSON)

	data = genBillPartTransfer(t, "userName", "123456789")
	testBroadcastTxCommitErr(t, data, clientURI)
	//send data2 to tendermint
}
*/

//TODO: BillTransferForcePay

func genBillTotalFinancing(t *testing.T, userName string, mideaDraftId, entCode string) []byte {
	priv := GetPrivateKey(userName)

	req := &mideaBill.RequestBillTotalFinancing{}
	req.MideaDraftId = mideaDraftId
	req.WaitRecvBillEntCode = entCode
	request := &mideaBill.Request{}
	request.Value = &mideaBill.Request_BillTotalFinancing{req}
	request.UserName = userName
	request.InstructionId = time.Now().Unix()*1000 + int64(rand.Intn(1000))
	request.Pubkey = priv[32:]
	request.ActionId = mideaBill.MessageType_MsgBillTotalFinancing
	request.OperatorTime = time.Now().Unix()
	data, err := mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	request.Sign = mideaBill.Signdata(priv, data)

	data, err = mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestBillTotalFinancing(t *testing.T) {
	data := genBillTotalFinancing(t, "userName", "123456789", "recvEntCode")
	testBroadcastTxCommit(t, data, clientJSON)

	data = genBillTotalFinancing(t, "userName", "123456789", "recvEntCode")
	testBroadcastTxCommitErr(t, data, clientURI)
	//send data2 to tendermint
}

func genBillFinancingCheckFail(t *testing.T, userName string, mideaDraftId string) []byte {
	priv, err := hex.DecodeString(sdkKey)
	if err != nil {
		t.Fatal(err)
	}

	req := &mideaBill.RequestBillFinancingCheckFail{}
	req.MideaDraftId = mideaDraftId
	request := &mideaBill.Request{}
	request.Value = &mideaBill.Request_BillFinancingCheckFail{req}
	request.UserName = mideaBill.SdkUser
	request.InstructionId = time.Now().Unix()*1000 + int64(rand.Intn(1000))
	request.Pubkey = priv[32:]
	request.ActionId = mideaBill.MessageType_MsgBillFinancingCheckFail
	request.OperatorTime = time.Now().Unix()
	data, err := mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	request.Sign = mideaBill.Signdata(priv, data)

	data, err = mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestBillFinancingCheckFail(t *testing.T) {
	data := genBillFinancingCheckFail(t, "userName", "123456789")
	testBroadcastTxCommit(t, data, clientJSON)

	data = genBillFinancingCheckFail(t, "userName", "123456789")
	testBroadcastTxCommitErr(t, data, clientURI)
	//send data2 to tendermint
}

func genBillFinancingCheckOk(t *testing.T, userName string, mideaDraftId string) []byte {
	priv, err := hex.DecodeString(sdkKey)
	if err != nil {
		t.Fatal(err)
	}

	req := &mideaBill.RequestBillFinancingCheckOk{}
	req.MideaDraftId = mideaDraftId
	request := &mideaBill.Request{}
	request.Value = &mideaBill.Request_BillFinancingCheckOk{req}
	request.UserName = mideaBill.SdkUser
	request.InstructionId = time.Now().Unix()*1000 + int64(rand.Intn(1000))
	request.Pubkey = priv[32:]
	request.ActionId = mideaBill.MessageType_MsgBillFinancingCheckOk
	request.OperatorTime = time.Now().Unix()
	data, err := mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	request.Sign = mideaBill.Signdata(priv, data)

	data, err = mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestBillFinancingCheckOk(t *testing.T) {
	data := genBillTotalFinancing(t, "userName", "123456789", "recvEntCode")
	testBroadcastTxCommit(t, data, clientJSON)

	data = genBillFinancingCheckOk(t, "userName", "123456789")
	testBroadcastTxCommit(t, data, clientJSON)

	data = genBillFinancingCheckOk(t, "userName", "123456789")
	testBroadcastTxCommitErr(t, data, clientURI)
	//send data2 to tendermint
}

func genBillFinancingFail(t *testing.T, userName string, mideaDraftId string) []byte {
	priv, err := hex.DecodeString(sdkKey)
	if err != nil {
		t.Fatal(err)
	}

	req := &mideaBill.RequestBillFinancingFail{}
	req.MideaDraftId = mideaDraftId
	request := &mideaBill.Request{}
	request.Value = &mideaBill.Request_BillFinancingFail{req}
	request.UserName = mideaBill.SdkUser
	request.InstructionId = time.Now().Unix()*1000 + int64(rand.Intn(1000))
	request.Pubkey = priv[32:]
	request.ActionId = mideaBill.MessageType_MsgBillFinancingFail
	request.OperatorTime = time.Now().Unix()
	data, err := mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	request.Sign = mideaBill.Signdata(priv, data)

	data, err = mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestBillFinancingFail(t *testing.T) {
	data := genBillFinancingFail(t, "recvUserName", "123456789")
	testBroadcastTxCommit(t, data, clientJSON)

	//data = genBillFinancingFail(t, "recvUserName", "123456789")
	//testBroadcastTxCommitErr(t, data, clientURI)
	//send data2 to tendermint
}

func genBillPay(t *testing.T, userName string, mideaDraftId string) []byte {
	priv, err := hex.DecodeString(sdkKey)
	if err != nil {
		t.Fatal(err)
	}

	req := &mideaBill.RequestBillPay{}
	req.MideaDraftId = mideaDraftId
	request := &mideaBill.Request{}
	request.Value = &mideaBill.Request_BillPay{req}
	request.UserName = mideaBill.SdkUser
	request.InstructionId = time.Now().Unix()*1000 + int64(rand.Intn(1000))
	request.Pubkey = priv[32:]
	request.ActionId = mideaBill.MessageType_MsgBillPay
	request.OperatorTime = time.Now().Unix()
	data, err := mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	request.Sign = mideaBill.Signdata(priv, data)

	data, err = mideaBill.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestBillPay(t *testing.T) {
	data := genBillPay(t, "userName", "123456789")
	testBroadcastTxCommit(t, data, clientJSON)

	data = genBillPay(t, "userName", "123456789")
	testBroadcastTxCommitErr(t, data, clientURI)
	//send data2 to tendermint
}


func testBroadcastTxCommitErr(t *testing.T, tx []byte, client rpc.HTTPClient) {
	require := require.New(t)

	result := new(ctypes.ResultBroadcastTxCommit)
	_, err := client.Call("broadcast_tx_commit", map[string]interface{}{"tx": tx}, result)
	require.Nil(err)

	checkTx := result.CheckTx
	require.Equal(abci.CodeType_InternalError, checkTx.Code)
	// TODO: find tx in block
}

func testBroadcastTxCommit(t *testing.T, tx []byte, client rpc.HTTPClient) {
	require := require.New(t)

	result := new(ctypes.ResultBroadcastTxCommit)
	_, err := client.Call("broadcast_tx_commit", map[string]interface{}{"tx": tx}, result)
	require.Nil(err)

	checkTx := result.CheckTx
	require.Equal(abci.CodeType_OK, checkTx.Code)
	deliverTx := result.DeliverTx
	require.Equal(abci.CodeType_OK, deliverTx.Code)
	mem := node.MempoolReactor().Mempool
	require.Equal(0, mem.Size())
	// TODO: find tx in block
}

func GetPrivateKeyByPwd(userName, pwd string) []byte {
	prv := mideaBill.GetPrivateKey(pwd)
	prv = mideaBill.GetPrivateKey(userName + "_" + hex.EncodeToString(prv))
	pub := mideaBill.GetPublicKey(prv)
	var p [64]byte
	copy(p[:32], prv[:])
	copy(p[32:], pub[:])
	return p[:]
}

func GetPrivateKey(key string) []byte {
	prv := mideaBill.GetPrivateKey(key)
	pub := mideaBill.GetPublicKey(prv)
	var p [64]byte
	copy(p[:32], prv[:])
	copy(p[32:], pub[:])
	return p[:]
}


/*
func testGwpoints(t *testing.T, app types.Application, tx []byte) {
	ar := app.DeliverTx(tx)
	require.False(t, ar.IsErr(), ar)
	// repeating tx doesn't raise error
	ar = app.DeliverTx(tx)
	require.True(t, ar.IsErr(), ar)
}

func TestGwpointsKV(t *testing.T) {
	mideaBill := mideaBill.NewGwpointsApplication()
	tx := genUserCreateTx(t, 8, 3)
	testGwpoints(t, mideaBill, tx)
}

func TestPersistentGwpointsKV(t *testing.T) {
	dir, err := ioutil.TempDir("/tmp", "abci-mideaBill-test") // TODO
	if err != nil {
		t.Fatal(err)
	}
	mideaBill := mideaBill.NewPersistentGwpointsApplication(dir)
	tx := genUserCreateTx(t, 9, 4)
	testGwpoints(t, mideaBill, tx)
}

func TestPersistentGwpointsInfo(t *testing.T) {
	dir, err := ioutil.TempDir("/tmp", "abci-mideaBill-test") // TODO
	if err != nil {
		t.Fatal(err)
	}
	mideaBill := mideaBill.NewPersistentGwpointsApplication(dir)
	height := uint64(0)

	resInfo := mideaBill.Info()
	if resInfo.LastBlockHeight != height {
		t.Fatalf("expected height of %d, got %d", height, resInfo.LastBlockHeight)
	}

	// make and apply block
	height = uint64(1)
	hash := []byte("foo")
	header := &types.Header{
		Height: uint64(height),
	}
	mideaBill.BeginBlock(hash, header)
	mideaBill.EndBlock(height)
	mideaBill.Commit()

	resInfo = mideaBill.Info()
	if resInfo.LastBlockHeight != height {
		t.Fatalf("expected height of %d, got %d", height, resInfo.LastBlockHeight)
	}

}

// add a validator, remove a validator, update a validator
func TestValSetChanges(t *testing.T) {
	dir, err := ioutil.TempDir("/tmp", "abci-mideaBill-test") // TODO
	if err != nil {
		t.Fatal(err)
	}
	app := mideaBill.NewPersistentGwpointsApplication(dir)

	// init with some validators
	total := 10
	nInit := 5
	vals := make([]*types.Validator, total)
	for i := 0; i < total; i++ {
		pubkey := crypto.GenPrivKeyEd25519FromSecret([]byte(cmn.Fmt("test%d", i))).PubKey().Bytes()
		power := cmn.RandInt()
		vals[i] = &types.Validator{pubkey, uint64(power)}
	}
	// iniitalize with the first nInit
	app.InitChain(vals[:nInit])

	vals1, vals2 := vals[:nInit], app.Validators()
	valsEqual(t, vals1, vals2)

	var v1, v2, v3 *types.Validator

	// add some validators
	v1, v2 = vals[nInit], vals[nInit+1]
	diff := []*types.Validator{v1, v2}
	tx1 := mideaBill.MakeValSetChangeTx(v1.PubKey, v1.Power)
	tx2 := mideaBill.MakeValSetChangeTx(v2.PubKey, v2.Power)

	makeApplyBlock(t, app, 1, diff, tx1, tx2)

	vals1, vals2 = vals[:nInit+2], app.Validators()
	valsEqual(t, vals1, vals2)

	// remove some validators
	v1, v2, v3 = vals[nInit-2], vals[nInit-1], vals[nInit]
	v1.Power = 0
	v2.Power = 0
	v3.Power = 0
	diff = []*types.Validator{v1, v2, v3}
	tx1 = mideaBill.MakeValSetChangeTx(v1.PubKey, v1.Power)
	tx2 = mideaBill.MakeValSetChangeTx(v2.PubKey, v2.Power)
	tx3 := mideaBill.MakeValSetChangeTx(v3.PubKey, v3.Power)

	makeApplyBlock(t, app, 2, diff, tx1, tx2, tx3)

	vals1 = append(vals[:nInit-2], vals[nInit+1])
	vals2 = app.Validators()
	valsEqual(t, vals1, vals2)

	// update some validators
	v1 = vals[0]
	if v1.Power == 5 {
		v1.Power = 6
	} else {
		v1.Power = 5
	}
	diff = []*types.Validator{v1}
	tx1 = mideaBill.MakeValSetChangeTx(v1.PubKey, v1.Power)

	makeApplyBlock(t, app, 3, diff, tx1)

	vals1 = append([]*types.Validator{v1}, vals1[1:len(vals1)]...)
	vals2 = app.Validators()
	valsEqual(t, vals1, vals2)

}

func makeApplyBlock(t *testing.T, mideaBill types.Application, heightInt int, diff []*types.Validator, txs ...[]byte) {
	// make and apply block
	height := uint64(heightInt)
	hash := []byte("foo")
	header := &types.Header{
		Height: height,
	}

	mideaBill.BeginBlock(hash, header)
	for _, tx := range txs {
		if r := mideaBill.DeliverTx(tx); r.IsErr() {
			t.Fatal(r)
		}
	}
	resEndBlock := mideaBill.EndBlock(height)
	mideaBill.Commit()

	valsEqual(t, diff, resEndBlock.Diffs)

}

// order doesn't matter
func valsEqual(t *testing.T, vals1, vals2 []*types.Validator) {
	if len(vals1) != len(vals2) {
		t.Fatalf("vals dont match in len. got %d, expected %d", len(vals2), len(vals1))
	}
	sort.Sort(types.Validators(vals1))
	sort.Sort(types.Validators(vals2))
	for i, v1 := range vals1 {
		v2 := vals2[i]
		if !bytes.Equal(v1.PubKey, v2.PubKey) ||
			v1.Power != v2.Power {
			t.Fatalf("vals dont match at index %d. got %X/%d , expected %X/%d", i, v2.PubKey, v2.Power, v1.PubKey, v1.Power)
		}
	}
}

func makeSocketClientServer(app types.Application, name string) (abcicli.Client, cmn.Service, error) {
	// Start the listener
	socket := cmn.Fmt("unix://%s.sock", name)
	logger := log.TestingLogger()

	server := server.NewSocketServer(socket, app)
	server.SetLogger(logger.With("module", "abci-server"))
	if _, err := server.Start(); err != nil {
		return nil, nil, err
	}

	// Connect to the socket
	client := abcicli.NewSocketClient(socket, false)
	client.SetLogger(logger.With("module", "abci-client"))
	if _, err := client.Start(); err != nil {
		server.Stop()
		return nil, nil, err
	}

	return client, server, nil
}

func makeGRPCClientServer(app types.Application, name string) (abcicli.Client, cmn.Service, error) {
	// Start the listener
	socket := cmn.Fmt("unix://%s.sock", name)
	logger := log.TestingLogger()

	gapp := types.NewGRPCApplication(app)
	server := server.NewGRPCServer(socket, gapp)
	server.SetLogger(logger.With("module", "abci-server"))
	if _, err := server.Start(); err != nil {
		return nil, nil, err
	}

	client := abcicli.NewGRPCClient(socket, true)
	client.SetLogger(logger.With("module", "abci-client"))
	if _, err := client.Start(); err != nil {
		server.Stop()
		return nil, nil, err
	}
	return client, server, nil
}

func TestClientServer(t *testing.T) {
	// set up socket app
	t.Skip()
	app := mideaBill.NewGwpointsApplication()
	client, server, err := makeSocketClientServer(app, "mideaBill-socket")
	require.Nil(t, err)
	defer server.Stop()
	defer client.Stop()

	runClientTests(t, client)

	// set up grpc app
	app = mideaBill.NewGwpointsApplication()
	gclient, gserver, err := makeGRPCClientServer(app, "mideaBill-grpc")
	require.Nil(t, err)
	defer gserver.Stop()
	defer gclient.Stop()

	runClientTests(t, gclient)
}

func runClientTests(t *testing.T, client abcicli.Client) {
	// run some tests....
	key := "abc"
	value := key
	tx := []byte(key)
	testClient(t, client, tx, key, value)

	value = "def"
	tx = []byte(key + "=" + value)
	testClient(t, client, tx, key, value)
}

func testClient(t *testing.T, app abcicli.Client, tx []byte, key, value string) {
	ar := app.DeliverTxSync(tx)
	require.False(t, ar.IsErr(), ar)
	// repeating tx doesn't raise error
	ar = app.DeliverTxSync(tx)
	require.False(t, ar.IsErr(), ar)

	// make sure query is fine
	resQuery, err := app.QuerySync(types.RequestQuery{
		Path: "/store",
		Data: []byte(key),
	})
	require.Nil(t, err)
	require.Equal(t, types.CodeType_OK, resQuery.Code)
	require.Equal(t, value, string(resQuery.Value))

	// make sure proof is fine
	resQuery, err = app.QuerySync(types.RequestQuery{
		Path:  "/store",
		Data:  []byte(key),
		Prove: true,
	})
	require.Nil(t, err)
	require.Equal(t, types.CodeType_OK, resQuery.Code)
	require.Equal(t, value, string(resQuery.Value))
	proof, err := iavl.ReadProof(resQuery.Proof)
	require.Nil(t, err)
	require.True(t, proof.Verify([]byte(key), resQuery.Value, proof.RootHash)) // NOTE: we have no way to verify the RootHash
}
*/
