package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

// Password hashing using bcrypt
const BCRYPT_COST int = 8
const DefaultExpirationJWT = 5 * time.Minute
const DefaultExpirationRefresh = 24 * time.Hour

// TODO: replace with random strings
var Pepper = []byte("lucy-is-a-good-kitty")
var Secret = "lucy-is-a-good-cat"
var RefreshSecret = []byte("lucy-is-naughty-sometimes")



func HashPassword(password string) (string, error) {

	pepperPW := append([]byte(password), Pepper...)
	hashedPW, err := bcrypt.GenerateFromPassword(pepperPW, BCRYPT_COST)
	if err != nil {
        return "", err
    }
    return string(hashedPW), nil
}

func VerifyPassword(hashedPW string, password string) bool {
	pepperPW := append([]byte(password), Pepper...)
	return bcrypt.CompareHashAndPassword([]byte(hashedPW), pepperPW) == nil
}

// defining JWT for auth to be stored in secure cookies
type JWT struct {
	Header PayloadHeader `json:"header"`
	Payload Payload		 `json:"payload"`
	Signature string	 `json:"signature"`
}

type PayloadHeader struct {
	Alg string 	`json:"alg"`
	Typ string  `json:"typ"`
}

type Payload struct {
	Sub string  `json:"sub"`
	Iat int64 	`json:"iat"`
	Exp int64	`json:"exp"`
}

type RefreshToken struct {
    SessionID string `json:"sid"`
    ExpiresAt int64  `json:"exp"`
}

func NewPayload(sub string) Payload {
	return Payload{
		Sub: sub,
		Iat: time.Now().Unix(),
		Exp: time.Now().Add(DefaultExpirationRefresh).Unix(),
	}
}

func CreateRefreshToken(sessionID string, secret []byte) (string, error) {
    refreshToken := &RefreshToken{
        SessionID: sessionID,
        ExpiresAt: time.Now().Add(DefaultExpirationRefresh).Unix(),
    }

    tokenJSON, err := json.Marshal(refreshToken)
    if err != nil {
        return "", err
    }

    h := hmac.New(sha256.New, secret)
    h.Write(tokenJSON)
    signature := h.Sum(nil)

    tokenBase64 := base64.RawURLEncoding.EncodeToString(tokenJSON)
    signatureBase64 := base64.RawURLEncoding.EncodeToString(signature)

    return fmt.Sprintf("%s.%s", tokenBase64, signatureBase64), nil
}

func VerifyRefreshToken(token string, secret []byte) (*RefreshToken, error) {
    parts := strings.Split(token, ".")
    if len(parts) != 2 {
        return nil, fmt.Errorf("invalid token format")
    }

    tokenJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
    if err != nil {
        return nil, err
    }

    signature, err := base64.RawURLEncoding.DecodeString(parts[1])
    if err != nil {
        return nil, err
    }

    h := hmac.New(sha256.New, secret)
    h.Write(tokenJSON)
    expectedSignature := h.Sum(nil)

    if !hmac.Equal(signature, expectedSignature) {
        return nil, fmt.Errorf("invalid signature")
    }

    var refreshToken RefreshToken
    if err := json.Unmarshal(tokenJSON, &refreshToken); err != nil {
        return nil, err
    }

    if time.Now().Unix() > refreshToken.ExpiresAt {
        return nil, fmt.Errorf("token expired")
    }

    return &refreshToken, nil
}

func ToJSON(data any) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
        return "", err
    }
    return base64.RawURLEncoding.EncodeToString(jsonData), nil
}

func SignPayload(secret string, payload Payload) (string, error) {
	header := PayloadHeader{Alg: "HS256", Typ: "JWT"}

	headerJSON, err := ToJSON(header)
	if err != nil {
        return "", err
    }

    payloadJSON, err := ToJSON(payload)
    if err != nil {
        return "", err
    }
    unsignedToken  := fmt.Sprintf("%s.%s", headerJSON, payloadJSON)

    h := hmac.New(sha256.New, []byte(secret))
    h.Write([]byte(unsignedToken))
    signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

    jwt := fmt.Sprintf("%s.%s.%s", headerJSON, payloadJSON, signature)
    return jwt, nil
}

// TODO: here I can also implement a refresh token?
func VerifyPayload(secret, token string) (*Payload, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("Invalid token format!")
	}

	header := parts[0]
	payload:= parts[1]
	signature := parts[2]

	unsignedToken := fmt.Sprintf("%s.%s", header, payload)

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(unsignedToken))
	expectedSig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	if expectedSig != signature {
		return nil, fmt.Errorf("Invalid signature!")
	}

	jsonPayload, err :=  base64.RawURLEncoding.DecodeString(payload)
    if err != nil {
        return nil, err
    }

    var payloadData Payload
    if err := json.Unmarshal(jsonPayload, &payloadData)
    err != nil {
            return nil, err
    }

    if time.Now().Unix() > payloadData.Exp {
    	return nil, fmt.Errorf("Token expired!")
    }
    return &payloadData, nil
}



