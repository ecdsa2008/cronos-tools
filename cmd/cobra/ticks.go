package cobra

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"log"
	"net/http"
	"sort"
	"time"
)

var ticksCmd = &cobra.Command{
	Use:   "ticks",
	Short: "Get ticks holders and minting progress, and deployed time",

	Run: func(cmd *cobra.Command, args []string) {
		log.Println("ticks called")
		sortByDeployedTime, err := cmd.Flags().GetBool("sort-by-deployed-time")
		if err != nil {
			log.Panicln(err)
		}
		sortByMintingProgress, err := cmd.Flags().GetBool("sort-by-minting-progress")
		if err != nil {
			log.Panicln(err)
		}
		sortByHolders, err := cmd.Flags().GetBool("sort-by-holders")
		if err != nil {
			log.Panicln(err)
		}

		if sortByDeployedTime == false && sortByMintingProgress == false && sortByHolders == false {
			sortByDeployedTime = true
		}
		ticksInfo, err := getTicksInfo()
		if err != nil {
			log.Panicln("Error fetching ticks info:", err)
		}
		if len(ticksInfo.Content) == 0 {
			log.Println("No ticks info")
			return
		}

		if sortByDeployedTime {
			log.Println("Sort by deployed time:")
			sort.Slice(ticksInfo.Content, func(i, j int) bool {
				return ticksInfo.Content[i].DeployTime.After(ticksInfo.Content[j].DeployTime)
			})

			for _, tick := range ticksInfo.Content {
				log.Println("Tick:", tick.Tick, "Holder count:", tick.HolderCount, "Minting progress:", tick.Progress, "Deployed time:", tick.DeployTime.Format(time.RFC3339))
			}
			return
		}

		if sortByMintingProgress {
			log.Println("Sort by minting progress:")
			sort.Slice(ticksInfo.Content, func(i, j int) bool {
				return ticksInfo.Content[i].Progress > ticksInfo.Content[j].Progress
			})

			for _, tick := range ticksInfo.Content {
				log.Println("Tick:", tick.Tick, "Holder count:", tick.HolderCount, "Minting progress:", tick.Progress, "Deployed time:", tick.DeployTime.Format(time.RFC3339))
			}
			return
		}

		if sortByHolders {
			log.Println("Sort by holders:")
			sort.Slice(ticksInfo.Content, func(i, j int) bool {
				return ticksInfo.Content[i].HolderCount > ticksInfo.Content[j].HolderCount
			})

			for _, tick := range ticksInfo.Content {
				log.Println("Tick:", tick.Tick, "Holder count:", tick.HolderCount, "Minting progress:", tick.Progress, "Deployed time:", tick.DeployTime.Format(time.RFC3339))
			}
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(ticksCmd)
	ticksCmd.Flags().BoolP("sort-by-deployed-time", "", false, "Sort by deployed time, default is sort by deployed time")
	ticksCmd.Flags().BoolP("sort-by-minting-progress", "", false, "Sort by minting progress, default is sort by deployed time")
	ticksCmd.Flags().BoolP("sort-by-holders", "", false, "Sort by holders, default is sort by deployed time")
}

func getTicksInfo() (*TicksInfo, error) {
	url := "https://api.croscribe.com/v2/inscriptions?page=0&size=100000"
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
	var ticksInfo TicksInfo
	if err := json.Unmarshal(body, &ticksInfo); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil, err

	}
	return &ticksInfo, nil
}

type TicksInfo struct {
	Content []struct {
		Id          int       `json:"id"`
		Protocol    string    `json:"protocol"`
		Tick        string    `json:"tick"`
		DeployTime  time.Time `json:"deploy_time"`
		Progress    float64   `json:"progress"`
		HolderCount int       `json:"holder_count"`
		TotalSupply int64     `json:"total_supply"`
		MintedCount int64     `json:"minted_count"`
	} `json:"content"`
	Pageable struct {
		PageNumber int `json:"page_number"`
		PageSize   int `json:"page_size"`
		Sort       struct {
			Sorted   bool `json:"sorted"`
			Empty    bool `json:"empty"`
			Unsorted bool `json:"unsorted"`
		} `json:"sort"`
		Offset  int  `json:"offset"`
		Paged   bool `json:"paged"`
		Unpaged bool `json:"unpaged"`
	} `json:"pageable"`
	Last          bool `json:"last"`
	TotalPages    int  `json:"total_pages"`
	TotalElements int  `json:"total_elements"`
	First         bool `json:"first"`
	Size          int  `json:"size"`
	Number        int  `json:"number"`
	Sort          struct {
		Sorted   bool `json:"sorted"`
		Empty    bool `json:"empty"`
		Unsorted bool `json:"unsorted"`
	} `json:"sort"`
	NumberOfElements int  `json:"number_of_elements"`
	Empty            bool `json:"empty"`
}
