package secp256k1

// Use the ethereum implementation of the secp256k1
// elliptic curve digital signature algorithm, which
// bridges to the C-implementation of Bitcoin

const (
	// The length of a secp256k1 signature.
	SignatureByteLength = 65
)
