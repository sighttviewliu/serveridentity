package identity

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"

	ed "github.com/FactomProject/ed25519"
	"github.com/FactomProject/factom"
	"github.com/FactomProject/factomd/common/primitives"
)

type BlockSigningKey struct {
	version          []byte
	message          []byte
	rootChainID      []byte
	newPubKey        []byte
	timestamp        []byte
	identityPreimage []byte
	signiture        []byte

	subchain string
}

func MakeBlockSigningKeySeeded(rootChainIDStr string, subchainID string, privateKey *[64]byte, signingkey *[64]byte) (*BlockSigningKey, []byte, error) {
	block, _, err := MakeBlockSigningKeyFixed(rootChainIDStr, subchainID, privateKey, false)
	if err != nil {
		return nil, nil, err
	}

	block.newPubKey = signingkey[32:]

	t := primitives.NewTimestampNow().GetTimeSeconds()
	by := make([]byte, 8)
	binary.BigEndian.PutUint64(by, uint64(t))
	block.timestamp = by

	preI := make([]byte, 0)
	preI = append(preI, []byte{0x01}...)
	preI = append(preI, privateKey[32:]...)
	block.identityPreimage = preI

	sig := ed.Sign(privateKey, block.versionToTimestamp())
	block.signiture = sig[:]

	return block, signingkey[:32], nil
}

func MakeBlockSigningKey(rootChainIDStr string, subchainID string, privateKey *[64]byte) (*BlockSigningKey, []byte, error) {
	a, b, c := MakeBlockSigningKeyFixed(rootChainIDStr, subchainID, privateKey, false)
	return a, b, c
}

// Creates a new BlockSigningKey type. Used to change keys in identity chain
func MakeBlockSigningKeyFixed(rootChainIDStr string, subchainID string, privateKey *[64]byte, random bool) (*BlockSigningKey, []byte, error) {
	rootChainID, err := hex.DecodeString(rootChainIDStr)
	if err != nil {
		return nil, nil, err
	}
	if bytes.Compare(rootChainID[:ProofOfWorkLength], ProofOfWorkChainID[:ProofOfWorkLength]) != 0 {
		return nil, nil, errors.New("Error making a new block signing key: Root ChainID invalid")
	}
	b := new(BlockSigningKey)
	b.version = []byte{0x00}
	b.message = []byte("New Block Signing Key")
	b.rootChainID = rootChainID
	b.subchain = subchainID

	pub, priv, err := ed.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	if !random {
		buf := new(bytes.Buffer)
		buf.WriteString(rootChainIDStr)
		pub, priv, err = ed.GenerateKey(buf)
		if err != nil {
			return nil, nil, err
		}
	}
	b.newPubKey = pub[:32]
	t := primitives.NewTimestampNow().GetTimeSeconds()
	by := make([]byte, 8)
	binary.BigEndian.PutUint64(by, uint64(t))
	b.timestamp = by

	preI := make([]byte, 0)
	preI = append(preI, []byte{0x01}...)
	preI = append(preI, privateKey[32:]...)
	b.identityPreimage = preI

	sig := ed.Sign(privateKey, b.versionToTimestamp())
	b.signiture = sig[:]
	return b, priv[:], nil
}

func (b *BlockSigningKey) GetEntry() *factom.Entry {
	e := new(factom.Entry)
	e.ChainID = b.subchain
	e.Content = []byte{}
	e.ExtIDs = b.extIdList()

	return e
}

func (b *BlockSigningKey) versionToTimestamp() []byte {
	buf := new(bytes.Buffer)
	buf.Write(b.version)
	buf.Write(b.message)
	buf.Write(b.rootChainID)
	buf.Write(b.newPubKey)
	buf.Write(b.timestamp)

	return buf.Bytes()
}

func (b *BlockSigningKey) extIdList() [][]byte {
	list := make([][]byte, 0)
	list = append(list, b.version)
	list = append(list, b.message)
	list = append(list, b.rootChainID)
	list = append(list, b.newPubKey)
	list = append(list, b.timestamp)
	list = append(list, b.identityPreimage)
	list = append(list, b.signiture)

	return list
}
