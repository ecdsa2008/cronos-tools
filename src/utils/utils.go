package utils

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

func GetPrivateKey(mnemonic string, accountIndex uint) *ecdsa.PrivateKey {
	seed := bip39.NewSeed(mnemonic, "") // 可以提供密码短语
	masterKey, _ := bip32.NewMasterKey(seed)
	// 44'
	purposeKey, _ := masterKey.NewChildKey(0x8000002C)
	// 60'
	coinTypeKey, _ := purposeKey.NewChildKey(0x8000003C)
	// 0'
	accountKey, _ := coinTypeKey.NewChildKey(0x80000000)
	// 0
	changeKey, _ := accountKey.NewChildKey(0)
	// addressIndex
	addressKey, _ := changeKey.NewChildKey(uint32(accountIndex))
	privateKey, _ := crypto.ToECDSA(addressKey.Key)
	return privateKey
}

func GetPublicKey(privateKey *ecdsa.PrivateKey) *ecdsa.PublicKey {
	return &privateKey.PublicKey
}

func GetAddressFromPublicKey(publicKey *ecdsa.PublicKey) common.Address {
	return crypto.PubkeyToAddress(*publicKey)
}

func GetAddressFromPrivateKey(privateKey *ecdsa.PrivateKey) common.Address {
	return crypto.PubkeyToAddress(*GetPublicKey(privateKey))
}
