package appstest

import (
	"testing"
)

func init() {
}

func TestMainTransfer(t *testing.T) {
	toaddr := mainRegSenderKey
	var signs []*[64]byte
	signs = append(signs, mainTransferAdminKey)
	resp, err := mainTransfer(mainTransferAdminKey, signs, 1, toaddr[32:], 100*1e8)
	if err != nil {
		t.Error(err)
		return
	}
	resp, err = mainWalletInfo(toaddr, signs, 1, toaddr[32:])
	if err != nil {
		t.Error(err)
		return
	}
	if resp.GetWalletInfo().Active != 100*1e8 {
		t.Error("send to : balance error")
	}

	//reg app
	resp, err = mainRegApp(mainRegSenderKey, mainAppUserKey1, mainAppUserKey2[32:], 100)
	if err != nil {
		t.Error(err)
		return
	}

	//query appconf

	//mainAppconf()
}

func TestMainWeight(t *testing.T) {
	var signs []*[64]byte
	signs = append(signs, mainTransferAdminKey)
	//signs = append(signs, mainTransferAdminKey)
	account := mainTransferAdminKey
	subaccount := mainAppUserKey1

	resp, err := mainTransfer(mainTransferAdminKey, signs, 1, subaccount[32:], 100*1e8)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(resp)
	// set account
	resp, err = mainSetWeight(mainTransferAdminKey, signs, 1, account[32:], 1, 2, 3)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(resp)

	resp, err = mainWeightInfo(mainTransferAdminKey, signs, 1, account[32:])
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(resp)

	resp, err = mainTransfer(mainTransferAdminKey, signs, 1, subaccount[32:], 100*1e8)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(resp)

	// set subaccount
	resp, err = mainSetWeight(mainWeightSenderKey, signs, 1, subaccount[32:], 1, 2, 3)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(resp)

	resp, err = mainWeightInfo(mainWeightSenderKey, signs, 1, subaccount[32:])
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(resp)

	resp, err = mainDelWeight(mainWeightSenderKey, signs, 1, subaccount[32:])
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(resp)

}

func TestMainAccountManage(t *testing.T) {
	var signs []*[64]byte
	signs = append(signs, mainTransferAdminKey)
	//signs = append(signs, mainTransferAdminKey)
	account := mainTransferAdminKey
	//subaccount := mainAppUserKey1

	// set account (key *[64]byte, signs []*[64]byte, account []byte, frozen int64, active int64, coinId int32, transfer int32)
	resp, err := mainSetAccountManage(mainTransferAdminKey, signs, account[32:], 1, 2, 1, 1)
	//resp, err := mainSetAccountManage(mainAccountSenderKey, signs, account[32:], 1, 2, 1,1)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(resp)

	//query account
	resp, err = mainAccountInfo(mainAccountSenderKey, signs, account[32:], 1, 2, 1, 1)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(resp)
}
