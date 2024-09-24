const axios = require('axios');

const baseUrl = 'https://explorer.layerk.com/api/v2';
const address = '0xE01B9E7A53629D23ee7571A3cF05C3188883f35e';

// Function to get the LYK balance of the account
async function getBalance(address) {
  try {
    const response = await axios.get(`${baseUrl}/addresses/${address}`);
    const data = response.data;
    const balanceWei = data.coin_balance;
    // Assuming LYK has 18 decimals like Ethereum
    const balanceLYK = balanceWei / 1e18;
    console.log(`Balance of ${address}: ${balanceLYK} LYK`);
    return balanceLYK;
  } catch (error) {
    console.error(`Error getting balance: ${error}`);
  }
}

// Function to get all LYK transactions to and from the account
async function getLYKTransactions(address) {
  let transactions = [];
  let nextPageParams = null;
  try {
    do {
      let url = `${baseUrl}/addresses/${address}/transactions`;
      if (nextPageParams) {
        // Append next_page_params to the query string
        const params = new URLSearchParams(nextPageParams);
        url += `?${params.toString()}`;
      }
      const response = await axios.get(url);
      const data = response.data;
      const items = data.items;
      // Filter transactions where tx_types includes 'coin_transfer'
      const coinTransfers = items.filter((tx) =>
        tx.tx_types.includes('coin_transfer')
      );
      transactions = transactions.concat(coinTransfers);
      nextPageParams = data.next_page_params;
    } while (nextPageParams);

    console.log(`Found ${transactions.length} LYK coin transfer transactions`);
    // Output the transactions
    for (const tx of transactions) {
      // Format the timestamp
      const date = new Date(tx.timestamp);
      console.log(
        `Date: ${date.toLocaleString()}\nTx Hash: ${tx.hash}\nFrom: ${tx.from.hash}\nTo: ${tx.to.hash}\nValue: ${
          tx.value / 1e18
        } LYK\n---`
      );
    }
    return transactions;
  } catch (error) {
    console.error(`Error getting transactions: ${error}`);
  }
}

// Main function
async function main() {
  await getBalance(address);
  await getLYKTransactions(address);
}

main();
