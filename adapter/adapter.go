package adapter

import (
	"fmt"
	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v3"
	"github.com/centrifuge/go-substrate-rpc-client/v3/config"
	"github.com/centrifuge/go-substrate-rpc-client/v3/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"math/big"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

type Request struct {
	// Common params:
	RequestType string `json:"request_type"`
	Value       string `json:"value"`
	Result      string `json:"result"`
	Multiply    int64  `json:"multiply,omitempty"`

	// Runlog params:
	Function  string      `json:"function"`
	Type      string      `json:"type"`
	RequestId interface{} `json:"request_id"`

	// Fluxmonitor params:
	FeedId  string `json:"feed_id"`
	RoundId string `json:"round_id"`
}

type txType int

const (
	mortal txType = iota
	immortal
)

type substrateAdapter struct {
	client *substrateClient
}

type substrateClient struct {
	keyringPair signature.KeyringPair
	txType      txType

	nonceMutex *sync.Mutex
	nonce      *uint32

	api            *gsrpc.SubstrateAPI
	meta           *types.Metadata
	genesisHash    types.Hash
	runtimeVersion *types.RuntimeVersion
}

func NewSubstrateClient(u *url.URL, keyringPair signature.KeyringPair, t txType) (*substrateClient, error) {
	api, err := gsrpc.NewSubstrateAPI(u.String())
	if err != nil {
		return nil, errors.Wrap(err, "failed getting substrate API")
	}

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		return nil, errors.Wrap(err, "failed getting metadata")
	}

	genesisHash, err := api.RPC.Chain.GetBlockHash(0)
	if err != nil {
		return nil, errors.Wrap(err, "failed getting genesis block hash")
	}

	rv, err := api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		return nil, errors.Wrap(err, "failed getting runtime version")
	}

	client := &substrateClient{
		keyringPair:    keyringPair,
		txType:         t,
		api:            api,
		meta:           meta,
		genesisHash:    genesisHash,
		runtimeVersion: rv,

		nonceMutex: &sync.Mutex{},
	}

	nonce, err := client.FetchLatestNonce()
	if err != nil {
		return nil, err
	}
	client.nonce = &nonce

	return client, nil
}

func (c substrateClient) Sign(ext *types.Extrinsic) error {
	era := types.ExtrinsicEra{}
	var blockHash types.Hash
	var err error
	if c.txType == immortal {
		blockHash = c.genesisHash
		era.IsMortalEra = false
		era.IsImmortalEra = true
	} else {
		blockHash, err = c.api.RPC.Chain.GetBlockHashLatest()
		if err != nil {
			return errors.Wrap(err, "failed getting latest block hash")
		}
		era.IsMortalEra = true
		era.IsImmortalEra = false
	}

	o := types.SignatureOptions{
		Era:                era,
		Nonce:              types.NewUCompact(new(big.Int).SetInt64(int64(c.GetAndIncrementNonce()))),
		Tip:                types.NewUCompact(new(big.Int).SetInt64(0)),
		SpecVersion:        c.runtimeVersion.SpecVersion,
		GenesisHash:        c.genesisHash,
		BlockHash:          blockHash,
		TransactionVersion: c.runtimeVersion.TransactionVersion,
	}

	// Sign the transaction
	return ext.Sign(c.keyringPair, o)
}

func (c substrateClient) FetchLatestNonce() (uint32, error) {
	key, err := types.CreateStorageKey(c.meta, "System", "Account", c.keyringPair.PublicKey, nil)
	if err != nil {
		return 0, errors.Wrap(err, "failed getting account nonce key")
	}

	var nonce uint32
	ok, err := c.api.RPC.State.GetStorageLatest(key, &nonce)
	if !ok || err != nil {
		return 0, errors.Wrap(err, "failed getting account nonce")
	}

	return nonce, nil
}

func (c substrateClient) GetAndIncrementNonce() uint32 {
	c.nonceMutex.Lock()
	defer c.nonceMutex.Unlock()
	nonce := *c.nonce
	atomic.AddUint32(c.nonce, 1)
	return nonce
}

func (c substrateClient) SubmitCall(call types.Call) (types.Hash, error) {
	// Create the extrinsic
	ext := types.NewExtrinsic(call)
	err := c.Sign(&ext)
	if err != nil {
		return types.Hash{}, errors.Wrap(err, "failed to sign extrinsic")
	}

	// Send the extrinsic
	return c.api.RPC.Author.SubmitExtrinsic(ext)
}

func NewSubstrateAdapter(privkey, txtypeStr, endpoint string) (*substrateAdapter, error) {
	keypair, err := signature.KeyringPairFromSecret(privkey, 0)
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

	client, err := NewSubstrateClient(u, keypair, txtype)
	if err != nil {
		return nil, err
	}

	return &substrateAdapter{
		client: client,
	}, nil
}

func (adapter substrateAdapter) Handle(req Request) (interface{}, error) {
	// Set Value to whatever is defined in the "result" key
	// if the default "value" is empty
	if req.Result != "" {
		req.Value = req.Result
	}

	var call types.Call
	var err error
	if req.RequestType == "fluxmonitor" {
		call, err = NewFluxMonitorCall(adapter.client.meta, req)
	} else {
		call, err = NewCall(adapter.client.meta, req)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed creating call")
	}

	// Send the extrinsic
	return adapter.client.SubmitCall(call)
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

func NewFluxMonitorCall(m *types.Metadata, req Request) (types.Call, error) {
	i, err := strconv.ParseUint(req.FeedId, 10, 32)
	if err != nil {
		return types.Call{}, errors.Wrap(err, "failed parsing uint32")
	}
	feedID := types.NewUCompactFromUInt(i)

	i, err = strconv.ParseUint(req.RoundId, 10, 32)
	if err != nil {
		return types.Call{}, errors.Wrap(err, "failed parsing uint32")
	}
	roundID := types.NewUCompactFromUInt(i)

	num, err := ParseNumericString(req.Value)
	if err != nil {
		return types.Call{}, err
	}
	if req.Multiply != 0 {
		num = num.Mul(decimal.NewFromInt(req.Multiply))
	}
	value := types.NewUCompact(num.BigInt())

	return types.NewCall(m, "ChainlinkFeed.submit", feedID, roundID, value)
}
