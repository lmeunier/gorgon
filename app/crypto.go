package app

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"strconv"
	"time"
)

// ErrorKey is a base error for public/private key operations
type ErrorKey struct {
	s string // error message
}

// Error returns the error message
func (e *ErrorKey) Error() string {
	return e.s
}

// PublicKey represents an RSA public key that implements the Marshaler interface.
type PublicKey struct {
	*rsa.PublicKey // anonymous field to the real RSA public key
}

// MarshalJSON returns the json representation of the RSA public key
func (pub *PublicKey) MarshalJSON() ([]byte, error) {
	data := map[string]string{}
	data["algorithm"] = "RS"
	data["n"] = pub.N.String()
	data["e"] = strconv.Itoa(pub.E)
	return json.Marshal(data)
}

// LoadPublicKey returns a PublicKey created from the content of a PEM encoded
// file containing an RSA public key.
func LoadPublicKey(filename string) (*PublicKey, error) {
	// load content from filename
	pem_data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// iterate an all blocks in the file and search a "PUBLIC KEY" block
	var block *pem.Block
	rest := []byte(pem_data)
	for {
		block, rest = pem.Decode(rest)
		if block == nil {
			return nil, &ErrorKey{"No PUBLIC KEY bloc found"}
		}
		if block.Type == "PUBLIC KEY" {
			break
		}
	}

	// we have a "PUBLIC KEY" block, parse it to create an public key
	public_key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	// cast the public key as an RSA public key
	if rsa_key, ok := public_key.(*rsa.PublicKey); ok {
		return &PublicKey{rsa_key}, nil
	}
	return nil, &ErrorKey{"Not an RSA public key"}
}

// PrivateKey represents an RSA private key.
type PrivateKey struct {
	*rsa.PrivateKey // anonymous field to the read RSA private key
}

// Sign returns the signature of the SHA256 hash of the given data.
func (k *PrivateKey) Sign(data []byte) ([]byte, error) {
	hashFunc := crypto.SHA256
	h := hashFunc.New()
	h.Write(data)
	digest := h.Sum(nil)
	return rsa.SignPKCS1v15(rand.Reader, k.PrivateKey, hashFunc, digest)
}

// LoadPrivateKey returns a PrivateKey created from the content of a PEM
// encoded file containing an RSA private key.
func LoadPrivateKey(filename string) (*PrivateKey, error) {
	// load content from filename
	pem_data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// iterate an all blocks in the file and search a "RSA PRIVATE KEY" block
	var block *pem.Block
	rest := []byte(pem_data)
	for {
		block, rest = pem.Decode(rest)
		if block == nil {
			return nil, &ErrorKey{"No RSA PRIVATE KEY bloc found"}
		}
		if block.Type == "RSA PRIVATE KEY" {
			break
		}
	}

	// we have a "RSA PRIVATE KEY" block, parse it to create an RSA private key
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return &PrivateKey{key}, nil
}

// CreateCertificate returns the string representation of a token signed with
// the given private_key. The token contains the following claims:
// - iat:
// - exp: expiry date of the certificate
// - iss: issuer of the certificate (the domain used by the IdP)
// - public-key: the public key provided by the browser
// - principal:
//   - email : the email address of the authenticated user
func CreateCertificate(private_key *PrivateKey, public_key *PublicKey, email string, cert_duration time.Duration, pubkey map[string]string, iss string) ([]byte, error) {
	// create a new JSON Web Token
	token := jwt.New(jwt.GetSigningMethod("RS256"))
	token.Claims["iat"] = time.Now().Add(-time.Duration(10)*time.Second).Unix() * 1000
	token.Claims["exp"] = time.Now().Add(cert_duration).Unix() * 1000
	token.Claims["iss"] = iss
	token.Claims["public-key"] = pubkey
	token.Claims["principal"] = map[string]string{"email": email}

	// sign the token with the private key
	tokenString, err := token.SignedString(private_key.PrivateKey)
	if err != nil {
		return nil, err
	}

	// returns the strign representation of the signed token
	return []byte(tokenString), err
}
