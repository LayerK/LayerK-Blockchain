const axios = require('axios');

const baseUrl = 'https://explorer.layerk.com/api/v2';
const address = '0xE01B9E7A53629D23ee7571A3cF05C3188883f35e';
const lykDecimals = 18n;

const api = axios.create({
  baseURL: baseUrl,
  timeout: 10_000,
});

function formatBalance(raw) {
  if (raw === null || raw === undefined) {
    return '0';
  }
  const value = BigInt(raw);
  const whole = value / 10n ** lykDecimals;
  const fraction = value % 10n ** lykDecimals;
  const paddedFraction = fraction.toString().padStart(Number(lykDecimals), '0');
  return `${whole}.${paddedFraction.replace(/0+$/, '') || '0'}`;
}

function logError(prefix, error) {
  if (error.response) {
    console.error(`${prefix}: ${error.response.status}`, error.response.data);
  } else if (error.request) {
    console.error(`${prefix}: no response`, error.message);
  } else {
    console.error(`${prefix}:`, error);
  }
}

async function getBalance(targetAddress) {
  try {
    const { data } = await api.get(`/addresses/${targetAddress}`);
    if (!data || data.coin_balance === undefined) {
      console.warn('Balance response missing coin_balance');
      return null;
    }
    const balanceLYK = formatBalance(data.coin_balance);
    console.log(`Balance of ${targetAddress}: ${balanceLYK} LYK`);
    return balanceLYK;
  } catch (error) {
    logError('Error getting balance', error);
    return null;
  }
}

async function getLYKTransactions(targetAddress) {
  const transactions = [];
  let nextPageParams = null;

  try {
    do {
      let url = `/addresses/${targetAddress}/transactions`;
      if (nextPageParams) {
        const params = new URLSearchParams(nextPageParams);
        url += `?${params.toString()}`;
      }

      const { data } = await api.get(url);
      const items = Array.isArray(data.items) ? data.items : [];

      items.forEach((tx) => {
        if (!tx || !Array.isArray(tx.tx_types)) {
          return;
        }
        if (tx.tx_types.includes('coin_transfer')) {
          transactions.push(tx);
        }
      });

      nextPageParams = data.next_page_params || null;
    } while (nextPageParams);

    console.log(`Found ${transactions.length} LYK coin transfer transactions`);
    transactions.forEach((tx) => {
      const timestamp = tx.timestamp ? new Date(tx.timestamp).toLocaleString() : 'Unknown';
      const value = formatBalance(tx.value || 0);
      console.log(
        `Date: ${timestamp}\nTx Hash: ${tx.hash}\nFrom: ${tx.from?.hash}\nTo: ${tx.to?.hash}\nValue: ${value} LYK\n---`
      );
    });

    return transactions;
  } catch (error) {
    logError('Error getting transactions', error);
    return [];
  }
}

async function main() {
  await getBalance(address);
  await getLYKTransactions(address);
}

main().catch((error) => {
  console.error('Fatal error running example:', error);
});
