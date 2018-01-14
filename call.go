package monero

import (
	"bytes"
	// "crypto/md5"
	// "encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	// "errors"
	"io"
	// "github.com/op/go-logging"
	"math/rand"
	"net/http"
	"io/ioutil"
	// "github.com/gorilla/rpc/json"
	// "github.com/haisum/rpcexample"
	// "strings"
)

// var log = logging.MustGetLogger("monero")

// ----------------------------------------------------------------------------
// Request and Response
// ----------------------------------------------------------------------------

// clientRequest represents a JSON-RPC request sent by a client.
type clientRequest struct {
	// JSON-RPC protocol.
	Version string `json:"jsonrpc"`

	// A String containing the name of the method to be invoked.
	Method string `json:"method"`

	// Object to pass as request parameter to the method.
	Params interface{} `json:"params"`

	// The request id. This can be of any type. It is used to match the
	// response with the request that it is replying to.
	Id uint64 `json:"id"`
}

// clientResponse represents a JSON-RPC response returned to a client.
type clientResponse struct {
	Version string           `json:"jsonrpc"`
	Result  *json.RawMessage `json:"result"`
	Error   *json.RawMessage `json:"error"`
}

type CallClient struct {
	endpoint string
	username string
	password string
}

func NewCallClient(endpoint, username, password string) *CallClient {
	return &CallClient{endpoint, username, password}
}

func (c *CallClient) Daemon(method string, req, rep interface{}) error {
	client := &http.Client{}
	data, _ :=  EncodeClientRequest(method, req)
	reqest, _ := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(data))
	reqest.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(reqest)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return DecodeClientResponse(resp.Body, rep)
}

func (c *CallClient) Wallet2(method string, req, rep interface{}) error {
	client := &http.Client{}
	log.Println("Rep: ", rep)
	data, _ :=  EncodeClientRequest(method, req)
	reqest, _ := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(data))
	reqest.Header.Set("Content-Type", "application/json")
	reqest.Header.Set("Connection", "Keep-Alive")
	// fmt.Println("request: %s", reqest)
	log.Println("request:", reqest)
	resp, err := client.Do(reqest)
	log.Println("response:", resp)
	log.Println("response body:", resp.Body)
	log.Println("response header:", resp.Header)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return DecodeClientResponse(resp.Body, rep)
}

func (c *CallClient) Wallet3(method string, req, rep interface{}) error {
	client := &http.Client{}
  data, err := EncodeClientRequest(method, req)
  if err != nil {
      return err
  }
  reqest, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(data))
	log.Println("request:", reqest)
	log.Println("request:", bytes.NewBuffer(data))
  if err != nil {
      return err
  }
  resp, err := client.Do(reqest)
  if err != nil {
      return err
  }
	respData, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(respData))
	log.Println("response:", resp)
  defer resp.Body.Close()

  // get a reader that can be "rewound"
  buf := bytes.NewBuffer(nil)
  if _, err := io.Copy(buf, resp.Body); err != nil {
      return err
  }
  br := bytes.NewReader(buf.Bytes())

  if _, err := io.Copy(ioutil.Discard, br); err != nil {
      return err
  }

  // rewind
  if _, err := br.Seek(0, 0); err != nil {
      return err
  }
  return DecodeClientResponse(br, rep)
}

func (c *CallClient) Wallet(method string, req, rep interface{}) error {
	client := &http.Client{}
  data, err := EncodeClientRequest(method, req)
  if err != nil {
      return err
  }
  reqest, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(data))
	//log.Println("request:", reqest)
	log.Println("request wallet:", bytes.NewBuffer(data))
  if err != nil {
      return err
  }
  resp, err := client.Do(reqest)
  if err != nil {
      return err
  }
	respData, err := ioutil.ReadAll(resp.Body)
	fmt.Println("Response wallet: ==>")

	log.Println("response wallet:", resp)
	err = DecodeClientResponse(bytes.NewReader(respData), rep)
	//log.Println("rep:", rep)
	//fmt.Println(string(rep))
	return err
}

// EncodeClientRequest encodes parameters for a JSON-RPC client request.
func EncodeClientRequest(method string, args interface{}) ([]byte, error) {
	c := &clientRequest{
		Version: "2.0",
		Method:  method,
		Params:  args,
		Id:      uint64(rand.Int63()),
	}
	return json.Marshal(c)
}

func DecodeClientResponse(r io.Reader, reply interface{}) error {
	var c clientResponse
	if err := json.NewDecoder(r).Decode(&c); err != nil {
		log.Println("read Decode Error:", c)
		return err
	}
	// log.Println("read body Result:", string(*c.Result))
	if c.Error != nil {
		jsonErr := &Error{}
		if err := json.Unmarshal(*c.Error, jsonErr); err != nil {
			log.Println("read Error Error:", string(*c.Error))
			return &Error{
				Code:    E_SERVER,
				Message: string(*c.Error),
			}
		}
		log.Println("read body Error:", string(*c.Error))
		return jsonErr
	}

	if c.Result == nil {
		return ErrNullResult
	}
	// log.Println("read body Result:", string(*c.Result))
	return json.Unmarshal(*c.Result, reply)
}
