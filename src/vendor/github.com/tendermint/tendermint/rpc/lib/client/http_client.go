package rpcclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	types "github.com/tendermint/tendermint/rpc/lib/types"
	cmn "github.com/tendermint/tmlibs/common"
)

// HTTPClient is a common interface for JSONRPCClient and URIClient.
type HTTPClient interface {
	Call(method string, params map[string]interface{}, result interface{}) (interface{}, error)
}

// TODO: Deprecate support for IP:PORT or /path/to/socket
func makeHTTPDialer(remoteAddr string) (string, func(string, string) (net.Conn, error)) {

	parts := strings.SplitN(remoteAddr, "://", 2)
	var protocol, address string
	if len(parts) != 2 {
		cmn.PanicSanity(fmt.Sprintf("Expected fully formed listening address, including the tcp:// or unix:// prefix, given %s", remoteAddr))
	} else {
		protocol, address = parts[0], parts[1]
	}

	trimmedAddress := strings.Replace(address, "/", ".", -1) // replace / with . for http requests (dummy domain)
	return trimmedAddress, func(proto, addr string) (net.Conn, error) {
		return net.Dial(protocol, address)
	}
}

// We overwrite the http.Client.Dial so we can do http over tcp or unix.
// remoteAddr should be fully featured (eg. with tcp:// or unix://)
func makeHTTPClient(remoteAddr string) (string, *http.Client) {
	address, dialer := makeHTTPDialer(remoteAddr)
	return "http://" + address, &http.Client{
		Transport: &http.Transport{
			Dial: dialer,
		},
	}
}

//------------------------------------------------------------------------------------

// JSON rpc takes params as a slice
type JSONRPCClient struct {
	address string
	client  *http.Client
}

func NewJSONRPCClient(remote string) *JSONRPCClient {
	address, client := makeHTTPClient(remote)
	return &JSONRPCClient{
		address: address,
		client:  client,
	}
}

func (c *JSONRPCClient) Call(method string, params map[string]interface{}, result interface{}) (interface{}, error) {
	request, err := types.MapToRequest("", method, params)
	if err != nil {
		return nil, err
	}
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	//println(string(requestBytes))
	requestBuf := bytes.NewBuffer(requestBytes)
	//println(cmn.Fmt("RPC request to (%v): %v", method, string(requestBytes)))
	httpResponse, err := c.client.Post(c.address, "text/json", requestBuf)
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close()
	responseBytes, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}
	//println(cmn.Fmt("RPC response: %v", string(responseBytes)))
	return unmarshalResponseBytes(responseBytes, result)
}

//-------------------------------------------------------------

// URI takes params as a map
type URIClient struct {
	address string
	client  *http.Client
}

func NewURIClient(remote string) *URIClient {
	address, client := makeHTTPClient(remote)
	return &URIClient{
		address: address,
		client:  client,
	}
}

func (c *URIClient) Call(method string, params map[string]interface{}, result interface{}) (interface{}, error) {
	values, err := argsToURLValues(params)
	if err != nil {
		return nil, err
	}
	//println(cmn.Fmt("URI request to %v (%v): %v", c.address, method, values))
	resp, err := c.client.PostForm(c.address+"/"+method, values)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	//println(cmn.Fmt("RPC response: %v", string(responseBytes)))
	return unmarshalResponseBytes(responseBytes, result)
}

//------------------------------------------------

func unmarshalResponseBytes(responseBytes []byte, result interface{}) (interface{}, error) {
	// read response
	// if rpc/core/types is imported, the result will unmarshal
	// into the correct type
	// log.Notice("response", "response", string(responseBytes))
	var err error
	response := &types.RPCResponse{}
	err = json.Unmarshal(responseBytes, response)
	if err != nil {
		return nil, errors.Errorf("Error unmarshalling rpc response: %v", err)
	}
	errorStr := response.Error
	if errorStr != "" {
		return nil, errors.Errorf("Response error: %v", errorStr)
	}
	// unmarshal the RawMessage into the result
	err = json.Unmarshal(*response.Result, result)
	if err != nil {
		return nil, errors.Errorf("Error unmarshalling rpc response result: %v", err)
	}
	return result, nil
}

func argsToURLValues(args map[string]interface{}) (url.Values, error) {
	values := make(url.Values)
	if len(args) == 0 {
		return values, nil
	}
	err := argsToJson(args)
	if err != nil {
		return nil, err
	}
	for key, val := range args {
		values.Set(key, val.(string))
	}
	return values, nil
}

func argsToJson(args map[string]interface{}) error {
	for k, v := range args {
		rt := reflect.TypeOf(v)
		isByteSlice := rt.Kind() == reflect.Slice && rt.Elem().Kind() == reflect.Uint8
		if isByteSlice {
			bytes := reflect.ValueOf(v).Bytes()
			args[k] = fmt.Sprintf("0x%X", bytes)
			continue
		}

		// Pass everything else to go-wire
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		args[k] = string(data)
	}
	return nil
}