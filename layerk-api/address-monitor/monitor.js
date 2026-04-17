const { ethers } = require('ethers');

const DEFAULT_RPC_URL = 'https://mainnet-rpc.layerk.com';
const DEFAULT_MONITORED_ADDRESSES = [
  '0xE01B9E7A53629D23ee7571A3cF05C3188883f35e',
  '0xDe96e7Ed414943Ebb73aE64B517166Ad22e39729',
];
const DEFAULT_MIN_CONFIRMATIONS = 3;
const DEFAULT_MAX_BLOCK_QUEUE = 32;
const DEFAULT_PROCESSED_BLOCK_CACHE = 256;

function parsePositiveIntEnv(name, fallback) {
  const raw = process.env[name];
  if (!raw) return fallback;
  const parsed = Number.parseInt(raw, 10);
  if (!Number.isFinite(parsed) || parsed <= 0) {
    console.warn(`Invalid ${name}=${raw}; using default ${fallback}`);
    return fallback;
  }
  return parsed;
}

function parseNonNegativeIntEnv(name, fallback) {
  const raw = process.env[name];
  if (!raw) return fallback;
  const parsed = Number.parseInt(raw, 10);
  if (!Number.isFinite(parsed) || parsed < 0) {
    console.warn(`Invalid ${name}=${raw}; using default ${fallback}`);
    return fallback;
  }
  return parsed;
}

function normalizeAddress(address) {
  try {
    return ethers.utils.getAddress(address).toLowerCase();
  } catch {
    return null;
  }
}

function loadRpcUrl() {
  const rpcUrl = process.env.LAYERK_RPC_URL || DEFAULT_RPC_URL;
  try {
    const parsed = new URL(rpcUrl);
    if (!['http:', 'https:'].includes(parsed.protocol)) {
      throw new Error(`Unsupported protocol ${parsed.protocol}`);
    }
    return parsed.toString();
  } catch (error) {
    throw new Error(`Invalid LAYERK_RPC_URL "${rpcUrl}": ${error.message}`);
  }
}

function loadMonitoredAddresses() {
  const raw = process.env.MONITORED_ADDRESSES;
  const candidates = raw
    ? raw
        .split(',')
        .map((value) => value.trim())
        .filter(Boolean)
    : DEFAULT_MONITORED_ADDRESSES;

  const normalized = [];
  const seen = new Set();
  for (const address of candidates) {
    const value = normalizeAddress(address);
    if (!value) {
      console.warn(`Skipping invalid monitored address: ${address}`);
      continue;
    }
    if (seen.has(value)) continue;
    seen.add(value);
    normalized.push(value);
  }

  if (normalized.length === 0) {
    throw new Error('No valid monitored addresses configured');
  }

  return normalized;
}

const RPC_URL = loadRpcUrl();
const MIN_CONFIRMATIONS = parseNonNegativeIntEnv('MIN_CONFIRMATIONS', DEFAULT_MIN_CONFIRMATIONS);
const MAX_BLOCK_QUEUE = parsePositiveIntEnv('MAX_BLOCK_QUEUE', DEFAULT_MAX_BLOCK_QUEUE);
const PROCESSED_BLOCK_CACHE = parsePositiveIntEnv(
  'PROCESSED_BLOCK_CACHE',
  DEFAULT_PROCESSED_BLOCK_CACHE
);
const monitoredSet = new Set(loadMonitoredAddresses());

const provider = new ethers.providers.JsonRpcProvider(RPC_URL, undefined, {
  polling: true,
  staticNetwork: true,
  batchMaxCount: 5,
});

function isMonitoredAddress(address) {
  return Boolean(address && monitoredSet.has(address.toLowerCase()));
}

const processedBlocks = new Map();
let lastProcessedBlockNumber = null;
let lastProcessedBlockHash = null;
let lastQueuedFinalizedBlock = 0;

function rememberProcessedBlock(blockNumber, blockHash) {
  processedBlocks.set(blockNumber, blockHash);
  while (processedBlocks.size > PROCESSED_BLOCK_CACHE) {
    const oldest = processedBlocks.keys().next().value;
    processedBlocks.delete(oldest);
  }
}

async function checkBlock(blockNumber) {
  const block = await provider.getBlockWithTransactions(blockNumber);

  if (!block || !block.hash || !Array.isArray(block.transactions)) {
    console.warn(`Block ${blockNumber} returned no transactions`);
    return;
  }

  if (processedBlocks.get(block.number) === block.hash) {
    return;
  }

  if (
    lastProcessedBlockNumber !== null &&
    block.number === lastProcessedBlockNumber + 1 &&
    lastProcessedBlockHash &&
    block.parentHash !== lastProcessedBlockHash
  ) {
    console.warn(
      `Detected non-contiguous finalized block stream at ${block.number}; expected parent ${lastProcessedBlockHash}, got ${block.parentHash}`
    );
  }

  rememberProcessedBlock(block.number, block.hash);
  lastProcessedBlockNumber = block.number;
  lastProcessedBlockHash = block.hash;

  for (const tx of block.transactions) {
    if (!tx) continue;
    if (!isMonitoredAddress(tx.from) && !isMonitoredAddress(tx.to)) {
      continue;
    }
    console.log('-----------------------------------------');
    console.log(`Block Number: ${blockNumber}`);
    console.log(`Transaction Hash: ${tx.hash}`);
    console.log(`From: ${tx.from}`);
    console.log(`To: ${tx.to || 'Contract Creation'}`);
    console.log(`Value: ${ethers.utils.formatEther(tx.value || 0)} LYK`);
    console.log('-----------------------------------------\n');
  }
}

const blockQueue = [];
const queuedBlocks = new Set();
let isProcessingQueue = false;

function enqueueFinalizedBlocks(headBlockNumber) {
  const latestFinalizedBlock = headBlockNumber - MIN_CONFIRMATIONS;
  if (latestFinalizedBlock <= 0 || latestFinalizedBlock <= lastQueuedFinalizedBlock) {
    return;
  }

  if (lastQueuedFinalizedBlock === 0 && lastProcessedBlockNumber === null && blockQueue.length === 0) {
    lastQueuedFinalizedBlock = latestFinalizedBlock;
    return;
  }

  let startBlock = lastQueuedFinalizedBlock + 1;
  const blocksToQueue = latestFinalizedBlock - startBlock + 1;
  if (blocksToQueue > MAX_BLOCK_QUEUE) {
    startBlock = latestFinalizedBlock - MAX_BLOCK_QUEUE + 1;
    blockQueue.length = 0;
    queuedBlocks.clear();
    console.warn(
      `Backlog detected (${blocksToQueue} finalized blocks). Only queueing the latest ${MAX_BLOCK_QUEUE} finalized blocks.`
    );
  }

  for (let blockNumber = startBlock; blockNumber <= latestFinalizedBlock; blockNumber += 1) {
    if (queuedBlocks.has(blockNumber)) continue;
    blockQueue.push(blockNumber);
    queuedBlocks.add(blockNumber);
  }

  lastQueuedFinalizedBlock = latestFinalizedBlock;
}

async function drainQueue() {
  if (isProcessingQueue) return;
  isProcessingQueue = true;

  try {
    while (blockQueue.length > 0) {
      const blockNumber = blockQueue.shift();
      queuedBlocks.delete(blockNumber);
      try {
        await checkBlock(blockNumber);
      } catch (err) {
        console.error(`Error checking block ${blockNumber}:`, err);
      }
    }
  } finally {
    isProcessingQueue = false;
  }
}

provider.on('block', (blockNumber) => {
  enqueueFinalizedBlocks(blockNumber);
  void drainQueue();
});

provider.on('error', (error) => {
  console.error('Provider error:', error);
});

console.log(
  `Monitoring ${monitoredSet.size} address(es) using ${RPC_URL} with ${MIN_CONFIRMATIONS} confirmation(s)`
);
