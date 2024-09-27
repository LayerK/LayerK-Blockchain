const { ethers } = require('ethers');

// RPC endpoint provided
const RPC_URL = 'https://mainnet-rpc.layerk.com';

// Initialize a provider
const provider = new ethers.providers.JsonRpcProvider(RPC_URL);

// Addresses to monitor
const addressesToMonitor = [
  '0xE01B9E7A53629D23ee7571A3cF05C3188883f35e',
  '0xDe96e7Ed414943Ebb73aE64B517166Ad22e39729',
];

// Function to check transactions in a block
async function checkBlock(blockNumber) {
  console.log(`Checking block ${blockNumber}...`);
  const block = await provider.getBlockWithTransactions(blockNumber);

  block.transactions.forEach((tx) => {
    if (
      addressesToMonitor.includes(tx.from) ||
      addressesToMonitor.includes(tx.to)
    ) {
      console.log('-----------------------------------------');
      console.log(`Block Number: ${blockNumber}`);
      console.log(`Transaction Hash: ${tx.hash}`);
      console.log(`From: ${tx.from}`);
      console.log(`To: ${tx.to}`);
      console.log(`Value: ${ethers.utils.formatEther(tx.value)} LYK`);
      console.log('-----------------------------------------\n');
    }
  });
}

// Listen for new blocks
provider.on('block', (blockNumber) => {
  checkBlock(blockNumber).catch((err) => {
    console.error(`Error checking block ${blockNumber}:`, err);
  });
});

// Handle errors
provider.on('error', (error) => {
  console.error('Provider error:', error);
});
