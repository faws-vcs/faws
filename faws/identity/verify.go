package identity

import "crypto/ed25519"

const SignatureSize = ed25519.SignatureSize

type Signature [SignatureSize]byte

func Verify(id ID, signature *Signature, message []byte) bool {
	return ed25519.Verify(ed25519.PublicKey(id[:]), message, signature[:])
}

func Sign(pair *Pair, message []byte, signature *Signature) {
	ed_sign := ed25519.Sign(ed25519.PrivateKey(pair[:]), message)
	copy(signature[:], ed_sign)
}
