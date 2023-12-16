/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cobra

import (
	"context"
	"cronos-tools/src/utils"
	"encoding/hex"
	"errors"
	"fmt"
	_ "github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
	"log"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// mintCmd represents the mint command
var mintCmd = &cobra.Command{
	Use:   "mint",
	Short: "Auto mint inscriptions through mnemonic with multi bip-44 sequence addresses",
	Long:  `Auto mint inscriptions through mnemonic with multi bip-44 sequence addresses, you must support enough native coin to pay for gas fee`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("mint called")
		mnemonic, err := cmd.Flags().GetString("mnemonic")
		if err != nil {
			log.Panicln(errors.New("mnemonic is required"))
		}
		if mnemonic == "" {
			log.Panicln(errors.New("mnemonic is required"))
		}
		rpc, err := cmd.Flags().GetString("rpc")
		if err != nil {
			log.Panicln(errors.New("rpc is required"))
		}
		if rpc == "" {
			log.Panicln(errors.New("rpc is required"))
		}

		startIndex, err := cmd.Flags().GetUint("start-index")
		if err != nil {
			log.Panicln(errors.New("start-index is required"))
		}
		endIndex, err := cmd.Flags().GetUint("end-index")
		if err != nil {
			log.Panicln(errors.New("end-index is required"))
		}

		if startIndex > endIndex {
			log.Panicln(errors.New("start-index must less than or equal to end-index"))
		}

		hexContent, err := cmd.Flags().GetString("hex-content")
		if err != nil {
			log.Panicln(errors.New("hex-content is required"))
		}
		hexContent = strings.TrimPrefix(hexContent, "0x")
		textContent, err := cmd.Flags().GetString("text-content")
		if err != nil {
			log.Panicln(errors.New("text-content is required"))
		}
		if hexContent == "" && textContent == "" {
			log.Panicln(errors.New("hex-content or text-content is required"))
		}
		useHexContent := hexContent != ""

		perAddressMinted, err := cmd.Flags().GetUint("per-address-minted")
		if err != nil {
			log.Panicln(errors.New("per-address-minted is required"))
		}
		if perAddressMinted == 0 {
			log.Panicln(errors.New("per-address-minted must bigger than 0"))
		}

		client, err := ethclient.Dial(rpc)
		if err != nil {
			log.Panicln(err)
		}

		networkID, err := client.NetworkID(context.Background())
		if err != nil {
			log.Panicln(err)
		}
		gasLimit := uint64(21944)

	batchMint:
		for i := startIndex; i <= endIndex; i++ {
			// 获取当前账户的私钥
			accountPrivateKey := utils.GetPrivateKey(mnemonic, i)
			// 获取当前账户的地址
			accountAddress := utils.GetAddressFromPrivateKey(accountPrivateKey)
			// 获取当前账户的nonce
			localNonce, err := client.PendingNonceAt(context.Background(), accountAddress)
			if err != nil {
				log.Println("Can not get nonce ", err)
				// 如果获取nonce失败，则等待10秒后
				for j := 0; j < 5; j++ {
					time.Sleep(10 * time.Second)
					localNonce, err = client.PendingNonceAt(context.Background(), accountAddress)
					if err == nil {
						continue
					}
				}
				if err != nil {
					log.Panicln("Can not get nonce after retry 5 times ", err)
				}
			}
			for j := uint(0); j < perAddressMinted; j++ {
				// 获取当前账户的gasPrice
				gasPrice, err := client.SuggestGasPrice(context.Background())
				if err != nil {
					log.Println("Can not get gas price ", err)
					// 如果获取gasPrice失败，则等待10秒后
					for j := 0; j < 5; j++ {
						time.Sleep(10 * time.Second)
						gasPrice, err = client.SuggestGasPrice(context.Background())
						if err == nil {
							continue
						}
					}
					if err != nil {
						log.Panicln("Can not get gas price after retry 5 times ", err)
					}
				}
				bufferedGasPrice := decimal.NewFromBigInt(gasPrice, 0).Mul(decimal.NewFromFloat32(1)).BigInt()
				// 构造payload
				var payload []byte
				if useHexContent {
					payload, err = hex.DecodeString(hexContent)
					if err != nil {
						log.Panicln(err)
					}
				} else {
					payload = []byte(textContent)
				}

				// 检查当前账户的native coin余额是否足够支付gas fee
				balance, err := client.BalanceAt(context.Background(), accountAddress, nil)
				if err != nil {
					log.Println("Can not get balance ", err)
					// 如果获取balance失败，则等待10秒后
					for j := 0; j < 5; j++ {
						time.Sleep(10 * time.Second)
						balance, err = client.BalanceAt(context.Background(), accountAddress, nil)
						if err == nil {
							continue
						}
					}
					if err != nil {
						log.Panicln("Can not get balance after retry 5 times ", err)
					}
				}

				// 计算gas fee
				gasFee := decimal.NewFromBigInt(bufferedGasPrice, 0).Mul(decimal.NewFromInt(int64(gasLimit))).BigInt()
				if balance.Cmp(gasFee) < 0 {
					log.Println("account " + accountAddress.Hex() + " balance is not enough to pay for gas fee")
					log.Println("Switch to next account")
					break
				}

				// 构造交易
				tx := types.NewTx(&types.LegacyTx{
					Nonce:    localNonce,
					To:       &accountAddress,
					Value:    decimal.Zero.BigInt(),
					Gas:      gasLimit,
					GasPrice: bufferedGasPrice,
					Data:     payload,
				})
				// 签名交易
				signedTx, err := types.SignTx(tx, types.NewEIP155Signer(networkID), accountPrivateKey)
				if err != nil {
					log.Panicln("Can not sign transaction ", err)
				}
				// 发送交易
				err = client.SendTransaction(context.Background(), signedTx)
				if err != nil {
					log.Panicln(err)
				}
				txHash := signedTx.Hash()
				txHashString := txHash.Hex()

				log.Println("Account index: ", i, " Address: ", accountAddress.Hex(), " Tx hash: ", txHashString, " Payload: ", string(payload))

				time.Sleep(3 * time.Second)
				localNonce++
				retryTimes := 0
				maxRetryTimes := 10
				for {
					remoteNonce, err := client.PendingNonceAt(context.Background(), accountAddress)
					if err != nil {
						log.Panicln(err)
					}
					if remoteNonce == localNonce {
						break
					} else {
						time.Sleep(5 * time.Second)
						retryTimes++
						if retryTimes > maxRetryTimes {
							log.Println("Can not get remote nonce after retry " + string(rune(maxRetryTimes)) + " times")
							log.Println("Switch to next account")
							continue batchMint
						}
					}
				}
			}
		}
		log.Println("Mint finished")
	},
}

func init() {
	rootCmd.AddCommand(mintCmd)
	mintCmd.Flags().StringP("mnemonic", "m", "", "Set mnemonic")
	mintCmd.Flags().StringP("rpc", "r", "", "Set rpc")
	mintCmd.Flags().StringP("hex-content", "", "", "Set inscriptions with hex content")
	mintCmd.Flags().StringP("text-content", "", "", "Set inscriptions with text content")
	mintCmd.Flags().UintP("per-address-minted", "p", 10, "Each address can mint how many inscriptions,default 10")
	mintCmd.Flags().UintP("start-index", "s", 0, "Start index of bip-44 sequence addresses,default 0")
	mintCmd.Flags().UintP("end-index", "e", 0, "End index of bip-44 sequence addresses,default 0")

}
