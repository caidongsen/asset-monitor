package supplychain_test

import (
	"encoding/hex"
	"testing"

	abci "github.com/tendermint/abci/types"

	"github.com/tendermint/tendermint/apps/gmgop"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	rpc "github.com/tendermint/tendermint/rpc/lib/client"

	"bytes"
	"io/ioutil"
	"sort"

	"github.com/stretchr/testify/require"
	abcicli "github.com/tendermint/abci/client"
	"github.com/tendermint/abci/server"
	"github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/merkleeyes/iavl"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/log"
)

func genUserCreateTx(t *testing.T, uid string, instructionId int64) []byte {
	req := &gmgop.RequestUserCreate{}
	req.Uid = uid
	req.Pubkey, _ = hex.DecodeString("f6e008a45ffc21a902ed34b4b4de04b743e20677855cc278c70192fb1f476f34")
	req.Info, _ = hex.DecodeString("fa7fa0acc089f59aea2e21d8b23c67503a8025d27a5cde3bfd1445f01f5ad8e5")
	req.Role = 1
	request := &gmgop.Request{}
	request.Value = &gmgop.Request_UserCreate{req}
	request.Uid = uid
	request.InstructionId = instructionId
	request.Pubkey, _ = hex.DecodeString("f6e008a45ffc21a902ed34b4b4de04b743e20677855cc278c70192fb1f476f34")
	request.ActionId = gmgop.MessageType_MsgUserCreate
	data2, err := gmgop.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	priv, _ := hex.DecodeString("8a1a5a9ce333e10704e58bc331f9693ee2e126a98dc365bed57020bedc22a2cbf6e008a45ffc21a902ed34b4b4de04b743e20677855cc278c70192fb1f476f34")
	request.Sign = gmgop.Signdata(priv, data2)

	data2, err = gmgop.MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	return data2
}

func TestUserCreate(t *testing.T) {
	data := genUserCreateTx(t, "uid007", 1)
	testBroadcastTxCommit(t, data, clientJSON)

	data = genUserCreateTx(t, "uid007", 2)
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

func testGmgop(t *testing.T, app types.Application, tx []byte) {
	ar := app.DeliverTx(tx)
	require.False(t, ar.IsErr(), ar)
	// repeating tx doesn't raise error
	ar = app.DeliverTx(tx)
	require.True(t, ar.IsErr(), ar)
}

func TestGmgopKV(t *testing.T) {
	gmgop := gmgop.NewGmgopApplication()
	tx := genUserCreateTx(t, "uid008", 3)
	testGmgop(t, gmgop, tx)
}

func TestPersistentGmgopKV(t *testing.T) {
	dir, err := ioutil.TempDir("/tmp", "abci-gmgop-test") // TODO
	if err != nil {
		t.Fatal(err)
	}
	gmgop := gmgop.NewPersistentGmgopApplication(dir)
	tx := genUserCreateTx(t, "uid009", 4)
	testGmgop(t, gmgop, tx)
}

func TestPersistentGmgopInfo(t *testing.T) {
	dir, err := ioutil.TempDir("/tmp", "abci-gmgop-test") // TODO
	if err != nil {
		t.Fatal(err)
	}
	gmgop := gmgop.NewPersistentGmgopApplication(dir)
	height := uint64(0)

	resInfo := gmgop.Info()
	if resInfo.LastBlockHeight != height {
		t.Fatalf("expected height of %d, got %d", height, resInfo.LastBlockHeight)
	}

	// make and apply block
	height = uint64(1)
	hash := []byte("foo")
	header := &types.Header{
		Height: uint64(height),
	}
	gmgop.BeginBlock(hash, header)
	gmgop.EndBlock(height)
	gmgop.Commit()

	resInfo = gmgop.Info()
	if resInfo.LastBlockHeight != height {
		t.Fatalf("expected height of %d, got %d", height, resInfo.LastBlockHeight)
	}

}

// add a validator, remove a validator, update a validator
func TestValSetChanges(t *testing.T) {
	dir, err := ioutil.TempDir("/tmp", "abci-gmgop-test") // TODO
	if err != nil {
		t.Fatal(err)
	}
	app := gmgop.NewPersistentGmgopApplication(dir)

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
	tx1 := gmgop.MakeValSetChangeTx(v1.PubKey, v1.Power)
	tx2 := gmgop.MakeValSetChangeTx(v2.PubKey, v2.Power)

	makeApplyBlock(t, app, 1, diff, tx1, tx2)

	vals1, vals2 = vals[:nInit+2], app.Validators()
	valsEqual(t, vals1, vals2)

	// remove some validators
	v1, v2, v3 = vals[nInit-2], vals[nInit-1], vals[nInit]
	v1.Power = 0
	v2.Power = 0
	v3.Power = 0
	diff = []*types.Validator{v1, v2, v3}
	tx1 = gmgop.MakeValSetChangeTx(v1.PubKey, v1.Power)
	tx2 = gmgop.MakeValSetChangeTx(v2.PubKey, v2.Power)
	tx3 := gmgop.MakeValSetChangeTx(v3.PubKey, v3.Power)

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
	tx1 = gmgop.MakeValSetChangeTx(v1.PubKey, v1.Power)

	makeApplyBlock(t, app, 3, diff, tx1)

	vals1 = append([]*types.Validator{v1}, vals1[1:len(vals1)]...)
	vals2 = app.Validators()
	valsEqual(t, vals1, vals2)

}

func makeApplyBlock(t *testing.T, gmgop types.Application, heightInt int, diff []*types.Validator, txs ...[]byte) {
	// make and apply block
	height := uint64(heightInt)
	hash := []byte("foo")
	header := &types.Header{
		Height: height,
	}

	gmgop.BeginBlock(hash, header)
	for _, tx := range txs {
		if r := gmgop.DeliverTx(tx); r.IsErr() {
			t.Fatal(r)
		}
	}
	resEndBlock := gmgop.EndBlock(height)
	gmgop.Commit()

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
	app := gmgop.NewGmgopApplication()
	client, server, err := makeSocketClientServer(app, "gmgop-socket")
	require.Nil(t, err)
	defer server.Stop()
	defer client.Stop()

	runClientTests(t, client)

	// set up grpc app
	app = gmgop.NewGmgopApplication()
	gclient, gserver, err := makeGRPCClientServer(app, "gmgop-grpc")
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
