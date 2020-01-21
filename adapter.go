package main

import (
	"fmt"
	gsrpc "github.com/centrifuge/go-substrate-rpc-client"
	"github.com/centrifuge/go-substrate-rpc-client/config"
	"github.com/centrifuge/go-substrate-rpc-client/signature"
	"github.com/centrifuge/go-substrate-rpc-client/types"
	"github.com/pkg/errors"
	"math/big"
	"net/url"
	"strconv"
	"strings"
)

type Argument struct {
	Type  string
	Value string
}

type Request struct {
	Function string
	Args     []Argument
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
	c, err := m.FindCallIndex(req.Function)
	if err != nil {
		return types.Call{}, err
	}

	args, err := convertTypes(req.Args)
	if err != nil {
		return types.Call{}, err
	}

	var a []byte
	for _, arg := range args {
		e, err := types.EncodeToBytes(arg)
		if err != nil {
			return types.Call{}, err
		}
		a = append(a, e...)
	}

	return types.Call{CallIndex: c, Args: a}, nil
}

func convertTypes(args []Argument) ([]interface{}, error) {
	var res []interface{}
	for _, arg := range args {
		switch strings.ToLower(arg.Type) {
		case "bool":
			var b types.Bool
			switch strings.ToLower(arg.Value) {
			case "true":
				b = types.NewBool(true)
			case "false":
				b = types.NewBool(false)
			default:
				return nil, errors.New("unable to parse bool")
			}
			res = append(res, b)
		case "uint256":
			i, err := strconv.ParseInt(arg.Value, 10, 64)
			if err != nil {
				return nil, errors.Wrap(err, "failed parsing uint256")
			}
			val := types.NewU256(*big.NewInt(i))
			res = append(res, val)
		case "int256":
			i, err := strconv.ParseInt(arg.Value, 10, 64)
			if err != nil {
				return nil, errors.Wrap(err, "failed parsing int256")
			}
			val := types.NewI256(*big.NewInt(i))
			res = append(res, val)
		case "ucompact":
			i, err := strconv.ParseInt(arg.Value, 10, 64)
			if err != nil {
				return nil, errors.Wrap(err, "failed parsing ucompact")
			}
			val := types.UCompact(i)
			res = append(res, val)
		case "bytes":
			res = append(res, types.Bytes(arg.Value))
		case "address":
			addr, err := types.NewAddressFromHexAccountID(fmt.Sprintf("%s", arg.Value))
			if err != nil {
				return nil, errors.Wrap(err, "unable to parse address")
			}
			res = append(res, addr)
		}
	}
	return res, nil
}
