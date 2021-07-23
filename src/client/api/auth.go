package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"strconv"
	"time"
)

type Auth struct {
	APIKey    string
	APISecret string
}

func (auth *Auth) Authentication(endPoint string, postData string) (nonce string, authent string) {
	nonce = strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	input := postData + nonce + endPoint
	hash := sha256.Sum256([]byte(input))
	macKey, _ := base64.StdEncoding.DecodeString(auth.APISecret)
	mac := hmac.New(sha512.New, macKey)
	mac.Write(hash[:])
	authent = base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return
}
