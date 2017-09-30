package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
	"errors"
	"math/big"
)

type rsaKeyValue struct {
	XMLName  xml.Name `xml:"RSAKeyValue"`
	Modulus  string   `xml:"Modulus"`
	Exponent string   `xml:"Exponent"`
	P        string   `xml:"P"`
	Q        string   `xml:"Q"`
	DP       string   `xml:"DP"`
	DQ       string   `xml:"DQ"`
	InverseQ string   `xml:"InverseQ"`
	D        string   `xml:"D"`
}

func Base64ToInt(s string) (*big.Int, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	i := new(big.Int)
	i.SetBytes(data)
	return i, nil
}

type CryptoRsaHelp struct {
	privateKey *rsa.PrivateKey
}

func (rsahelp *CryptoRsaHelp) parxml(xmlstr string) (*rsaKeyValue, error) {
	v := &rsaKeyValue{}
	if err := xml.Unmarshal([]byte(xmlstr), v); err != nil {
		return nil, err
	}
	return v, nil
}

func (rsahelp *CryptoRsaHelp) InitPublicKeyPem(pembytes []byte) error {
	block, _ := pem.Decode(pembytes)
	if block == nil {
		return errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}
	rsahelp.privateKey = new(rsa.PrivateKey)
	rsahelp.privateKey.PublicKey = *(pubInterface.(*rsa.PublicKey))
	return nil
}

func (rsahelp *CryptoRsaHelp) InitPrivateKeyPem(pembytes []byte) error {
	block, _ := pem.Decode(pembytes)
	if block == nil {
		return errors.New("private key error!")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return err
	}
	rsahelp.privateKey = priv
	return nil
}

//使用.NET格式初始化
func (rsahelp *CryptoRsaHelp) InitPublicKeyXml(xmlstr string) error {
	v, err := rsahelp.parxml(xmlstr)
	if err != nil {
		return err
	}
	rsahelp.privateKey = new(rsa.PrivateKey)
	rsahelp.privateKey.PublicKey.N, err = Base64ToInt(v.Modulus)
	if err != nil {
		return err
	}
	var E *big.Int
	E, err = Base64ToInt(v.Exponent)
	if err != nil {
		return err
	}
	rsahelp.privateKey.PublicKey.E = int(E.Int64())
	return nil
}

//使用.NET格式初始化
func (rsahelp *CryptoRsaHelp) InitPrivateKeyXml(xmlstr string) error {
	v, err := rsahelp.parxml(xmlstr)
	if err != nil {
		return err
	}
	rsahelp.privateKey = new(rsa.PrivateKey)
	rsahelp.privateKey.Primes = make([]*big.Int, 2)
	rsahelp.privateKey.PublicKey.N, err = Base64ToInt(v.Modulus)
	if err != nil {
		return err
	}
	var E *big.Int
	E, err = Base64ToInt(v.Exponent)
	if err != nil {
		return err
	}
	rsahelp.privateKey.PublicKey.E = int(E.Int64())
	rsahelp.privateKey.D, err = Base64ToInt(v.D)
	if err != nil {
		return err
	}
	rsahelp.privateKey.Primes[0], err = Base64ToInt(v.P)
	if err != nil {
		return err
	}
	rsahelp.privateKey.Primes[1], err = Base64ToInt(v.Q)
	if err != nil {
		return err
	}
	rsahelp.privateKey.Precomputed.Dp, err = Base64ToInt(v.DP)
	if err != nil {
		return err
	}
	rsahelp.privateKey.Precomputed.Dq, err = Base64ToInt(v.DQ)
	if err != nil {
		return err
	}
	rsahelp.privateKey.Precomputed.Qinv, err = Base64ToInt(v.InverseQ)
	if err != nil {
		return err
	}
	return nil
}

func (rsahelp *CryptoRsaHelp) DecryptBase64(base64Str string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return "", err
	}
	return rsahelp.Decrypt(data)
}

func (rsahelp *CryptoRsaHelp) Decrypt(data []byte) (string, error) {
	var dedata []byte
	keySize := rsahelp.privateKey.PublicKey.N.BitLen() / 8
	datalen := len(data)
	frag := datalen / keySize
	left := datalen % keySize
	for i := 0; i < frag; i++ {
		out, err := rsa.DecryptPKCS1v15(rand.Reader, rsahelp.privateKey, data[i*keySize:(i+1)*keySize])
		if err != nil {
			return "", err
		}
		dedata = append(dedata, out...)
	}
	if left > 0 {
		out, err := rsa.DecryptPKCS1v15(rand.Reader, rsahelp.privateKey, data[frag*keySize:])
		if err != nil {
			return "", err
		}
		dedata = append(dedata, out...)
	}
	return string(dedata), nil
}

func (rsahelp *CryptoRsaHelp) EncryptBase64(ecStr string) (string, error) {
	data, err := rsahelp.Encrypt(ecStr)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func (rsahelp *CryptoRsaHelp) Encrypt(ecStr string) ([]byte, error) {
	var endata []byte
	data := []byte(ecStr)
	keySize := (rsahelp.privateKey.N.BitLen() / 8) - 11
	datalen := len(data)
	frag := datalen / keySize
	left := datalen % keySize
	for i := 0; i < frag; i++ {
		out, err := rsa.EncryptPKCS1v15(rand.Reader, &rsahelp.privateKey.PublicKey, data[i*keySize:(i+1)*keySize])
		if err != nil {
			return nil, err
		}
		endata = append(endata, out...)
	}
	if left > 0 {
		out, err := rsa.EncryptPKCS1v15(rand.Reader, &rsahelp.privateKey.PublicKey, data[frag*keySize:])
		if err != nil {
			return nil, err
		}
		endata = append(endata, out...)
	}
	return endata, nil
}
