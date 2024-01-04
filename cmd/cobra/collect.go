package cobra

import (
	"context"
	"cronos-tools/src/utils"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
	"log"
	"strconv"
	"strings"
	"time"
)

var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Collect all inscriptions about one tick",
	Run: func(cmd *cobra.Command, args []string) {
		mnemonic, err := cmd.Flags().GetString("mnemonic")
		if err != nil {
			log.Panicln(errors.New("mnemonic is required"))
		}
		if mnemonic == "" {
			log.Panicln(errors.New("mnemonic is required"))
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

		tick, err := cmd.Flags().GetString("tick")
		if err != nil {
			log.Panicln(errors.New("tick is required"))
		}

		if tick == "" {
			log.Panicln(errors.New("tick is required"))
		}
		tick = strings.TrimSpace(tick)

		rpc, err := cmd.Flags().GetString("rpc")
		if err != nil {
			log.Panicln(errors.New("rpc is required"))
		}

		if rpc == "" {
			log.Panicln(errors.New("rpc is required"))
		}

		collector, err := cmd.Flags().GetString("collector")
		if err != nil {
			log.Panicln(errors.New("collector is required"))
		}
		if collector == "" {
			log.Panicln(errors.New("collector is required"))
		}
		collector = strings.TrimPrefix(collector, "0x")
		collectorAddress := common.HexToAddress(collector)

		client, err := ethclient.Dial(rpc)
		if err != nil {
			log.Panicln(err)
		}
		gasLimit := uint64(22100)

		networkID, err := client.NetworkID(context.Background())
		if err != nil {
			log.Panicln(err)
		}

		for i := startIndex; i <= endIndex; i++ {
			// 获取当前账户的私钥
			accountPrivateKey := utils.GetPrivateKey(mnemonic, i)
			// 获取当前账户的地址
			accountAddress := utils.GetAddressFromPrivateKey(accountPrivateKey)
			if accountAddress == collectorAddress {
				continue
			}
			// 获取当前账户的所有铭文余额
			allTicksBalance, err := GetInscriptionBalance(accountAddress)
			if err != nil {
				log.Panicln("Error fetching inscription balance:", err)
			}
			if len(allTicksBalance.Data) == 0 {
				log.Println("Account index:", i, "Address:", accountAddress.Hex(), "No balance")
				continue
			}
			// 获取当前账户的指定铭文余额
			var tickBalance *TickBalanceInfo
			for _, tb := range allTicksBalance.Data {
				if strings.ToLower(tb.Tick) == strings.ToLower(tick) {
					tickBalance = &tb
					log.Println("Account index:", i, "Address:", accountAddress.Hex(), "Tick:", tick, "Amount:", tb.Amount)
					break
				}
			}
			if tickBalance == nil {
				log.Println("Account index:", i, "Address:", accountAddress.Hex(), "No balance for tick", tick)
				continue
			}

			// 获取当前账户的gasPrice
			gasPrice, err := client.SuggestGasPrice(context.Background())
			if err != nil {
				log.Println("Can not get gas price ", err)
				// 如果获取gasPrice失败，则等待10秒后
				for x := 0; x < 5; x++ {
					time.Sleep(5 * time.Second)
					gasPrice, err = client.SuggestGasPrice(context.Background())
					if err == nil {
						continue
					}
				}
				if err != nil {
					log.Panicln("Can not get gas price after retry 5 times ", err)
				}
			}

			// 检查当前账户的native coin余额是否足够支付gas fee
			nativeCoinBalance, err := client.BalanceAt(context.Background(), accountAddress, nil)
			if err != nil {
				log.Println("Can not get native coin balance ", err)
				// 如果获取balance失败，则等待10秒后
				for y := 0; y < 5; y++ {
					time.Sleep(5 * time.Second)
					nativeCoinBalance, err = client.BalanceAt(context.Background(), accountAddress, nil)
					if err == nil {
						continue
					}
				}
				if err != nil {
					log.Panicln("Can not get native coin balance after retry 5 times ", err)
				}
			}

			// 计算gas fee
			gasFee := decimal.NewFromBigInt(gasPrice, 0).Mul(decimal.NewFromInt(int64(gasLimit))).BigInt()
			if nativeCoinBalance.Cmp(gasFee) < 0 {
				log.Println("Account " + accountAddress.Hex() + " native coin balance is not enough to pay for gas fee")
				log.Println("Switch to next account")
				continue
			}

			// 获取当前账户的nonce
			nonce, err := client.PendingNonceAt(context.Background(), accountAddress)
			if err != nil {
				log.Println("Can not get nonce ", err)
				// 如果获取nonce失败，则等待10秒后
				for z := 0; z < 5; z++ {
					time.Sleep(5 * time.Second)
					nonce, err = client.PendingNonceAt(context.Background(), accountAddress)
					if err == nil {
						continue
					}
				}
				if err != nil {
					log.Panicln("Can not get nonce after retry 5 times ", err)
				}
			}
			// 构建payload
			payloadString := fmt.Sprintf(`data:,{"p":"crc-20","op":"transfer","tick":"%s","amt":"%s"}`, tick, strconv.Itoa(tickBalance.Amount))
			payload := []byte(payloadString)
			log.Println("Account index:", i, "Address:", accountAddress.Hex(), "Tick:", tick, "Amount:", tickBalance.Amount, "To:", collectorAddress.Hex())
			log.Println("Account index:", i, "Address:", accountAddress.Hex(), "Payload:", string(payload), "To:", collectorAddress.Hex())
			// 构造交易
			tx := types.NewTx(&types.LegacyTx{
				Nonce:    nonce,
				To:       &collectorAddress,
				Value:    decimal.Zero.BigInt(),
				Gas:      gasLimit,
				GasPrice: gasPrice,
				Data:     payload,
			})
			log.Println("Account index:", i, "Address:", accountAddress.Hex(), "build tx:", tx.Hash().Hex())

			signedTx, err := types.SignTx(tx, types.NewEIP155Signer(networkID), accountPrivateKey)
			if err != nil {
				log.Panicln("Can not sign transaction ", err)
			}
			// 发送交易
			err = client.SendTransaction(context.Background(), signedTx)
			if err != nil {
				log.Println("Account " + accountAddress.Hex() + " send transaction failed")
				log.Panicln("Can not send transaction ", err)
			}
			txHash := signedTx.Hash()
			txHashString := txHash.Hex()
			log.Println("Account index: ", i, " Address: ", accountAddress.Hex(), " Tx hash: ", txHashString, " Payload: ", string(payload))
		}
	},
}

func init() {
	rootCmd.AddCommand(collectCmd)
	collectCmd.Flags().StringP("mnemonic", "m", "", "Set mnemonic")
	collectCmd.Flags().StringP("tick", "t", "", "Specify the tick")
	collectCmd.Flags().StringP("rpc", "r", "", "Specify the rpc url")
	collectCmd.Flags().StringP("collector", "c", "", "Specify the collector address")
	collectCmd.Flags().UintP("start-index", "s", 0, "Start index of bip-44 sequence addresses,default 0")
	collectCmd.Flags().UintP("end-index", "e", 0, "End index of bip-44 sequence addresses,default 0")
}
