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

type ErrorKey struct {
	s string
}

func (e *ErrorKey) Error() string {
	return e.s
}

type PublicKey struct {
	*rsa.PublicKey
}

func (pub *PublicKey) MarshalJSON() ([]byte, error) {
	data := map[string]string{}
	data["algorithm"] = "RS"
	data["n"] = pub.N.String()
	data["e"] = strconv.Itoa(pub.E)
	return json.Marshal(data)
}

func LoadPublicKey(filename string) (*PublicKey, error) {
	pem_data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

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
	public_key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	if rsa_key, ok := public_key.(*rsa.PublicKey); ok {
		return &PublicKey{rsa_key}, nil
	}
	return nil, &ErrorKey{"Not an RSA public key"}
}

type PrivateKey struct {
	*rsa.PrivateKey
}

func (k *PrivateKey) Sign(data []byte) ([]byte, error) {
	hashFunc := crypto.SHA256
	h := hashFunc.New()
	h.Write(data)
	digest := h.Sum(nil)
	return rsa.SignPKCS1v15(rand.Reader, k.PrivateKey, hashFunc, digest)
}

func LoadPrivateKey(filename string) (*PrivateKey, error) {
	pem_data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

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
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return &PrivateKey{key}, nil
}

func CreateCertificate(private_key *PrivateKey, public_key *PublicKey, email string, cert_duration int, pubkey map[string]string, iss string) ([]byte, error) {
	token := jwt.New(jwt.GetSigningMethod("RS256"))
	token.Claims["iat"] = time.Now().Add(-time.Duration(10)*time.Second).Unix() * 1000
	token.Claims["exp"] = time.Now().Add(time.Duration(cert_duration)*time.Second).Unix() * 1000
	token.Claims["iss"] = iss
	token.Claims["public-key"] = pubkey
	token.Claims["principal"] = map[string]string{"email": email}

	tokenString, err := token.SignedString(private_key.PrivateKey)
	if err != nil {
		return nil, err
	}

	return []byte(tokenString), err
}
