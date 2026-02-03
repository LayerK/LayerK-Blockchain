const { ethers } = require('ethers');

const RPC_URL = 'https://mainnet-rpc.layerk.com';
const provider = new ethers.providers.JsonRpcProvider(RPC_URL, undefined, {
  polling: true,
  staticNetwork: true,
  batchMaxCount: 5,
});

const addressesToMonitor = [
  '0xE01B9E7A53629D23ee7571A3cF05C3188883f35e',
  '0xDe96e7Ed414943Ebb73aE64B517166Ad22e39729',
].map((address) => address.toLowerCase());

const monitoredSet = new Set(addressesToMonitor);

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

  block.transactions.forEach((tx) => {
    if (!tx) {
      return;
    }
    if (isMonitoredAddress(tx.from) || isMonitoredAddress(tx.to)) {
      console.log('-----------------------------------------');
      console.log(`Block Number: ${blockNumber}`);
      console.log(`Transaction Hash: ${tx.hash}`);
      console.log(`From: ${tx.from}`);
      console.log(`To: ${tx.to || 'Contract Creation'}`);
      console.log(`Value: ${ethers.utils.formatEther(tx.value || 0)} LYK`);
      console.log('-----------------------------------------\n');
    }
  });
}

let pending = Promise.resolve();

provider.on('block', (blockNumber) => {
  pending = pending
    .then(() => checkBlock(blockNumber))
    .catch((err) => {
      console.error(`Error checking block ${blockNumber}:`, err);
    });
});

provider.on('error', (error) => {
  console.error('Provider error:', error);
});
