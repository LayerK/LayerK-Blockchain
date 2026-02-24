const { ethers } = require('ethers');

const DEFAULT_RPC_URL = 'https://mainnet-rpc.layerk.com';
const DEFAULT_MONITORED_ADDRESSES = [
  '0xE01B9E7A53629D23ee7571A3cF05C3188883f35e',
  '0xDe96e7Ed414943Ebb73aE64B517166Ad22e39729',
];
const DEFAULT_MAX_BLOCK_QUEUE = 32;

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
const MAX_BLOCK_QUEUE = parsePositiveIntEnv('MAX_BLOCK_QUEUE', DEFAULT_MAX_BLOCK_QUEUE);
const monitoredSet = new Set(loadMonitoredAddresses());

const provider = new ethers.providers.JsonRpcProvider(RPC_URL, undefined, {
  polling: true,
  staticNetwork: true,
  batchMaxCount: 5,
});

function isMonitoredAddress(address) {
  return Boolean(address && monitoredSet.has(address.toLowerCase()));
}

async function checkBlock(blockNumber) {
  console.log(`Checking block ${blockNumber}...`);
  const block = await provider.getBlockWithTransactions(blockNumber);

  if (!block || !Array.isArray(block.transactions)) {
    console.warn(`Block ${blockNumber} returned no transactions`);
    return;
  }

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
  if (queuedBlocks.has(blockNumber)) {
    return;
  }

  if (blockQueue.length >= MAX_BLOCK_QUEUE) {
    const droppedBlock = blockQueue.shift();
    queuedBlocks.delete(droppedBlock);
    console.warn(
      `Block queue limit (${MAX_BLOCK_QUEUE}) reached; dropping oldest queued block ${droppedBlock}`
    );
  }

  blockQueue.push(blockNumber);
  queuedBlocks.add(blockNumber);
  void drainQueue();
});

provider.on('error', (error) => {
  console.error('Provider error:', error);
});

console.log(`Monitoring ${monitoredSet.size} address(es) using ${RPC_URL}`);
