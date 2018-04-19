package appstest

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/abci/types"
	wire "github.com/tendermint/go-wire"
	wdata "github.com/tendermint/go-wire/data"
	"github.com/tendermint/tendermint/apps"
	cfg "github.com/tendermint/tendermint/config"
	nm "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/proxy"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	core_grpc "github.com/tendermint/tendermint/rpc/grpc"
	client "github.com/tendermint/tendermint/rpc/lib/client"
	"github.com/tendermint/tendermint/types"
	tmlog "github.com/tendermint/tmlibs/log"
)

var (
	conf *cfg.Config
)

type TMResult interface {
	//rpctypes.RPCResponse.Result
}

// f**ing long, but unique for each test
func makePathname() string {
	// get path
	p, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	sep := string(filepath.Separator)
	return strings.Replace(p, sep, "_", -1)
}

func randPort() int {
	return 26656
	// returns between base and base + spread
	//base, spread := 20000, 20000
	//return base + rand.Intn(spread)
}

func makeAddrs() (string, string, string) {
	start := randPort()
	return fmt.Sprintf("tcp://0.0.0.0:%d", start),
		fmt.Sprintf("tcp://0.0.0.0:%d", start+1),
		fmt.Sprintf("tcp://0.0.0.0:%d", start+2)
}

// GetConfig returns a config for the test cases as a singleton
func GetConfig() *cfg.Config {
	if conf == nil {
		pathname := makePathname()
		conf = cfg.ResetTestRoot(pathname)
		// and we use random ports to run in parallel
		tm, rpc, grpc := makeAddrs()
		println(conf.Consensus.TimeoutCommit)
		conf.P2P.ListenAddress = tm
		conf.RPCListenAddress = rpc
		conf.GRPCListenAddress = grpc
	}
	return conf
}

// GetURIClient gets a uri client pointing to the test tendermint rpc
func GetURIClient() *client.URIClient {
	rpcAddr := GetConfig().RPCListenAddress
	return client.NewURIClient(rpcAddr)
}

// GetJSONClient gets a http/json client pointing to the test tendermint rpc
func GetJSONClient() *client.JSONRPCClient {
	rpcAddr := GetConfig().RPCListenAddress
	return client.NewJSONRPCClient(rpcAddr)
}

func GetGRPCClient() core_grpc.BroadcastAPIClient {
	grpcAddr := GetConfig().GRPCListenAddress
	return core_grpc.StartGRPCClient(grpcAddr)
}

func GetWSClient() *client.WSClient {
	rpcAddr := GetConfig().RPCListenAddress
	wsc := client.NewWSClient(rpcAddr, "/websocket")
	if _, err := wsc.Start(); err != nil {
		panic(err)
	}
	return wsc
}

// StartTendermint starts a test tendermint server in a go routine and returns when it is initialized
func StartTendermint(app abci.Application) *nm.Node {
	node := NewTendermint(app)
	node.Start()
	fmt.Println("Tendermint running!")
	return node
}

// NewTendermint creates a new tendermint server and sleeps forever
func NewTendermint(app abci.Application) *nm.Node {
	// Create & start node
	config := GetConfig()
	logger := tmlog.NewTMLogger(tmlog.NewSyncWriter(os.Stdout))
	logger = tmlog.NewFilter(logger, tmlog.AllowError())
	privValidatorFile := config.PrivValidatorFile()
	privValidator := types.LoadOrGenPrivValidator(privValidatorFile, logger)
	papp := proxy.NewLocalClientCreator(app)
	node := nm.NewNode(config, privValidator, papp, logger)
	return node
}

//--------------------------------------------------------------------------------
// Utilities for testing the websocket service

// wait for an event; do things that might trigger events, and check them when they are received
// the check function takes an event id and the byte slice read off the ws
func waitForEvent(t *testing.T, wsc *client.WSClient, eventid string, dieOnTimeout bool, f func(), check func(string, interface{}) error) {
	// go routine to wait for webscoket msg
	goodCh := make(chan interface{})
	errCh := make(chan error)

	// Read message
	go func() {
		var err error
	LOOP:
		for {
			select {
			case r := <-wsc.ResultsCh:
				result := new(TMResult)
				wire.ReadJSONPtr(result, r, &err)
				if err != nil {
					errCh <- err
					break LOOP
				}
				event, ok := (*result).(*ctypes.ResultEvent)
				if ok && event.Name == eventid {
					goodCh <- event.Data
					break LOOP
				}
			case err := <-wsc.ErrorsCh:
				errCh <- err
				break LOOP
			case <-wsc.Quit:
				break LOOP
			}
		}
	}()

	// do stuff (transactions)
	f()

	// wait for an event or timeout
	timeout := time.NewTimer(10 * time.Second)
	select {
	case <-timeout.C:
		if dieOnTimeout {
			wsc.Stop()
			require.True(t, false, "%s event was not received in time", eventid)
		}
		// else that's great, we didn't hear the event
		// and we shouldn't have
	case eventData := <-goodCh:
		if dieOnTimeout {
			// message was received and expected
			// run the check
			require.Nil(t, check(eventid, eventData))
		} else {
			wsc.Stop()
			require.True(t, false, "%s event was not expected", eventid)
		}
	case err := <-errCh:
		panic(err) // Show the stack trace.
	}
}

//------------------------------------

var clientJSON *client.JSONRPCClient //

var mainTransferAdminKey = apps.HexToPrivkey("229524531c56aa785c0cf091c9df9297c72473049ef04a4abefba306dea81a48fc19b9384e58c968455f080a3247290d3f061fb4843b1513f5bca26e0a656bea")

var mainRegSenderKey = apps.HexToPrivkey("31854ca7a7f25c3c9b2befe6d109df533b46205c81863305cb50dff5ac4e792385faa6aaa166752ab9e0175d2d6e551596c50ea11f4bc3337748d099dcf94ac0")

var mainAppUserKey1 = apps.HexToPrivkey("cc345dca2be0d4d24105ccfe11dc57efb163afbb8dd97932f6b86c0ac2139ba89a71f301616ac645924b04bbce992fb9e161f0af09d074dcc79cab7ced9e9441")

var mainAppUserKey2 = apps.HexToPrivkey("258bdc42dbb578e94b0bdd47f798b27052ba325ed50e5dddcd0a24673bb021dd122a6b5ffc3b5adb939bcf929d2faea1de339dd4ce9d00933d9ab28580f10502")

var mainWeightSenderKey = apps.HexToPrivkey("31854ca7a7f25c3c9b2befe6d109df533b46205c81863305cb50dff5ac4e792385faa6aaa166752ab9e0175d2d6e551596c50ea11f4bc3337748d099dcf94ac0")

var mainAccountSenderKey = apps.HexToPrivkey("31854ca7a7f25c3c9b2befe6d109df533b46205c81863305cb50dff5ac4e792385faa6aaa166752ab9e0175d2d6e551596c50ea11f4bc3337748d099dcf94ac0")

func mainSend(key *[64]byte, request *apps.WriteRequest) (*apps.Response, error) {
	d, err := apps.Encode(request)
	if err != nil {
		return nil, err
	}
	log.Printf("tx hash = %s\n", hex.EncodeToString(apps.HashSha256(d)))
	var tx apps.MainTx
	tx.AppId = int32(request.GetAppId())
	tx.Data = d
	signdata := make([]byte, 96)
	copy(signdata[:], key[32:])
	copy(signdata[32:], apps.Signdata(key, tx.Data))
	tx.Sign = signdata
	tx.SignType = int32(apps.Sign_Ed25519)
	data, err := apps.Encode(&tx)
	if err != nil {
		return nil, err
	}

	//send to block
	tmResult := new(ctypes.ResultBroadcastTxCommit) //ctypes.TMResult
	_, err = clientJSON.Call("broadcast_tx_commit", map[string]interface{}{"tx": data}, tmResult)
	if err != nil {
		return nil, err
	}

	res := (*tmResult) //.(*ctypes.ResultBroadcastTxCommit)
	if !res.CheckTx.Code.IsOK() {
		log.Printf("where error happens")
		return nil, fmt.Errorf("err:code=%v,str=%v", res.CheckTx.Code, string(res.CheckTx.Log))
	}

	if !res.DeliverTx.Code.IsOK() {
		return nil, fmt.Errorf("err:code=%v,str=%v", res.DeliverTx.Code, string(res.DeliverTx.Log))
	}
	//log.Println(res)
	var resp apps.Response
	err = apps.Decode(res.DeliverTx.Data, &resp)
	if err != nil {
		return nil, err
	}

	//log.Println(hex.EncodeToString(res.Data))
	log.Printf("result = %s\n", hex.EncodeToString(resp.GetHash().GetTxHash()))
	return &resp, nil
}

func mainQuery(request *apps.ReadRequest) (*apps.Response, error) {
	data, err := apps.Encode(request)
	if err != nil {
		return nil, err
	}
	tmResult := new(ctypes.ResultABCIQuery)
	_, err = clientJSON.Call("abci_query", map[string]interface{}{"path": "", "data": wdata.Bytes(data), "prove": false}, tmResult)
	if err != nil {
		return nil, err
	}
	query := tmResult
	if !query.Code.IsOK() {
		return nil, fmt.Errorf("%v", query.Code)
	}
	var resp apps.Response
	log.Println("query", hex.EncodeToString(query.Value))
	err = apps.Decode(query.Value, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func mainRegApp(sender *[64]byte, account *[64]byte, admin []byte, appId int32) (*apps.Response, error) {
	req := &apps.RequestCreateApp{}
	req.Admin = admin
	req.AppId = appId
	request := &apps.WriteRequest{}
	request.Account = account[32:]
	request.AppId = apps.App_Reg
	request.Value = &apps.WriteRequest_CreateApp{req}
	request.InstructionId = time.Now().UnixNano()
	request.ActionId = apps.MessageType_MsgCreateApp
	var accs []*[64]byte
	accs = append(accs, account)
	request.Signdatas(accs)
	return mainSend(sender, request)
}

func mainWalletInfo(key *[64]byte, signs []*[64]byte, coinId int32, addr []byte) (*apps.Response, error) {
	req := &apps.RequestWalletInfo{}
	req.CoinId = coinId

	request := &apps.ReadRequest{}
	request.ActionId = apps.MessageType_MsgWalletInfo
	request.Account = addr
	request.Value = &apps.ReadRequest_Walletinfo{req}
	request.AppId = apps.App_Transfer
	request.ActionId = apps.MessageType_MsgWalletInfo
	request.Signdatas(signs)
	return mainQuery(request)
}

func mainTransfer(key *[64]byte, signs []*[64]byte, coinId int32, toaddr []byte, amount int64) (*apps.Response, error) {
	req := &apps.RequestTransfer{}
	req.Amount = amount
	req.CoinId = coinId
	req.ToAddr = toaddr
	request := &apps.WriteRequest{}
	request.AppId = apps.App_Transfer
	request.Value = &apps.WriteRequest_Transfer{req}
	request.ActionId = apps.MessageType_MsgTransfer
	request.Account = key[32:]
	request.InstructionId = time.Now().UnixNano()
	request.Signdatas(signs)
	return mainSend(key, request)
}

func getBlock(height int) (txs types.Txs, err error) {
	tmResult := new(TMResult)
	_, err = clientJSON.Call("block", map[string]interface{}{"height": height}, tmResult)
	if err != nil {
		return nil, err
	}
	block := (*tmResult).(*ctypes.ResultBlock)
	return block.Block.Txs, nil
}

func mainSetWeight(key *[64]byte, signs []*[64]byte, coinId int32, subaccount []byte, transfer int32, query int32, manager int32) (*apps.Response, error) {
	req := &apps.RequestSetWeight{}
	req.Subaccount = subaccount
	req.CoinId = coinId
	req.TransferWeight = transfer
	req.QueryWeight = query
	req.ManagerWeight = manager

	request := &apps.WriteRequest{}
	request.AppId = apps.App_AccountManage
	request.Value = &apps.WriteRequest_SetWeight{req}
	request.ActionId = apps.MessageType_MsgSetWeight
	request.Account = key[32:]
	request.InstructionId = time.Now().UnixNano()
	request.Signdatas(signs)

	return mainSend(key, request)
}

func mainDelWeight(key *[64]byte, signs []*[64]byte, coinId int32, subaccount []byte) (*apps.Response, error) {
	req := &apps.RequestDelWeight{}
	req.Subaccount = subaccount
	req.CoinId = coinId

	request := &apps.WriteRequest{}
	request.AppId = apps.App_AccountManage
	request.Value = &apps.WriteRequest_DelWeight{req}
	request.ActionId = apps.MessageType_MsgDelWeight
	request.Account = key[32:]
	request.InstructionId = time.Now().UnixNano()
	request.Signdatas(signs)

	return mainSend(key, request)
}

func mainWeightInfo(key *[64]byte, signs []*[64]byte, coinId int32, subaccount []byte) (*apps.Response, error) {
	req := &apps.RequestWeightInfo{}
	req.Subaccount = subaccount
	req.CoinId = coinId

	request := &apps.ReadRequest{}
	request.AppId = apps.App_AccountManage
	request.Value = &apps.ReadRequest_Weightinfo{req}
	request.ActionId = apps.MessageType_MsgWeightInfo
	request.Account = key[32:]
	request.Signdatas(signs)

	return mainQuery(request)
}

func mainSetAccountManage(key *[64]byte, signs []*[64]byte, account []byte, frozen int64, active int64, coinId int32, transfer int32) (*apps.Response, error) {
	req := &apps.RequestSetAccount{}
	req.Account = account
	req.Frozen = frozen
	req.Active = active
	req.CoinId = coinId
	req.Transfer = transfer

	request := &apps.WriteRequest{}
	request.AppId = apps.App_AccountManage
	request.Value = &apps.WriteRequest_SetAccount{req}
	request.ActionId = apps.MessageType_MsgSetAccount
	request.Account = key[32:]
	request.InstructionId = time.Now().UnixNano()
	request.Signdatas(signs)
	return mainSend(key, request)
}

func mainAccountInfo(key *[64]byte, signs []*[64]byte, account []byte, frozen int64, active int64, coinId int32, transfer int32) (*apps.Response, error) {
	req := &apps.RequestAccountInfo{}
	req.Account = account
	req.Frozen = frozen
	req.Active = active
	req.CoinId = coinId
	req.Transfer = transfer

	request := &apps.ReadRequest{}
	request.AppId = apps.App_AccountManage
	request.Value = &apps.ReadRequest_Accountinfo{req}
	request.ActionId = apps.MessageType_MsgAccountInfo
	request.Account = key[32:]
	request.Signdatas(signs)
	return mainQuery(request)
}
