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
const pollInterval = 5 * time.Second

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

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
	req, err := http.NewRequest(http.MethodPost, rpcURL, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected RPC status code: %d", resp.StatusCode)
	}

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
	if len(weiStr) < 3 || !strings.HasPrefix(weiStr, "0x") {
		return "0"
	}
	if _, ok := value.SetString(weiStr[2:], 16); !ok {
		return "0"
	}
	etherValue := new(big.Float).Quo(new(big.Float).SetInt(value), big.NewFloat(1e18))
	return etherValue.Text('f', 18)
}

func hexToBigInt(hexStr string) (*big.Int, error) {
	if len(hexStr) < 3 || !strings.HasPrefix(hexStr, "0x") {
		return nil, fmt.Errorf("invalid hex string: %s", hexStr)
	}
	value := new(big.Int)
	if _, ok := value.SetString(hexStr[2:], 16); !ok {
		return nil, fmt.Errorf("cannot parse hex string: %s", hexStr)
	}
	return value, nil
}

// Function to check transactions in a block
func checkBlock(blockNumberHex string) error {
	fmt.Printf("Checking block %s...\n", blockNumberHex)
	params := []interface{}{blockNumberHex, true}
	response, err := makeRPCRequest("eth_getBlockByNumber", params)
	if err != nil {
		return err
	}

	rawResult, hasResult := response["result"]
	if !hasResult || rawResult == nil {
		return fmt.Errorf("block %s returned empty result", blockNumberHex)
	}

	blockData, ok := rawResult.(map[string]interface{})
	if !ok {
		return fmt.Errorf("block %s result has unexpected shape", blockNumberHex)
	}

	transactions, ok := blockData["transactions"].([]interface{})
	if !ok {
		return fmt.Errorf("block %s missing transactions array", blockNumberHex)
	}

	for _, txInterface := range transactions {
		tx, ok := txInterface.(map[string]interface{})
		if !ok {
			continue
		}

		from, _ := tx["from"].(string)
		to, _ := tx["to"].(string)

		if isMonitoredAddress(from) || isMonitoredAddress(to) {
			fmt.Println("-----------------------------------------")
			fmt.Printf("Block Number: %s\n", blockNumberHex)
			if hash, ok := tx["hash"].(string); ok {
				fmt.Printf("Transaction Hash: %s\n", hash)
			}
			fmt.Printf("From: %s\n", from)
			fmt.Printf("To: %s\n", to)
			if value, ok := tx["value"].(string); ok {
				fmt.Printf("Value: %s LYK\n", weiToEther(value))
			}
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
			time.Sleep(pollInterval)
			continue
		}

		currentBlockNumberHex, ok := response["result"].(string)
		if !ok {
			fmt.Println("RPC response missing block number result")
			time.Sleep(pollInterval)
			continue
		}

		if lastBlockNumberHex == "" {
			lastBlockNumberHex = currentBlockNumberHex
			time.Sleep(pollInterval)
			continue
		}

		if currentBlockNumberHex != lastBlockNumberHex {
			// Convert hex block numbers to integers
			lastBlockNum, err := hexToBigInt(lastBlockNumberHex)
			if err != nil {
				fmt.Println("Error parsing last block number:", err)
				lastBlockNumberHex = currentBlockNumberHex
				time.Sleep(pollInterval)
				continue
			}
			currentBlockNum, err := hexToBigInt(currentBlockNumberHex)
			if err != nil {
				fmt.Println("Error parsing current block number:", err)
				time.Sleep(pollInterval)
				continue
			}

			// Check all new blocks since the last known block
			for i := new(big.Int).Add(lastBlockNum, big.NewInt(1)); i.Cmp(currentBlockNum) <= 0; i.Add(i, big.NewInt(1)) {
				blockNumberHex := "0x" + i.Text(16)
				if err := checkBlock(blockNumberHex); err != nil {
					fmt.Printf("Error checking block %s: %v\n", blockNumberHex, err)
				}
			}

			lastBlockNumberHex = currentBlockNumberHex
		}

		time.Sleep(pollInterval)
	}
}
