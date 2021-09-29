package test

import (
	"crypto/sha256"
	"math/big"
	"math/rand"
	"time"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
	"github.com/keep-network/keep-core/pkg/altbn128"
)

func GenRandomBytes(size int) (blk []byte) {
	rand.Seed(time.Now().UnixNano())
	blk = make([]byte, size)
	_, _ = rand.Read(blk)
	return
}

func GenRandomKeys(total int) ([]*big.Int, []*bn256.G2) {
	privs, pubs := make([]*big.Int, total), make([]*bn256.G2, total)
	for i := 0; i < total; i++ {
		privs[i] = new(big.Int).SetBytes(GenRandomBytes(64))
		pubs[i] = new(bn256.G2).ScalarBaseMult(privs[i])
	}
	return privs, pubs
}

func CalculateAntiRogueCoefficients(pubs []*bn256.G2) []*big.Int {
	as := make([]*big.Int, len(pubs))
	data := pubs[0].Marshal()
	for i := 0; i < len(pubs); i++ {
		data = append(data, pubs[i].Marshal()...)
	}

	for i := 0; i < len(pubs); i++ {
		cur := pubs[i].Marshal()
		copy(data[0:len(cur)], cur)
		hash := sha256.Sum256(data)
		as[i] = new(big.Int).SetBytes(hash[:])
	}
	return as
}

func AggregateMembershipKeys(privs []*big.Int, pubs []*bn256.G2, aggPub *bn256.G2, coefs []*big.Int) []*bn256.G1 {
	res := make([]*bn256.G1, len(pubs))
	for i := 0; i < len(pubs); i++ {
		res[i] = new(bn256.G1).ScalarMult(HashToPointByte(aggPub, byte(i)), privs[0])
		res[i] = new(bn256.G1).ScalarMult(res[i], coefs[0])
		for j := 1; j < len(pubs); j++ {
			tmp := new(bn256.G1).ScalarMult(HashToPointByte(aggPub, byte(i)), privs[j])
			tmp = new(bn256.G1).ScalarMult(tmp, coefs[j])
			res[i] = new(bn256.G1).Add(res[i], tmp)
		}
	}
	return res
}

// Sign creates a point on a curve G1 by hashing and signing provided message
// using the provided secret key.
func Sign(secretKey *big.Int, message []byte) *bn256.G1 {
	return SignG1(secretKey, altbn128.G1HashToPoint(message))
}

// SignG1 creates a point on a curve G1 by signing the provided
// G1 point message using the provided secret key.
func SignG1(secretKey *big.Int, message *bn256.G1) *bn256.G1 {
	return new(bn256.G1).ScalarMult(message, secretKey)
}

// Verify performs the pairing operation to check if the signature is correct
// for the provided message and the corresponding public key.
func Verify(publicKey *bn256.G2, message []byte, signature *bn256.G1) bool {
	return VerifyG1(publicKey, altbn128.G1HashToPoint(message), signature)
}

// VerifyG1 performs the pairing operation to check if the signature is correct
// for the provided G1 point message and the corresponding public key.
func VerifyG1(publicKey *bn256.G2, message *bn256.G1, signature *bn256.G1) bool {
	// Generator point of G2 group.
	p2 := new(bn256.G2).ScalarBaseMult(big.NewInt(1))

	a := []*bn256.G1{new(bn256.G1).Neg(signature), message}
	b := []*bn256.G2{p2, publicKey}

	return bn256.PairingCheck(a, b)
}

func HashToPointMsg(p *bn256.G2, message []byte) *bn256.G1 {
	var data []byte
	data = append(data, p.Marshal()...)
	data = append(data, message...)
	return altbn128.G1HashToPoint(data)
}

func HashToPointByte(p *bn256.G2, index byte) *bn256.G1 {
	data := make([]byte, 32)
	data[31] = index
	return HashToPointMsg(p, data)
}

func SignAggregated(secretKey *big.Int, message []byte, publicKey *bn256.G2, membershipKey *bn256.G1) *bn256.G1 {
	plainSig := new(bn256.G1).ScalarMult(HashToPointMsg(publicKey, message), secretKey)
	return new(bn256.G1).Add(plainSig, membershipKey)
}

func AggregatePointsOnG2(points []*bn256.G2, coefs []*big.Int) *bn256.G2 {
	res := new(bn256.G2).ScalarMult(points[0], coefs[0])
	for i := 1; i < len(points); i++ {
		cur := new(bn256.G2).ScalarMult(points[i], coefs[i])
		res = new(bn256.G2).Add(res, cur)
	}
	return res
}

func VerifyAggregated(aggregatedPublicKey *bn256.G2, partPublicKey *bn256.G2, message []byte, partSignature *bn256.G1, bitmask *big.Int) bool {
	// Generator point of G2 group.
	g2 := new(bn256.G2).ScalarBaseMult(big.NewInt(1))
	g1 := new(bn256.G1).ScalarBaseMult(big.NewInt(1))

	sum := new(bn256.G1).Add(g1, new(bn256.G1).Neg(g1))
	for index := 0; bitmask.Sign() != 0; index++ {
		if bitmask.Bit(index) != 0 {
			bitmask = new(big.Int).SetBit(bitmask, index, 0)
			sum = new(bn256.G1).Add(sum, HashToPointByte(aggregatedPublicKey, byte(index)))
		}
	}

	a := []*bn256.G1{new(bn256.G1).Neg(partSignature), HashToPointMsg(aggregatedPublicKey, message), sum}
	b := []*bn256.G2{g2, partPublicKey, aggregatedPublicKey}

	return bn256.PairingCheck(a, b)
}
