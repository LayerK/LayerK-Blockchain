Monitoring transactions involving a configured set of addresses.

## Start

```bash
npm install
npm start
```

`npm start` runs [`monitor.js`](./monitor.js), and [`index.js`](./index.js) is kept as a compatibility wrapper for `require('address-monitor')`.

## Configuration

- `LAYERK_RPC_URL`: RPC endpoint to poll. Defaults to `https://mainnet-rpc.layerk.com`.
- `MONITORED_ADDRESSES`: Comma-separated list of addresses to watch.
- `MIN_CONFIRMATIONS`: Number of confirmations to wait before processing a block. Defaults to `3`.
- `MAX_BLOCK_QUEUE`: Maximum finalized blocks to buffer per poll. Defaults to `32`.
- `PROCESSED_BLOCK_CACHE`: Number of processed finalized blocks to retain in memory. Defaults to `256`.

## Default addresses

The monitor starts with:

- `0xE01B9E7A53629D23ee7571A3cF05C3188883f35e`
- `0xDe96e7Ed414943Ebb73aE64B517166Ad22e39729`

