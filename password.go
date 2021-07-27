package wilhelmiina

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	PW_SALT_BYTES  = 128
	PW_HASH_LEN    = 512
	ARGON2_TIME    = 5
	ARGON2_MEM     = 1024 * 1024 * 1
	ARGON2_THREADS = 4
)

func genSalt() (salt []byte, err error) {
	salt = make([]byte, PW_SALT_BYTES)
	_, err = io.ReadFull(rand.Reader, salt)
	if err != nil {
		return nil, err
	}
	return
}

func hashPassword(password string, salt []byte) []byte {
	hash := argon2.IDKey([]byte(password), salt, ARGON2_TIME, ARGON2_MEM, ARGON2_THREADS, PW_HASH_LEN)
	return hash

}

func encodeHash(hash []byte, salt []byte) (encoded string) {
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)

	encoded = fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, ARGON2_MEM, ARGON2_TIME, ARGON2_THREADS, b64Salt, b64Hash,
	)

	return
}

func validatePassword(password string, hashString string) (bool, error) {
	hash, salt, err := decodeHash(hashString)
	if err != nil {
		return false, err
	}

	hash2 := hashPassword(password, salt)
	return subtle.ConstantTimeCompare(hash, hash2) == 1, nil
}

func decodeHash(hashString string) (hash []byte, salt []byte, err error) {
	vals := strings.Split(hashString, "$")
	b64salt := vals[4]
	b64hash := vals[5]

	hash, err = base64.RawStdEncoding.Strict().DecodeString(b64hash)
	if err != nil {
		return nil, nil, err
	}
	salt, err = base64.RawStdEncoding.Strict().DecodeString(b64salt)
	if err != nil {
		return nil, nil, err
	}
	return
}

func genHashString(password string) (string, error) {
	salt, err := genSalt()
	if err != nil {
		return "", err
	}
	hash := hashPassword(password, salt)
	encoded := encodeHash(hash, salt)
	return encoded, nil
}
