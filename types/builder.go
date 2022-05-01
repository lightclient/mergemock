package types

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Generate SSZ encoding with the following:
// sszgen --path types --include ../go-ethereum/common/hexutil --objs Eth1Data,BeaconBlockHeader,SignedBeaconBlockHeader,ProposerSlashing,Checkpoint,AttestationData,IndexedAttestation,AttesterSlashing,Attestation,Deposit,VoluntaryExit,SyncAggregate,ExecutionPayloadHeaderV1,BlindedBeaconBlockBodyV1,BlindedBeaconBlockV1

type Eth1Data struct {
	DepositRoot  Root           `json:"depositRoot" ssz-size:"32"`
	DepositCount hexutil.Uint64 `json:"depositCount"`
	BlockHash    Hash           `json:"blockHash" ssz-size:"32"`
}

type BeaconBlockHeader struct {
	Slot          hexutil.Uint64 `json:"slot"`
	ProposerIndex hexutil.Uint64 `json:"proposerIndex"`
	ParentRoot    Root           `json:"parentRoot" ssz-size:"32"`
	StateRoot     Root           `json:"stateRoot" ssz-size:"32"`
	BodyRoot      Root           `json:"bodyRoot" ssz-size:"32"`
}

type SignedBeaconBlockHeader struct {
	Header    *BeaconBlockHeader `json:"message"`
	Signature Signature          `json:"signature" ssz-size:"96"`
}

type ProposerSlashing struct {
	A *SignedBeaconBlockHeader `json:"signedHeader1"`
	B *SignedBeaconBlockHeader `json:"signedHeader2"`
}

type Checkpoint struct {
	Epoch hexutil.Uint64 `json:"epoch"`
	Root  Root           `json:"root" ssz-size:"32"`
}

type AttestationData struct {
	Slot      hexutil.Uint64 `json:"slot"`
	Index     hexutil.Uint64 `json:"Index"`
	BlockRoot Root           `json:"beaconBlockRoot" ssz-size:"32"`
	Source    *Checkpoint    `json:"source"`
	Target    *Checkpoint    `json:"target"`
}

type IndexedAttestation struct {
	AttestingIndices []uint64         `json:"attestingIndices" ssz-max:"2048"` // MAX_VALIDATORS_PER_COMMITTEE
	Data             *AttestationData `json:"data"`
	Signature        Signature        `json:"signature" ssz-size:"96"`
}

type AttesterSlashing struct {
	A *IndexedAttestation `json:"attestation1"`
	B *IndexedAttestation `json:"attestation2"`
}

type Attestation struct {
	AggregationBits hexutil.Bytes    `json:"aggregationBits" ssz-max:"2048"` // MAX_VALIDATORS_PER_COMMITTEE
	Data            *AttestationData `json:"data"`
	Signature       Signature        `json:"signature" ssz-size:"96"`
}

type Deposit struct {
	Pubkey                PublicKey      `json:"pubkey" ssz-size:"48"`
	WithdrawalCredentials Hash           `json:"withdrawalCredentials" ssz-size:"32"`
	Amount                hexutil.Uint64 `json:"amount"`
	Signature             Signature      `json:"signature" ssz-size:"96"`
}

type VoluntaryExits struct {
	Epoch          hexutil.Uint64 `json:"epoch"`
	ValidatorIndex hexutil.Uint64 `json:"validatorIndex"`
}

type SyncAggregate struct {
	CommitteeBits      CommitteeBits `json:"syncCommitteeBits" ssz-size:"64"`
	CommitteeSignature Signature     `json:"syncCommitteeSignature" ssz-size:"96"`
}

type ExecutionPayloadHeaderV1 struct {
	ParentHash       Hash           `json:"parentHash" ssz-size:"32"`
	FeeRecipient     Address        `json:"feeRecipient" ssz-size:"20"`
	StateRoot        Root           `json:"stateRoot" ssz-size:"32"`
	ReceiptsRoot     Root           `json:"receiptsRoot" ssz-size:"32"`
	LogsBloom        Bloom          `json:"logsBloom" ssz-size:"256"`
	Random           Hash           `json:"prevRandao" ssz-size:"32"`
	Number           hexutil.Uint64 `json:"blockNumber"`
	GasLimit         hexutil.Uint64 `json:"gasLimit"`
	GasUsed          hexutil.Uint64 `json:"gasUsed"`
	Timestamp        hexutil.Uint64 `json:"timestamp"`
	ExtraData        Hash           `json:"extraData" ssz-size:"32"`
	BaseFeePerGas    Hash           `json:"baseFeePerGas" ssz-max:"32"` // TODO should be actual u256
	BlockHash        Hash           `json:"blockHash" ssz-size:"32"`
	TransactionsRoot Root           `json:"transactionsRoot" ssz-size:"32"`
}

type BlindedBeaconBlockBodyV1 struct {
	RandaoReveal           Signature                 `json:"randaoReveal" ssz-size:"96"`
	Eth1Data               *Eth1Data                 `json:"eth1Data"`
	Graffiti               Hash                      `json:"graffiti" ssz-size:"32"`
	ProposerSlashings      []*ProposerSlashing       `json:"proposerSlashings" ssz-max:"16"`
	AttesterSlashings      []*AttesterSlashing       `json:"attesterSlashings" ssz-max:"2"`
	Attestations           []*Attestation            `json:"attestations" ssz-max:"128"`
	Deposits               []*Deposit                `json:"deposits" ssz-max:"4"`
	VoluntaryExits         []*VoluntaryExits         `json:"voluntaryExits" ssz-max:"16"`
	SyncAggregate          *SyncAggregate            `json:"syncAggregate"`
	ExecutionPayloadHeader *ExecutionPayloadHeaderV1 `json:"executionPayloadHeader"`
}

type BlindedBeaconBlockV1 struct {
	Slot          hexutil.Uint64            `json:"slot"`
	ProposerIndex hexutil.Uint64            `json:"proposerIndex"`
	ParentRoot    Root                      `json:"parentRoot" ssz-size:"32"`
	StateRoot     Root                      `json:"stateRoot" ssz-size:"32"`
	Body          *BlindedBeaconBlockBodyV1 `json:"body"`
}

type RegisterValidatorRequestMessage struct {
	FeeRecipient Address        `json:"feeRecipient" ssz-size:"20"`
	GasTarget    hexutil.Uint64 `json:"gasTarget"`
	Timestamp    hexutil.Uint64 `json:"timestamp"`
	Pubkey       PublicKey      `json:"pubkey" ssz-size:"48"`
}

type BuilderBidV1 struct {
	Header *ExecutionPayloadHeaderV1 `json:"header"`
	Value  hexutil.Uint64            `json:"value"` // TODO: make uint256
	Pubkey PublicKey                 `json:"pubkey" ssz-size:"48"`
}

type SignedBuilderBidV1 struct {
	Message   *BuilderBidV1 `json:"message"`
	Signature Signature     `json:"signature"`
}

func PayloadToPayloadHeader(p *ExecutionPayloadV1) (*ExecutionPayloadHeaderV1, error) {
	// txs, err := decodeTransactions(p.Transactions)
	// if err != nil {
	//         return nil, err
	// }
	return &ExecutionPayloadHeaderV1{
		// ParentHash:       p.ParentHash,
		// FeeRecipient:     p.FeeRecipient,
		// StateRoot:        p.StateRoot,
		// ReceiptsRoot:     p.ReceiptsRoot,
		// LogsBloom:        p.LogsBloom,
		// PrevRandao:       p.Random,
		// BlockNumber:      p.Number,
		// GasLimit:         p.GasLimit,
		// GasUsed:          p.GasUsed,
		// Timestamp:        p.Timestamp,
		// ExtraData:        p.ExtraData,
		// BaseFeePerGas:    (*big.Int)(p.BaseFeePerGas),
		// BlockHash:        p.BlockHash,
		// TransactionsRoot: types.DeriveSha(types.Transactions(txs), trie.NewStackTrie(nil)),
	}, nil
}
