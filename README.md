# LayerK Blockchain

## Technical Documentation

### Introduction
LayerK is a blockchain platform that provides a scalable, efficient, and secure environment for decentralized applications (dApps). This documentation provides an in-depth overview of the LayerK system architecture, functionalities, and how it integrates Arbitrum Nitro technology to enhance its performance.

### Table of Contents
1. [Overview](#1-overview)
2. [System Architecture](#2-system-architecture)
    - [LayerK Nodes](#21-layerk-nodes)
    - [Consensus Mechanism](#22-consensus-mechanism)
    - [Smart Contract Execution](#23-smart-contract-execution)
3. [Core Components](#3-core-components)
    - [LayerK Nodes](#31-layerk-nodes)
    - [Consensus Mechanism](#32-consensus-mechanism)
    - [Smart Contract Execution](#33-smart-contract-execution)
4. [Inbox/Outbox Mechanism in LayerK](#4-inboxoutbox-mechanism-in-layerk)
    - [Inbox Overview](#41-inbox-overview)
    - [Outbox Overview](#42-outbox-overview)
    - [Data Flow Between Inbox and Outbox](#43-data-flow-between-inbox-and-outbox)
5. [Arbitrum Nitro Integration](#5-arbitrum-nitro-integration)
6. [Bridge Contracts](#6-bridge-contracts)
    - [Functionality](#61-functionality)
    - [Security](#62-security)
7. [Development and Deployment](#7-development-and-deployment)
    - [Development Environment](#71-development-environment)
    - [Deployment Process](#72-deployment-process)
8. [Security Features](#8-security-features)
9. [Contracts Deployed](#9-contracts-deployed)
10. [Architecture Diagrams](#10-architecture-diagrams)

### 1. Overview
LayerK is designed to offer high throughput, low latency, and reduced gas fees, making it an ideal platform for dApp developers. By utilizing Arbitrum Nitro's proven technology, LayerK ensures compatibility with Ethereum while significantly enhancing performance and scalability.

**Core Innovations:**
- **Sequencing with Deterministic Execution:** Transactions are organized into a single ordered sequence and then executed by a deterministic state transition function.
- **Optimistic Rollup with Interactive Fraud Proofs:** Settles transactions to the Arbitrum chain using an optimistic rollup protocol with interactive fraud proofs.
- **Geth Integration:** Supports Ethereum's data structures, formats, and virtual machine by integrating Geth, ensuring compatibility with Ethereum.
- **Dual Compilation for Execution and Proving:** Compiles source code twice—once to native code for execution and once to WebAssembly (WASM - see layerk-stylus-sdk-wasm) for proving.
  
  <img width="642" alt="image" src="https://github.com/LayerK/LayerK-Blockchain/assets/174424424/4398113e-4155-4e8e-86b7-5bac34bf0335">

### 2. System Architecture

#### 2.1 LayerK Nodes
LayerK nodes are responsible for maintaining the blockchain ledger, processing transactions, and executing smart contracts. Types of nodes:
- **Validator Nodes:** Participate in consensus, proposing and validating new blocks.
- **Full Nodes:** Store the entire blockchain history, providing data to light clients and other nodes.
- **Sequencer:** A designated full node controlling the ordering of transactions.

#### 2.2 Consensus Mechanism
LayerK employs an efficient and secure consensus mechanism similar to Arbitrum Nitro, ensuring consistency and reliability across nodes.

#### 2.3 Smart Contract Execution
Smart contracts are executed in a manner compatible with Ethereum, leveraging Arbitrum Nitro technology for enhanced performance and cost reduction.

### 3. Core Components

#### 3.1 LayerK Nodes
Nodes perform essential functions such as:
- **Transaction Validation:** Ensuring incoming transactions are legitimate.
- **Block Creation:** Validator nodes create and add new blocks to the blockchain.
- **Data Storage:** Full nodes store complete blockchain history.

#### 3.2 Consensus Mechanism
The consensus mechanism involves proposing, voting, and finalizing blocks to ensure network integrity and prevent double-spending.

#### 3.3 Smart Contract Execution
LayerK's execution environment, built on Arbitrum Nitro, allows developers to deploy Ethereum-compatible dApps seamlessly.

### 4. Inbox/Outbox Mechanism in LayerK

#### 4.1 Inbox Overview
The Inbox receives and validates incoming transactions from external sources, queuing them for processing by the LayerK network.

#### 4.2 Outbox Overview
The Outbox handles outgoing messages or transactions, ensuring accurate communication of state changes to external entities.

#### 4.3 Data Flow Between Inbox and Outbox
- **Transaction Initiation:** External systems send transactions to the Inbox.
- **Inbox Processing:** Validates and queues transactions.
- **Transaction Execution:** The network processes transactions, updating the ledger.
- **State Change and Outbox Preparation:** Prepares state changes for dispatch.
- **Outbox Dispatch:** Sends validated transactions to external systems.
- **External Acknowledgment:** External systems process and acknowledge receipt.

### 5. Arbitrum Nitro Integration
LayerK is built on Arbitrum Nitro version 2.3.3, providing scalability, compatibility with Ethereum, and enhanced efficiency.

### 6. Bridge Contracts

#### 6.1 Functionality
Bridge contracts facilitate:
- **Asset Transfer:** Moving tokens and assets between LayerK and connected networks.
- **Data Interchange:** Exchanging data across blockchain boundaries.

#### 6.2 Security
LayerK leverages Arbitrum Nitro's robust security features to protect cross-chain transactions.

### 7. Development and Deployment

#### 7.1 Development Environment
LayerK offers a developer-friendly environment with tools such as SDKs, APIs, and comprehensive documentation.

#### 7.2 Deployment Process
Deploying smart contracts involves:
- Writing in Solidity.
- Compiling and testing.
- Deploying to the LayerK network using familiar Ethereum tools. (also use related test bridges)

### 8. Security Features
LayerK prioritizes security through:
- **Consensus Security:** Resistant consensus mechanism.
- **Smart Contract Audits:** Regular audits to mitigate vulnerabilities.
- **Network Monitoring:** Continuous activity monitoring for threat detection.

### 9. Contracts Deployed [Testnet contracts are different]

#### Core
- **Rollup**: [0x9D96a05467Bb546b50ad32a0d140A4A4706f55F6](https://arbiscan.io/address/0x9D96a05467Bb546b50ad32a0d140A4A4706f55F6)
- **Inbox**: [0x108964DDACAc3420Fe46c291031a59721dFFA637](https://arbiscan.io/address/0x108964DDACAc3420Fe46c291031a59721dFFA637)
- **Outbox**: [0xD6a1337bEC237D0BaFC8d4801463FA0e15Db481C](https://arbiscan.io/address/0xD6a1337bEC237D0BaFC8d4801463FA0e15Db481C)
- **AdminProxy**: [0xaa363224cE8053d464e64328B57Bce01f1c0270b](https://arbiscan.io/address/0xaa363224cE8053d464e64328B57Bce01f1c0270b)
- **SequencerInbox**: [0xA59c2F62C0d53fe4439df7b4B995b768d9088b9d](https://arbiscan.io/address/0xA59c2F62C0d53fe4439df7b4B995b768d9088b9d)
- **Bridge**: [0xe8c495583789A49b90d2DE3021bCc5bf42673F89](https://arbiscan.io/address/0xe8c495583789A49b90d2DE3021bCc5bf42673F89)
- **Utils**: [0x6c21303F5986180B1394d2C89f3e883890E2867b](https://arbiscan.io/address/0x6c21303F5986180B1394d2C89f3e883890E2867b)
- **ValidatorWalletCreator**: [0x2b0E04Dc90e3fA58165CB41E2834B44A56E766aF](https://arbiscan.io/address/0x2b0E04Dc90e3fA58165CB41E2834B44A56E766aF)
- **L3UpgradeExecutor**: [0x9a94627ca2ea9e72eF20D919C8A2AF9f205A89fC](https://arbiscan.io/address/0x9a94627ca2ea9e72eF20D919C8A2AF9f205A89fC)

#### Token Bridge (Arbitrum One - Mainnet)
- **CustomGateway**: [0x7e974088bAF57DC80f808DBF8E8112fb52D2D532](https://arbiscan.io/address/0x7e974088bAF57DC80f808DBF8E8112fb52D2D532)
- **Multicall**: [0x90B02D9F861017844F30dFbdF725b6aa84E63822](https://arbiscan.io/address/0x90B02D9F861017844F30dFbdF725b6aa84E63822)
- **ProxyAdmin**: [0xaa363224cE8053d464e64328B57Bce01f1c0270b](https://arbiscan.io/address/0xaa363224cE8053d464e64328B57Bce01f1c0270b)
- **Router**: [0xE9E6e749f76858E6A6EdcD1EE737D4A58b183CAb](https://arbiscan.io/address/0xE9E6e749f76858E6A6EdcD1EE737D4A58b183CAb)
- **StandardGateway**: [0x0f6Ac4f57Aa2053c17B3893b12c317248670D7F6](https://arbiscan.io/address/0x0f6Ac4f57Aa2053c17B3893b12c317248670D7F6)

#### Token Bridge (LayerK - Mainnet)
- **CustomGateway**: [0x60Dfa13eB86d30fBCd113923E1EA71c060e897e6](https://explorer.layerk.com/address/0x60Dfa13eB86d30fBCd113923E1EA71c060e897e6)
- **Multicall**: [0xfFB66a127c78e93A1970BbcD44ae7B39f4c04dbB](https://explorer.layerk.com/address/0xfFB66a127c78e93A1970BbcD44ae7B39f4c04dbB)
- **ProxyAdmin**: [0x0Da76E552AC613261f7CCE12CedD960De719Ca86](https://explorer.layerk.com/address/0x0Da76E552AC613261f7CCE12CedD960De719Ca86)
- **Router**: [0x7be9EF89F0fe4128aafc81c8fa827026A6fE44Cf](https://explorer.layerk.com/address/0x7be9EF89F0fe4128aafc81c8fa827026A6fE44Cf)
- **StandardGateway**: [0x0A5299d90A4E3928B7c4e3298055c87Db1b54B52](https://explorer.layerk.com/address/0x0A5299d90A4E3928B7c4e3298055c87Db1b54B52)

### 10. Architecture Diagrams

- **LayerK Detailed**: ![LayerK Tech](https://github.com/LayerK/LayerK-Blockchain/blob/main/diagrams/layerk-tech.svg)
- **LayerK Challenge**: ![LayerK Challenge](https://github.com/LayerK/LayerK-Blockchain/blob/main/diagrams/Layerk-challenge.png)
- **Bridging Diagram**: ![Bridging Diagram](https://github.com/LayerK/LayerK-Blockchain/blob/main/diagrams/bridging-diagram.svg)
