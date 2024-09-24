## LayerK Blockchain API Documentation

Welcome to the LayerK Blockchain API documentation. This guide provides an overview of all available endpoints and how to interact with them.

**Base URL:**  

All API requests must be prefixed with the base URL:

```

https://explorer.layerk.com/api/v2

```

---

### Table of Contents

- [Search](#1-search)

- [Search Redirect](#2-search-redirect)

- [Get Transactions](#3-get-transactions)

- [Get Blocks](#4-get-blocks)

- [Main Page Transactions](#5-main-page-transactions)

- [Main Page Blocks](#6-main-page-blocks)

- [Indexing Status](#7-indexing-status)

- [Stats](#8-stats)

- [Transaction Info](#9-get-transaction-info)

- [Token Transfers](#10-get-token-transfers)

- [Internal Transactions](#11-get-internal-transactions)

- [Address Info](#12-get-address-info)

- [Token Info](#13-get-token-info)

- [Smart Contracts](#14-get-verified-smart-contracts)

---

### Endpoints

---

#### 1. **Search**

- **GET** `/search`

  This endpoint searches for tokens, addresses, blocks, or transactions.

  **Query Parameters:**

  - `q`: Search query (e.g., `USDT`)

  **Response Example:**

  ```json

  {

    "items": [

      {

        "type": "token",

        "name": "Tether USD",

        "address_hash": "0x06953b1E8f60d5F5DaFACB3a51248287Fad95Fa1",

        "symbol": "USDT",

        "holder_count": 10000

      }

    ],

    "next_page_params": {

      "address_hash": "0x052Ad78E3aA0b0F2D3912FD3b50a9a289CF2f7Aa"

    }

  }

  ```

---

#### 2. **Search Redirect**

- **GET** `/search/check-redirect`

  Checks if a query redirects to a specific resource.

  **Query Parameters:**

  - `q`: Search query (e.g., `USDT`)

  **Response Example:**

  ```json

  {

    "type": "address",

    "parameter": "0x06953b1E8f60d5F5DaFACB3a51248287Fad95Fa1",

    "redirect": true

  }

  ```

---

#### 3. **Get Transactions**

- **GET** `/transactions`

  Retrieves transactions filtered by state, type, or method.

  **Query Parameters:**

  - `filter`: State of the transaction (e.g., "pending", "validated").

  - `type`: Transaction type (e.g., "token_transfer").

  - `method`: Method name (e.g., "approve", "transfer").

  **Response Example:**

  ```json

  {

    "items": [

      {

        "block_number": 27170298,

        "tx_hash": "0xa3b401d6f3124c9d1528cd8d4b692f523d86fd88e48c391ffe9c67e4436ae5ca",

        "gas_used": "9814987"

      }

    ],

    "next_page_params": {

      "block_number": 27170298

    }

  }

  ```

---

#### 4. **Get Blocks**

- **GET** `/blocks`

  Retrieves block information.

  **Query Parameters:**

  - `type`: Block type (e.g., "block", "uncle").

  **Response Example:**

  ```json

  {

    "items": [

      {

        "block_number": 27729304,

        "hash": "0xba83e9ba0b43e8d112a07fdab08d53f473d2b7fb0e585bd437ae739933db203e",

        "gas_used": "26618801760"

      }

    ],

    "next_page_params": {

      "block_number": 27729304

    }

  }

  ```

---

#### 5. **Main Page Transactions**

- **GET** `/main-page/transactions`

  Retrieves recent transactions displayed on the main page.

  **Response Example:**

  ```json

  [

    {

      "tx_hash": "0xe38c6772f33edfbd218f59853befe18391cb786f911fb6c0b00ed6dd72ef6e69",

      "block_number": 24816691,

      "timestamp": "2022-10-31T07:18:05.000000Z"

    }

  ]

  ```

---

#### 6. **Main Page Blocks**

- **GET** `/main-page/blocks`

  Retrieves recent blocks displayed on the main page.

  **Response Example:**

  ```json

  [

    {

      "block_number": 17615720,

      "hash": "0xf569ec751152b2f814001fc730f7797aa155e4bc3ba9cb6ba24bc2c8c9468c1a",

      "timestamp": "2023-07-03T20:09:59.000000Z"

    }

  ]

  ```

---

#### 7. **Indexing Status**

- **GET** `/main-page/indexing-status`

  Retrieves the current indexing status of blocks and internal transactions.

  **Response Example:**

  ```json

  {

    "finished_indexing": true,

    "indexed_blocks_ratio": "1.0",

    "indexed_internal_transactions_ratio": "1.0"

  }

  ```

---

#### 8. **Stats**

- **GET** `/stats`

  Retrieves general blockchain statistics.

  **Response Example:**

  ```json

  {

    "total_blocks": "508700",

    "total_addresses": "982340",

    "total_transactions": "1699427",

    "coin_price": "0.00254957"

  }

  ```

---

#### 9. **Get Transaction Info**

- **GET** `/transactions/{transaction_hash}`

  Retrieves detailed information about a specific transaction.

  **Path Parameters:**

  - `transaction_hash`: The transaction hash.

  **Response Example:**

  ```json

  {

    "timestamp": "2022-08-02T07:18:05.000000Z",

    "tx_hash": "0x5d90a9da2b8da402b11bc92c8011ec8a62a2d59da5c7ac4ae0f73ec51bb73368",

    "block": 23484035,

    "status": "ok",

    "method": "transferFrom"

  }

  ```

---

#### 10. **Get Token Transfers**

- **GET** `/transactions/{transaction_hash}/token-transfers`

  Retrieves token transfers for a specific transaction.

  **Path Parameters:**

  - `transaction_hash`: The transaction hash.

  **Query Parameters:**

  - `type`: Token type (e.g., "ERC-20", "ERC-721").

  **Response Example:**

  ```json

  {

    "items": [

      {

        "from": "0xEb533ee5687044E622C69c58B1B12329F56eD9ad",

        "to": "0x06953b1E8f60d5F5DaFACB3a51248287Fad95Fa1",

        "token": {

          "symbol": "USDT",

          "name": "Tether USD",

          "address": "0x06953b1E8f60d5F5DaFACB3a51248287Fad95Fa1"

        },

        "total": "1000"

      }

    ]

  }

  ```

---

#### 11. **Get Internal Transactions**

- **GET** `/transactions/{transaction_hash}/internal-transactions`

  Retrieves internal transactions for a specific transaction.

  **Response Example:**

  ```json

  {

    "items": [

      {

        "from": "0xEb533ee5687044E622C69c58B1B12329F56eD9ad",

        "to": "0x06953b1E8f60d5F5DaFACB3a51248287Fad95Fa1",

        "value": "10000000000000000"

      }

    ]

  }

  ```

---

#### 12. **Get Address Info**

- **GET** `/addresses/{address_hash}`

  Retrieves information about a specific address.

  **Path Parameters:**

  - `address_hash`: Address hash (e.g., `0x06953b1E8f60d5F5DaFACB3a51248287Fad95Fa1`)

  **Response Example:**

  ```json

  {

    "hash": "0x06953b1E8f60d5F5DaFACB3a51248287Fad95Fa1",

    "name": "Tether USD",

    "coin_balance": "1000000"

  }

  ```

---

#### 13. **Get Token Info**

- **GET** `/tokens/{address_hash}`

  Retrieves detailed information about a specific token.

  **Path Parameters:**

  - `address

_hash`: Token contract address (e.g., `0x06953b1E8f60d5F5DaFACB3a51248287Fad95Fa1`)

  **Response Example:**

  ```json

  {

    "name": "Tether USD",

    "symbol": "USDT",

    "address": "0x06953b1E8f60d5F5DaFACB3a51248287Fad95Fa1",

    "total_supply": "10000000",

    "holders": 837494

  }

  ```

---

#### 14. **Get Verified Smart Contracts**

- **GET** `/smart-contracts`

  Retrieves a list of verified smart contracts.

  **Query Parameters:**

  - `q`: Search query for smart contracts (e.g., "proxy").

  - `filter`: Filter by language (e.g., "vyper", "solidity").

  **Response Example:**

  ```json

  {

    "items": [

      {

        "address": "0x06953b1E8f60d5F5DaFACB3a51248287Fad95Fa1",

        "name": "Tether Proxy",

        "language": "solidity",

        "compiler_version": "v0.8.4"

      }

    ]

  }

  ```

---

### Error Handling

- `400`: Bad input parameter

All errors returned by the API follow the format below:

```json

{

  "message": "Error description"

}

```

---

### License

This API is licensed under Apache 2.0. For more details, visit [Apache License 2.0](http://www.apache.org/licenses/LICENSE-2.0.html).

---

This concludes the API documentation for LayerK Blockchain API. For more information, contact via email at `you@your-company.com`.
