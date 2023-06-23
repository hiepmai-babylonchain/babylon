package types_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/cometbft/cometbft/crypto/tmhash"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/stretchr/testify/require"
)

func newInvalidPoP(r *rand.Rand, babylonSK cryptotypes.PrivKey, btcSK *btcec.PrivateKey) *types.ProofOfPossession {
	pop := types.ProofOfPossession{}

	randomNum := datagen.RandomInt(r, 2) // 0 or 1

	btcPK := btcSK.PubKey()
	bip340PK := bbn.NewBIP340PubKeyFromBTCPK(btcPK)
	babylonSig, err := babylonSK.Sign(*bip340PK)
	if err != nil {
		panic(err)
	}

	var babylonSigHash []byte
	if randomNum == 0 {
		pop.BabylonSig = babylonSig                        // correct sig
		babylonSigHash = datagen.GenRandomByteArray(r, 32) // fake sig hash
	} else {
		pop.BabylonSig = datagen.GenRandomByteArray(r, uint64(len(babylonSig))) // fake sig
		babylonSigHash = tmhash.Sum(pop.BabylonSig)                             // correct sig hash
	}

	btcSig, err := schnorr.Sign(btcSK, babylonSigHash)
	if err != nil {
		panic(err)
	}
	bip340Sig := bbn.NewBIP340SignatureFromBTCSig(btcSig)
	pop.BtcSig = &bip340Sig

	return &pop
}

func FuzzPoP(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// generate BTC key pair
		btcSK, btcPK, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		bip340PK := bbn.NewBIP340PubKeyFromBTCPK(btcPK)

		// generate Babylon key pair
		babylonSK, babylonPK, err := datagen.GenRandomSecp256k1KeyPair(r)
		require.NoError(t, err)

		// generate and verify PoP, correct case
		pop, err := types.NewPoP(babylonSK, btcSK)
		require.NoError(t, err)
		err = pop.Verify(babylonPK, bip340PK)
		require.NoError(t, err)

		// generate and verify PoP, invalid case
		invalidPoP := newInvalidPoP(r, babylonSK, btcSK)
		err = invalidPoP.Verify(babylonPK, bip340PK)
		require.Error(t, err)
	})
}
