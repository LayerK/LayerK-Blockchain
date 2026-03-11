package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const defaultRPCURL = "https://mainnet-rpc.layerk.com"
const defaultPollInterval = 5 * time.Second
const defaultRequestTimeout = 10 * time.Second
const defaultMaxBlocksPerPoll = 128

var addressRegex = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)

var defaultAddressesToMonitor = []string{
	"0xE01B9E7A53629D23ee7571A3cF05C3188883f35e",
	"0xDe96e7Ed414943Ebb73aE64B517166Ad22e39729",
}

var rpcURL = loadRPCURL()
var pollInterval = loadDurationEnvMillis("POLL_INTERVAL_MS", defaultPollInterval)
var requestTimeout = loadDurationEnvMillis("REQUEST_TIMEOUT_MS", defaultRequestTimeout)
var maxBlocksPerPoll = loadPositiveIntEnv("MAX_BLOCKS_PER_POLL", defaultMaxBlocksPerPoll)

var httpClient = &http.Client{
	Timeout: requestTimeout,
}

// Addresses to monitor
var addressesToMonitor = loadMonitoredAddresses()

var monitoredSet = func() map[string]struct{} {
	set := make(map[string]struct{}, len(addressesToMonitor))
	for _, addr := range addressesToMonitor {
		set[strings.ToLower(addr)] = struct{}{}
	}
	return set
}()

func loadPositiveIntEnv(name string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		fmt.Printf("Invalid %s=%q; using default %d\n", name, raw, fallback)
		return fallback
	}
	return value
}

func loadDurationEnvMillis(name string, fallback time.Duration) time.Duration {
	ms := loadPositiveIntEnv(name, int(fallback/time.Millisecond))
	return time.Duration(ms) * time.Millisecond
}

func loadRPCURL() string {
	raw := strings.TrimSpace(os.Getenv("LAYERK_RPC_URL"))
	if raw == "" {
		return defaultRPCURL
	}

	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		fmt.Printf("Invalid LAYERK_RPC_URL=%q; using default %s\n", raw, defaultRPCURL)
		return defaultRPCURL
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		fmt.Printf("Unsupported LAYERK_RPC_URL scheme %q; using default %s\n", parsed.Scheme, defaultRPCURL)
		return defaultRPCURL
	}
	return parsed.String()
}

func loadMonitoredAddresses() []string {
	raw := strings.TrimSpace(os.Getenv("MONITORED_ADDRESSES"))
	candidates := defaultAddressesToMonitor
	if raw != "" {
		candidates = strings.Split(raw, ",")
	}

	result := make([]string, 0, len(candidates))
	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		address := strings.ToLower(strings.TrimSpace(candidate))
		if address == "" {
			continue
		}
		if !addressRegex.MatchString(address) {
			fmt.Printf("Skipping invalid monitored address: %s\n", candidate)
			continue
		}
		if _, ok := seen[address]; ok {
			continue
		}
		seen[address] = struct{}{}
		result = append(result, address)
	}

	if len(result) == 0 {
		fmt.Println("No valid MONITORED_ADDRESSES configured; falling back to defaults")
		for _, candidate := range defaultAddressesToMonitor {
			address := strings.ToLower(candidate)
			if _, ok := seen[address]; ok {
				continue
			}
			seen[address] = struct{}{}
			result = append(result, address)
		}
	}

	return result
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type rpcTx struct {
	From  string  `json:"from"`
	To    *string `json:"to"`
	Hash  string  `json:"hash"`
	Value string  `json:"value"`
}

type rpcBlock struct {
	Transactions []rpcTx `json:"transactions"`
}

// Function to make JSON-RPC calls
func makeRPCRequest(method string, params []interface{}, result interface{}) error {
	// Prepare the JSON payload
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  method,
		"params":  params,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Make the HTTP POST request
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rpcURL, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected RPC status code: %d", resp.StatusCode)
	}

	// Decode the JSON response
	var response rpcResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return err
	}

	// Check for errors in the response
	if response.Error != nil {
		return fmt.Errorf("RPC Error: %d %s", response.Error.Code, response.Error.Message)
	}

	if result == nil {
		return nil
	}

	if len(response.Result) == 0 || bytes.Equal(bytes.TrimSpace(response.Result), []byte("null")) {
		return fmt.Errorf("RPC result is null")
	}

	if err := json.Unmarshal(response.Result, result); err != nil {
		return err
	}

	return nil
}

// Function to check if an address is in the monitoring list
func isMonitoredAddress(address string) bool {
	if address == "" {
		return false
	}
	_, ok := monitoredSet[strings.ToLower(address)]
	return ok
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
	var block rpcBlock
	if err := makeRPCRequest("eth_getBlockByNumber", params, &block); err != nil {
		return fmt.Errorf("block %s returned empty result: %w", blockNumberHex, err)
	}

	if block.Transactions == nil {
		return fmt.Errorf("block %s missing transactions array", blockNumberHex)
	}

	for _, tx := range block.Transactions {
		from := tx.From
		to := ""
		if tx.To != nil {
			to = *tx.To
		}

		if isMonitoredAddress(from) || isMonitoredAddress(to) {
			fmt.Println("-----------------------------------------")
			fmt.Printf("Block Number: %s\n", blockNumberHex)
			fmt.Printf("Transaction Hash: %s\n", tx.Hash)
			fmt.Printf("From: %s\n", from)
			if to == "" {
				fmt.Printf("To: %s\n", "Contract Creation")
			} else {
				fmt.Printf("To: %s\n", to)
			}
			if tx.Value != "" {
				fmt.Printf("Value: %s LYK\n", weiToEther(tx.Value))
			}
			fmt.Println("-----------------------------------------\n")
		}
	}

	return nil
}

func main() {
	var lastBlockNumberHex string
	fmt.Printf(
		"Monitoring %d address(es) via %s (poll=%s timeout=%s maxBlocksPerPoll=%d)\n",
		len(monitoredSet),
		rpcURL,
		pollInterval,
		requestTimeout,
		maxBlocksPerPoll,
	)

	for {
		// Get the latest block number
		var currentBlockNumberHex string
		if err := makeRPCRequest("eth_blockNumber", []interface{}{}, &currentBlockNumberHex); err != nil {
			fmt.Println("Error fetching block number:", err)
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

            startBlockNum := new(big.Int).Set(lastBlockNum)
            diff := new(big.Int).Sub(currentBlockNum, lastBlockNum)
            maxRange := big.NewInt(int64(maxBlocksPerPoll))
            if diff.Cmp(maxRange) > 0 {
                startBlockNum = new(big.Int).Sub(currentBlockNum, maxRange)
                fmt.Printf(
                    "Backlog detected (%s blocks). Processing only the latest %d blocks this cycle.\n",
                    diff.String(),
                    maxBlocksPerPoll,
                )
            }

            // Check new blocks since the last known (bounded by maxBlocksPerPoll)
            for i := new(big.Int).Add(startBlockNum, big.NewInt(1)); i.Cmp(currentBlockNum) <= 0; i.Add(i, big.NewInt(1)) {
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
