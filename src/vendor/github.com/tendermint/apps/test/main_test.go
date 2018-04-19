package appstest

import (
	"os"
	"testing"
	//"time"

	"github.com/tendermint/tendermint/apps/mainapp"
	nm "github.com/tendermint/tendermint/node"
)

var node *nm.Node

func TestMain(m *testing.M) {
	// start a tendermint node (and merkleeyes) in the background to test against
	config := GetConfig()
	app := mainapp.NewApplication(config.RootDir)
	node = StartTendermint(app)
	clientJSON = GetJSONClient()
	code := m.Run()

	// and shut down proper at the end
	node.Stop()
	node.Wait()
	os.Exit(code)
}
