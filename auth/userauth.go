package auth

import (
	"crypto/md5"
	"encoding/hex"
)

type UserAccount struct {
	Login    string
	Password string
}

func HashAuthData(login string, password string) UserAccount {
	hashLogin := md5.Sum([]byte(login))
	hashPassword := md5.Sum([]byte(password))
	user := UserAccount{
		Login:    hex.EncodeToString(hashLogin[:]),
		Password: hex.EncodeToString(hashPassword[:]),
	}
	return user
}
