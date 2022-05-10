package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mergemock/types"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/prysmaticlabs/prysm/shared/bls/blst"
	bls "github.com/prysmaticlabs/prysm/shared/bls/common"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type testRelayBackend struct {
	*RelayBackend
}

func newTestRelay(t *testing.T) *testRelayBackend {
	relay, err := NewRelayBackend(logrus.New())
	if err != nil {
		t.Fatal("unable to create relay")
	}

	testRelay := testRelayBackend{relay}
	testRelay.engine.JwtSecretPath = newJwt(t)
	testRelay.engine.GenesisPath = newGenesis(t)
	return &testRelay
}

func (mr *testRelayBackend) testRequest(t *testing.T, method string, path string, payload any) *httptest.ResponseRecorder {
	var req *http.Request
	var err error

	if payload == nil {
		req, err = http.NewRequest(method, path, nil)
	} else {
		payloadBytes, err2 := json.Marshal(payload)
		require.NoError(t, err2)
		req, err = http.NewRequest(method, path, bytes.NewReader(payloadBytes))
	}

	require.NoError(t, err)
	rr := httptest.NewRecorder()
	mr.getRouter().ServeHTTP(rr, req)
	return rr
}

func newKeypair(t *testing.T) (pubkey []byte, privkey bls.SecretKey) {
	sk, err := blst.RandKey()
	if err != nil {
		t.Fatal("unable to generate bls key pair", err)
	}
	return sk.PublicKey().Marshal(), sk
}

func newJwt(t *testing.T) string {
	path := fmt.Sprintf("%s/jwt.hex", t.TempDir())
	jwt := []byte("ed6588309287e7dbbb0ca2ba8c8be6e6063a72dc0f2235999ee6a751e8459cbc")
	if err := os.WriteFile(path, jwt, 0644); err != nil {
		t.Fatal("unable to write tmp jwt file")
	}
	return path
}

func newGenesis(t *testing.T) string {
	path := fmt.Sprintf("%s/genesis.json", t.TempDir())
	genesis := core.DeveloperGenesisBlock(5, 30_000_000, common.Address{})
	genesis.Config.MergeForkBlock = common.Big0
	genesis.Config.TerminalTotalDifficulty = common.Big0
	buf, err := genesis.MarshalJSON()
	if err != nil {
		t.Fatal("cannot marshal tmp genesis")
	}
	if err := os.WriteFile(path, buf, 0644); err != nil {
		t.Fatal("unable to write tmp genesis file")
	}
	return path
}

func TestStatusEndpoint(t *testing.T) {
	relay := newTestRelay(t)
	rr := relay.testRequest(t, "GET", "/eth/v1/builder/status", nil)
	require.Equal(t, http.StatusOK, rr.Code)
}

func TestValidatorRegistration(t *testing.T) {
	relay := newTestRelay(t)
	pk, sk := newKeypair(t)

	var pubkey types.PublicKey
	pubkey.FromSlice(pk)
	require.Equal(t, pk[:], pubkey[:])

	msg := types.RegisterValidatorRequestMessage{
		FeeRecipient: types.Address{0x42},
		GasLimit:     15_000_000,
		Timestamp:    uint64(time.Now().Unix()),
		Pubkey:       pubkey,
	}
	root, err := msg.HashTreeRoot()
	require.NoError(t, err)

	// Success
	sig := sk.Sign(root[:]).Marshal()
	var signature types.Signature
	signature.FromSlice(sig)
	require.Equal(t, sig[:], signature[:])

	rr := relay.testRequest(t, "POST", "/eth/v1/builder/validators", types.RegisterValidatorRequest{
		Message:   &msg,
		Signature: signature,
	})
	require.Equal(t, http.StatusOK, rr.Code)

	// Invalid signature
	signature[len(signature)-1] = 0x00
	rr = relay.testRequest(t, "POST", "/eth/v1/builder/validators", types.RegisterValidatorRequest{
		Message:   &msg,
		Signature: signature,
	})
	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Equal(t, errInvalidSignature.Error()+"\n", rr.Body.String())
}

func TestGetHeader(t *testing.T) {
	ctx := context.Background()
	relay := newTestRelay(t)
	relay.engine.Run(ctx)
	pk, _ := newKeypair(t)
	parent := relay.engine.mockChain().CurrentHeader()
	parentHash := parent.Hash()

	if _, err := relay.engine.backend.ForkchoiceUpdatedV1(
		ctx,
		&types.ForkchoiceStateV1{
			HeadBlockHash:      parentHash,
			SafeBlockHash:      parentHash,
			FinalizedBlockHash: parentHash,
		},
		&types.PayloadAttributesV1{
			Timestamp:             parent.Time + 1,
			PrevRandao:            common.Hash{0x01},
			SuggestedFeeRecipient: common.Address{0x02},
		},
	); err != nil {
		t.Fatal("unable to initialize engine")
	}

	path := fmt.Sprintf("/eth/v1/builder/header/%d/%s/0x%x", 0, parentHash.Hex(), pk[:])
	fmt.Println("path", path)
	rr := relay.testRequest(t, "GET", path, nil)
	require.Equal(t, http.StatusOK, rr.Code)

	bid := new(types.GetHeaderResponse)
	err := json.Unmarshal(rr.Body.Bytes(), bid)
	require.NoError(t, err)

	require.Equal(t, parentHash[:], bid.Data.Message.Header.ParentHash[:], "didn't build on expected parent")
	ok, err := verifySignature(bid.Data.Message, relay.pk[:], bid.Data.Signature[:])
	require.NoError(t, err, "error verifying signature")
	require.True(t, ok, "bid signature not valid")

	require.Equal(t, pk, relay.latestPubkey[:])
}

func TestGetPayload(t *testing.T) {
	ctx := context.Background()
	relay := newTestRelay(t)
	relay.engine.Run(ctx)
	pk, sk := newKeypair(t)
	parent := relay.engine.mockChain().CurrentHeader()
	parentHash := parent.Hash()

	blockHash := types.Hash{0xa1}
	feeRecipient := types.Address{0xb1}

	// Set relay pubkey for signature validation
	relay.latestPubkey.FromSlice(pk)
	msg := &types.BlindedBeaconBlock{
		Slot:          1,
		ProposerIndex: 2,
		ParentRoot:    types.Root{0x03},
		StateRoot:     types.Root{0x04},
		Body: &types.BlindedBeaconBlockBody{
			Eth1Data: &types.Eth1Data{
				DepositRoot:  types.Root{0x05},
				DepositCount: 5,
				BlockHash:    types.Hash{0x06},
			},
			ProposerSlashings: []*types.ProposerSlashing{},
			AttesterSlashings: []*types.AttesterSlashing{},
			Attestations:      []*types.Attestation{},
			Deposits:          []*types.Deposit{},
			VoluntaryExits:    []*types.VoluntaryExit{},
			SyncAggregate: &types.SyncAggregate{
				CommitteeBits:      types.CommitteeBits{0x07},
				CommitteeSignature: types.Signature{0x08},
			},
			ExecutionPayloadHeader: &types.ExecutionPayloadHeader{
				ParentHash:       types.Hash(parentHash),
				FeeRecipient:     feeRecipient,
				StateRoot:        types.Root{0x09},
				ReceiptsRoot:     types.Root{0x0a},
				LogsBloom:        types.Bloom{0x0b},
				Random:           types.Hash{0x0c},
				BlockNumber:      5001,
				GasLimit:         5002,
				GasUsed:          5003,
				Timestamp:        5004,
				ExtraData:        types.Hash{0x0d},
				BaseFeePerGas:    types.IntToU256(123456789),
				BlockHash:        blockHash,
				TransactionsRoot: types.Root{0x0e},
			},
		},
	}

	// Success
	root, err := msg.HashTreeRoot()
	require.NoError(t, err)
	sig := sk.Sign(root[:]).Marshal()

	var signature types.Signature
	signature.FromSlice(sig)
	require.Equal(t, sig[:], signature[:])

	rr := relay.testRequest(t, "POST", "/eth/v1/builder/blinded_blocks", types.GetPayloadRequest{
		Message:   msg,
		Signature: signature,
	})
	require.Equal(t, http.StatusOK, rr.Code)
}
