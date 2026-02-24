const axios = require('axios');

const DEFAULT_BASE_URL = 'https://explorer.layerk.com/api/v2';
const DEFAULT_ADDRESS = '0xE01B9E7A53629D23ee7571A3cF05C3188883f35e';
const DEFAULT_TIMEOUT_MS = 10_000;
const DEFAULT_MAX_PAGES = 100;
const lykDecimals = 18n;
const lykDivisor = 10n ** lykDecimals;

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

function getApiBaseUrl() {
  const raw = process.env.LAYERK_EXPLORER_API_URL || DEFAULT_BASE_URL;
  try {
    const parsed = new URL(raw);
    if (!['http:', 'https:'].includes(parsed.protocol)) {
      throw new Error(`Unsupported protocol ${parsed.protocol}`);
    }
    return parsed.toString().replace(/\/$/, '');
  } catch (error) {
    throw new Error(`Invalid LAYERK_EXPLORER_API_URL "${raw}": ${error.message}`);
  }
}

function getTargetAddress() {
  const raw = (process.env.TARGET_ADDRESS || DEFAULT_ADDRESS).trim();
  if (!/^0x[a-fA-F0-9]{40}$/.test(raw)) {
    throw new Error(`Invalid TARGET_ADDRESS "${raw}"`);
  }
  return raw;
}

function safeBigInt(raw, fallback = 0n) {
  if (raw === null || raw === undefined) return fallback;
  try {
    return BigInt(raw);
  } catch {
    return fallback;
  }
}

const baseUrl = getApiBaseUrl();
const address = getTargetAddress();
const maxPages = parsePositiveIntEnv('MAX_PAGES', DEFAULT_MAX_PAGES);

const api = axios.create({
  baseURL: baseUrl,
  timeout: parsePositiveIntEnv('REQUEST_TIMEOUT_MS', DEFAULT_TIMEOUT_MS),
});

function formatBalance(raw) {
  if (raw === null || raw === undefined) {
    return '0';
  }
  const value = safeBigInt(raw);
  const whole = value / lykDivisor;
  const fraction = value % lykDivisor;
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
  let totalTransfers = 0;
  let nextPageParams = null;
  const seenPageKeys = new Set();
  let pageCount = 0;

  try {
    do {
      if (pageCount >= maxPages) {
        console.warn(`Stopping after ${maxPages} pages to avoid unbounded pagination`);
        break;
      }

      const cursorKey = JSON.stringify(nextPageParams || {});
      if (seenPageKeys.has(cursorKey)) {
        console.warn('Detected repeated next_page_params cursor; stopping pagination');
        break;
      }
      seenPageKeys.add(cursorKey);
      pageCount += 1;

      const { data } = await api.get(`/addresses/${targetAddress}/transactions`, {
        params: nextPageParams || undefined,
      });
      const items = Array.isArray(data.items) ? data.items : [];

      items.forEach((tx) => {
        if (!tx || !Array.isArray(tx.tx_types)) {
          return;
        }
        if (tx.tx_types.includes('coin_transfer')) {
          totalTransfers += 1;
          const timestamp = tx.timestamp ? new Date(tx.timestamp).toLocaleString() : 'Unknown';
          const value = formatBalance(tx.value || 0);
          console.log(
            `Date: ${timestamp}\nTx Hash: ${tx.hash}\nFrom: ${tx.from?.hash}\nTo: ${tx.to?.hash}\nValue: ${value} LYK\n---`
          );
        }
      });

      nextPageParams = data.next_page_params || null;
    } while (nextPageParams);

    console.log(`Found ${totalTransfers} LYK coin transfer transactions across ${pageCount} page(s)`);
    return totalTransfers;
  } catch (error) {
    logError('Error getting transactions', error);
    return 0;
  }
}

async function main() {
  console.log(`Using explorer API ${baseUrl}`);
  await getBalance(address);
  await getLYKTransactions(address);
}

main().catch((error) => {
  console.error('Fatal error running example:', error);
});
