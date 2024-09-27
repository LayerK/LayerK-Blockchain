package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"
)

const rpcURL = "https://mainnet-rpc.layerk.com"

// Addresses to monitor
var addressesToMonitor = []string{
	"0xE01B9E7A53629D23ee7571A3cF05C3188883f35e",
	"0xDe96e7Ed414943Ebb73aE64B517166Ad22e39729",
}

// Function to make JSON-RPC calls
func makeRPCRequest(method string, params []interface{}) (map[string]interface{}, error) {
	// Prepare the JSON payload
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  method,
		"params":  params,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Make the HTTP POST request
	resp, err := http.Post(rpcURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Decode the JSON response
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	// Check for errors in the response
	if errData, exists := result["error"]; exists {
		return nil, fmt.Errorf("RPC Error: %v", errData)
	}

	return result, nil
}

// Function to check if an address is in the monitoring list
func isMonitoredAddress(address string) bool {
	address = strings.ToLower(address)
	for _, addr := range addressesToMonitor {
		if strings.ToLower(addr) == address {
			return true
		}
	}
	return false
}

// Function to convert Wei to Ether
func weiToEther(weiStr string) string {
	value := new(big.Int)
	value.SetString(weiStr[2:], 16) // Remove '0x' prefix
	etherValue := new(big.Float).Quo(new(big.Float).SetInt(value), big.NewFloat(1e18))
	return etherValue.Text('f', 18)
}

// Function to check transactions in a block
func checkBlock(blockNumberHex string) error {
	fmt.Printf("Checking block %s...\n", blockNumberHex)
	params := []interface{}{blockNumberHex, true}
	response, err := makeRPCRequest("eth_getBlockByNumber", params)
	if err != nil {
		return err
	}

	blockData := response["result"].(map[string]interface{})
	transactions := blockData["transactions"].([]interface{})

	for _, txInterface := range transactions {
		tx := txInterface.(map[string]interface{})
		from := tx["from"].(string)
		to, _ := tx["to"].(string)

		if isMonitoredAddress(from) || isMonitoredAddress(to) {
			fmt.Println("-----------------------------------------")
			fmt.Printf("Block Number: %s\n", blockNumberHex)
			fmt.Printf("Transaction Hash: %s\n", tx["hash"].(string))
			fmt.Printf("From: %s\n", from)
			fmt.Printf("To: %s\n", to)
			fmt.Printf("Value: %s LYK\n", weiToEther(tx["value"].(string)))
			fmt.Println("-----------------------------------------\n")
		}
	}

	return nil
}

func main() {
	var lastBlockNumberHex string

	for {
		// Get the latest block number
		response, err := makeRPCRequest("eth_blockNumber", []interface{}{})
		if err != nil {
			fmt.Println("Error fetching block number:", err)
			time.Sleep(5 * time.Second)
			continue
		}

		currentBlockNumberHex := response["result"].(string)
		if currentBlockNumberHex != lastBlockNumberHex {
			// Convert hex block numbers to integers
			lastBlockNum := new(big.Int)
			currentBlockNum := new(big.Int)
			if lastBlockNumberHex != "" {
				lastBlockNum.SetString(lastBlockNumberHex[2:], 16)
			} else {
				lastBlockNum.SetInt64(0)
			}
			currentBlockNum.SetString(currentBlockNumberHex[2:], 16)

			// Check all new blocks since the last known block
			for i := new(big.Int).Add(lastBlockNum, big.NewInt(1)); i.Cmp(currentBlockNum) <= 0; i.Add(i, big.NewInt(1)) {
				blockNumberHex := "0x" + i.Text(16)
				err := checkBlock(blockNumberHex)
				if err != nil {
					fmt.Println("Error checking block:", err)
				}
			}

			lastBlockNumberHex = currentBlockNumberHex
		}

		time.Sleep(5 * time.Second)
	}
}
