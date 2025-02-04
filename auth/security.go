package auth

import (
	"crypto/hmac"
	"crypto/sha256"
    "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

// Password hashing using bcrypt
const BCRYPT_COST int = 8
const DefaultExpiration = 1 * time.Minute

// TODO: replace with random strings
var Pepper = []byte("lucy-is-a-good-kitty")
var Secret = "lucy-is-a-good-cat"

func GenUUID() (string, error) {
    uuidBytes := make([]byte, 16)
    _, err := rand.Read(uuidBytes)
    if err != nil {
        return "", err
    }

    uuidBytes[6] = (uuidBytes[6] & 0x0f) | 0x40
    uuidBytes[8] = (uuidBytes[8] & 0x3f) | 0x80

    uuidStr := fmt.Sprintf("%x-%x-%x-%x-%x",
        uuidBytes[0:4],
        uuidBytes[4:6],
        uuidBytes[6:8],
        uuidBytes[8:10],
        uuidBytes[10:16])

    return uuidStr, nil
}

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

func NewPayload(sub string) Payload {
	return Payload{
		Sub: sub,
		Iat: time.Now().Unix(),
		Exp: time.Now().Add(DefaultExpiration).Unix(),
	}
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
