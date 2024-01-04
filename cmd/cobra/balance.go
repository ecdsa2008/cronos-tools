package cobra

import (
	"cronos-tools/src/utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
	"io"
	"log"
	"net/http"
)

var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Get tick balance of an address",

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

		forAllTicks := false
		tick, err := cmd.Flags().GetString("tick")
		if err != nil {
			log.Panicln(errors.New("tick is required"))
		}
		if tick == "" {
			forAllTicks = true
		}
		totalInscriptions := make(map[string]int)

		for i := startIndex; i <= endIndex; i++ {
			// 获取当前账户的私钥
			accountPrivateKey := utils.GetPrivateKey(mnemonic, i)
			// 获取当前账户的地址
			accountAddress := utils.GetAddressFromPrivateKey(accountPrivateKey)
			// 获取当前账户的余额
			ticksBalance, err := GetInscriptionBalance(accountAddress)
			if err != nil {
				log.Panicln("Error fetching inscription balance:", err)
			}
			if len(ticksBalance.Data) == 0 {
				log.Println("Account index:", i, "Address:", accountAddress.Hex(), "No balance")
				continue
			}
			if forAllTicks {
				for _, balance := range ticksBalance.Data {
					totalInscriptions[balance.Tick] += balance.Amount
					log.Printf("Account index: %d, Address: %s, Tick: %s, Amount: %d\n", i, accountAddress.Hex(), balance.Tick, balance.Amount)
				}
			} else {
				hasBalance := false
				for _, balance := range ticksBalance.Data {
					if balance.Tick == tick {
						totalInscriptions[balance.Tick] += balance.Amount
						hasBalance = true
						log.Printf("Account index: %d, Address: %s, Tick: %s, Amount: %d\n", i, accountAddress.Hex(), balance.Tick, balance.Amount)
					}
				}
				if !hasBalance {
					log.Printf("Account index: %d, Address: %s, Tick: %s, Amount: %d\n", i, accountAddress.Hex(), tick, 0)
				}
			}
		}
		log.Println("totalInscriptions:", totalInscriptions)
	},
}

func init() {
	rootCmd.AddCommand(balanceCmd)
	balanceCmd.Flags().StringP("mnemonic", "m", "", "Set mnemonic")
	balanceCmd.Flags().StringP("tick", "t", "", "Specify the tick")
	balanceCmd.Flags().UintP("start-index", "s", 0, "Start index of bip-44 sequence addresses,default 0")
	balanceCmd.Flags().UintP("end-index", "e", 0, "End index of bip-44 sequence addresses,default 0")
}

// Get all ticks balance of an address
// https://api.croscribe.com/balance/0xeb0c56a29e13F1594d794158c507f77bfd5B6eC8

func GetInscriptionBalance(address common.Address) (*TicksBalance, error) {
	baseUrl := "https://api.croscribe.com/balance"
	url := fmt.Sprintf("%s/%s", baseUrl, address.Hex())
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching data:", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil, err
	}

	var ticksBalance TicksBalance
	if err := json.Unmarshal(body, &ticksBalance); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil, err

	}
	return &ticksBalance, nil
}

type TicksBalance struct {
	Data []TickBalanceInfo `json:"balances"`
}

type TickBalanceInfo struct {
	TokenId  int    `json:"token_id"`
	Chain    string `json:"chain"`
	Protocol string `json:"protocol"`
	Tick     string `json:"tick"`
	Amount   int    `json:"amount"`
}
