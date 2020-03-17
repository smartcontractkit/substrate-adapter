package main

import (
	"fmt"
	gsrpc "github.com/centrifuge/go-substrate-rpc-client"
	"github.com/centrifuge/go-substrate-rpc-client/config"
	"github.com/centrifuge/go-substrate-rpc-client/signature"
	"github.com/centrifuge/go-substrate-rpc-client/types"
	"github.com/pkg/errors"
	"net/url"
	"strings"
)

type Request struct {
	Function  string
	Type      string
	Value     interface{}
	Result    interface{}
	RequestId interface{} `json:"request_id"`
}

type txType int

const (
	mortal txType = iota
	immortal
)

type substrateAdapter struct {
	keyringPair signature.KeyringPair
	txType      txType
	endpoint    url.URL
}

func newSubstrateAdapter(privkey, txtypeStr, endpoint string) (*substrateAdapter, error) {
	keypair, err := signature.KeyringPairFromSecret(privkey)
	if err != nil {
		return nil, err
	}

	var txtype txType
	switch strings.ToLower(txtypeStr) {
	case "mortal":
		txtype = mortal
	case "immortal":
		txtype = immortal
	default:
		return nil, errors.New("unexpected txtype provided")
	}

	if endpoint == "" {
		endpoint = config.Default().RPCURL
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	return &substrateAdapter{
		keyringPair: keypair,
		txType:      txtype,
		endpoint:    *u,
	}, nil
}

func (adapter substrateAdapter) handle(req Request) (interface{}, error) {
	// Set Value to whatever is defined in the "result" key
	// if the default "value" is empty
	if req.Value == nil || req.Value == "" {
		req.Value = req.Result
	}

	api, err := gsrpc.NewSubstrateAPI(adapter.endpoint.String())
	if err != nil {
		return nil, errors.Wrap(err, "failed getting substrate API")
	}

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		return nil, errors.Wrap(err, "failed getting metadata")
	}

	// Create custom function call using
	// request arguments
	c, err := NewCall(meta, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed creating call")
	}

	// Create the extrinsic
	ext := types.NewExtrinsic(c)

	genesisHash, err := api.RPC.Chain.GetBlockHash(0)
	if err != nil {
		return nil, errors.Wrap(err, "failed getting genesis block hash")
	}

	rv, err := api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		return nil, errors.Wrap(err, "failed getting runtime version")
	}

	// Get account nonce
	key, err := types.CreateStorageKey(meta, "System", "AccountNonce", adapter.keyringPair.PublicKey, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed getting account nonce key")
	}

	var nonce uint32
	err = api.RPC.State.GetStorageLatest(key, &nonce)
	if err != nil {
		return nil, errors.Wrap(err, "failed getting account nonce")
	}

	era := types.ExtrinsicEra{}
	var blockHash types.Hash
	if adapter.txType == immortal {
		blockHash = genesisHash
		era.IsMortalEra = false
		era.IsImmortalEra = true
	} else {
		blockHash, err = api.RPC.Chain.GetBlockHashLatest()
		if err != nil {
			return nil, errors.Wrap(err, "failed getting latest block hash")
		}
		era.IsMortalEra = true
		era.IsImmortalEra = false
	}

	o := types.SignatureOptions{
		BlockHash:   blockHash,
		Era:         era,
		GenesisHash: genesisHash,
		Nonce:       types.UCompact(nonce),
		SpecVersion: rv.SpecVersion,
		Tip:         0,
	}

	// Sign the transaction
	err = ext.Sign(adapter.keyringPair, o)
	if err != nil {
		return nil, errors.Wrap(err, "failed signing the transaction")
	}

	// Send the extrinsic
	hash, err := api.RPC.Author.SubmitExtrinsic(ext)
	if err != nil {
		return nil, errors.Wrap(err, "failed submitting the transaction")
	}

	return hash, nil
}

func NewCall(m *types.Metadata, req Request) (types.Call, error) {
	reqId, err := convertTypes("uint64", fmt.Sprintf("%v", req.RequestId))
	if err != nil {
		return types.Call{}, err
	}

	value, err := convertTypes(req.Type, fmt.Sprintf("%v", req.Value))
	if err != nil {
		return types.Call{}, err
	}

	arg, err := types.EncodeToBytes(value)
	if err != nil {
		return types.Call{}, err
	}

	return types.NewCall(m, req.Function, reqId, arg)
}
