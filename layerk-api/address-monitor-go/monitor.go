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
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const defaultRPCURL = "https://mainnet-rpc.layerk.com"
const defaultMinConfirmations = 3
const defaultPollInterval = 5 * time.Second
const defaultRequestTimeout = 10 * time.Second
const defaultMaxBlocksPerPoll = 128

var addressRegex = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)

var defaultAddressesToMonitor = []string{
	"0xE01B9E7A53629D23ee7571A3cF05C3188883f35e",
	"0xDe96e7Ed414943Ebb73aE64B517166Ad22e39729",
}

var rpcURL = loadRPCURL()
var minConfirmations = loadNonNegativeIntEnv("MIN_CONFIRMATIONS", defaultMinConfirmations)
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

func loadNonNegativeIntEnv(name string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
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
	Number       string  `json:"number"`
	Hash         string  `json:"hash"`
	ParentHash   string  `json:"parentHash"`
	Transactions []rpcTx `json:"transactions"`
}

type rpcRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

// Function to make JSON-RPC calls
func makeRPCRequest(ctx context.Context, method string, params []interface{}, result interface{}) error {
	payload := rpcRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  method,
		Params:  params,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(
		requestCtx,
		http.MethodPost,
		rpcURL,
		bytes.NewReader(data),
	)
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

func waitForNextPoll(ctx context.Context) bool {
	timer := time.NewTimer(pollInterval)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// Function to check transactions in a block
func checkBlock(ctx context.Context, blockNumberHex string, expectedParentHash string) (string, error) {
	params := []interface{}{blockNumberHex, true}
	var block rpcBlock
	if err := makeRPCRequest(ctx, "eth_getBlockByNumber", params, &block); err != nil {
		return "", fmt.Errorf("block %s returned empty result: %w", blockNumberHex, err)
	}

	if block.Transactions == nil {
		return "", fmt.Errorf("block %s missing transactions array", blockNumberHex)
	}

	if block.Number == "" || block.Hash == "" {
		return "", fmt.Errorf("block %s missing hash metadata", blockNumberHex)
	}

	if expectedParentHash != "" && !strings.EqualFold(block.ParentHash, expectedParentHash) {
		return "", fmt.Errorf(
			"block %s parent hash mismatch: expected %s got %s",
			blockNumberHex,
			expectedParentHash,
			block.ParentHash,
		)
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

	return block.Hash, nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var lastProcessedBlockNum *big.Int
	var lastProcessedBlockHash string
	fmt.Printf(
		"Monitoring %d address(es) via %s (confirmations=%d poll=%s timeout=%s maxBlocksPerPoll=%d)\n",
		len(monitoredSet),
		rpcURL,
		minConfirmations,
		pollInterval,
		requestTimeout,
		maxBlocksPerPoll,
	)

	for {
		var currentBlockNumberHex string
		if err := makeRPCRequest(ctx, "eth_blockNumber", []interface{}{}, &currentBlockNumberHex); err != nil {
			if ctx.Err() != nil {
				return
			}
			fmt.Println("Error fetching block number:", err)
			if !waitForNextPoll(ctx) {
				return
			}
			continue
		}

		currentBlockNum, err := hexToBigInt(currentBlockNumberHex)
		if err != nil {
			fmt.Println("Error parsing current block number:", err)
			if !waitForNextPoll(ctx) {
				return
			}
			continue
		}

		finalizedBlockNum := new(big.Int).Set(currentBlockNum)
		if minConfirmations > 0 {
			confirmationWindow := big.NewInt(int64(minConfirmations))
			if currentBlockNum.Cmp(confirmationWindow) <= 0 {
				if !waitForNextPoll(ctx) {
					return
				}
				continue
			}
			finalizedBlockNum.Sub(finalizedBlockNum, confirmationWindow)
		}

		if lastProcessedBlockNum == nil {
			lastProcessedBlockNum = new(big.Int).Set(finalizedBlockNum)
			if !waitForNextPoll(ctx) {
				return
			}
			continue
		}

		if finalizedBlockNum.Cmp(lastProcessedBlockNum) < 0 {
			fmt.Printf(
				"Finalized head moved backwards from %s to %s; resetting cursor.\n",
				lastProcessedBlockNum.String(),
				finalizedBlockNum.String(),
			)
			lastProcessedBlockNum = new(big.Int).Set(finalizedBlockNum)
			lastProcessedBlockHash = ""
			if !waitForNextPoll(ctx) {
				return
			}
			continue
		}

		if finalizedBlockNum.Cmp(lastProcessedBlockNum) == 0 {
			if !waitForNextPoll(ctx) {
				return
			}
			continue
		}

		startBlockNum := new(big.Int).Add(lastProcessedBlockNum, big.NewInt(1))
		expectedParentHash := lastProcessedBlockHash
		diff := new(big.Int).Sub(finalizedBlockNum, lastProcessedBlockNum)
		maxRange := big.NewInt(int64(maxBlocksPerPoll))
		if diff.Cmp(maxRange) > 0 {
			startBlockNum = new(big.Int).Sub(finalizedBlockNum, maxRange)
			startBlockNum.Add(startBlockNum, big.NewInt(1))
			expectedParentHash = ""
			fmt.Printf(
				"Backlog detected (%s finalized blocks). Processing only the latest %d blocks this cycle.\n",
				diff.String(),
				maxBlocksPerPoll,
			)
		}

		one := big.NewInt(1)
		for i := new(big.Int).Set(startBlockNum); i.Cmp(finalizedBlockNum) <= 0; i.Add(i, one) {
			blockNumberHex := "0x" + i.Text(16)
			blockHash, err := checkBlock(ctx, blockNumberHex, expectedParentHash)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				fmt.Printf("Error checking block %s: %v\n", blockNumberHex, err)
				break
			}
			lastProcessedBlockNum = new(big.Int).Set(i)
			lastProcessedBlockHash = blockHash
			expectedParentHash = blockHash
		}

		if !waitForNextPoll(ctx) {
			return
		}
	}
}
