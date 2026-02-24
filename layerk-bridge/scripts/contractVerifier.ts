import { spawn } from 'child_process'

const VERIFY_TIMEOUT_MS = 120_000
export class ContractVerifier {
  chainId: number
  apiKey: string = ''

  readonly NUM_OF_OPTIMIZATIONS = 100
  readonly COMPILER_VERSION = '0.8.16'

  ///// List of contract addresses and their corresponding source code files
  readonly TUP =
    'node_modules/@openzeppelin/contracts/proxy/transparent/TransparentUpgradeableProxy.sol:TransparentUpgradeableProxy'
  readonly PROXY_ADMIN =
    'node_modules/@openzeppelin/contracts/proxy/transparent/ProxyAdmin.sol:ProxyAdmin'
  readonly EXECUTOR =
    'node_modules/@offchainlabs/upgrade-executor/src/UpgradeExecutor.sol:UpgradeExecutor'

  readonly contractToSource = {
    l1TokenBridgeCreatorProxyAdmin: this.PROXY_ADMIN,
    l1TokenBridgeCreatorLogic:
      'contracts/tokenbridge/ethereum/L1AtomicTokenBridgeCreator.sol:L1AtomicTokenBridgeCreator',
    l1TokenBridgeCreatorProxy: this.TUP,
    retryableSenderLogic:
      'contracts/tokenbridge/ethereum/L1TokenBridgeRetryableSender.sol:L1TokenBridgeRetryableSender',
    retryableSenderProxy: this.TUP,
    routerTemplate:
      'contracts/tokenbridge/ethereum/gateway/L1GatewayRouter.sol:L1GatewayRouter',
    standardGatewayTemplate:
      'contracts/tokenbridge/ethereum/gateway/L1ERC20Gateway.sol:L1ERC20Gateway',
    customGatewayTemplate:
      'contracts/tokenbridge/ethereum/gateway/L1CustomGateway.sol:L1CustomGateway',
    wethGatewayTemplate:
      'contracts/tokenbridge/ethereum/gateway/L1WethGateway.sol:L1WethGateway',
    feeTokenBasedRouterTemplate:
      'contracts/tokenbridge/ethereum/gateway/L1OrbitGatewayRouter.sol:L1OrbitGatewayRouter',
    feeTokenBasedStandardGatewayTemplate:
      'contracts/tokenbridge/ethereum/gateway/L1OrbitERC20Gateway.sol:L1OrbitERC20Gateway',
    feeTokenBasedCustomGatewayTemplate:
      'contracts/tokenbridge/ethereum/gateway/L1OrbitCustomGateway.sol:L1OrbitCustomGateway',
    upgradeExecutor: this.EXECUTOR,
    l2TokenBridgeFactoryOnL1:
      'contracts/tokenbridge/arbitrum/L2AtomicTokenBridgeFactory.sol:L2AtomicTokenBridgeFactory',
    l2GatewayRouterOnL1:
      'contracts/tokenbridge/arbitrum/gateway/L2GatewayRouter.sol:L2GatewayRouter',
    l2StandardGatewayAddressOnL1:
      'contracts/tokenbridge/arbitrum/gateway/L2ERC20Gateway.sol:L2ERC20Gateway',
    l2CustomGatewayAddressOnL1:
      'contracts/tokenbridge/arbitrum/gateway/L2CustomGateway.sol:L2CustomGateway',
    l2WethGatewayAddressOnL1:
      'contracts/tokenbridge/arbitrum/gateway/L2WethGateway.sol:L2WethGateway',
    l2WethAddressOnL1: 'contracts/tokenbridge/libraries/aeWETH.sol:aeWETH',
    l2MulticallAddressOnL1: 'contracts/rpc-utils/MulticallV2.sol:ArbMulticall2',
    l1Multicall: 'contracts/rpc-utils/MulticallV2.sol:Multicall2',
  }

  constructor(chainId: number, apiKey: string) {
    this.chainId = chainId
    if (apiKey) {
      this.apiKey = apiKey
    }
  }

  async verifyWithAddress(
    name: string,
    contractAddress: string,
    constructorArgs?: string,
    _numOfOptimization?: number
  ) {
    if (!this.apiKey) {
      throw new Error('Missing explorer API key for contract verification')
    }
    if (!/^0x[a-fA-F0-9]{40}$/.test(contractAddress)) {
      throw new Error(`Invalid contract address: ${contractAddress}`)
    }

    // avoid rate limiting
    await new Promise(resolve => setTimeout(resolve, 1000))

    const sourceFile =
      this.contractToSource[name as keyof typeof this.contractToSource]
    if (!sourceFile) {
      throw new Error(`Unknown contract key: ${name}`)
    }

    const args = [
      'verify-contract',
      '--chain-id',
      String(this.chainId),
      '--compiler-version',
      this.COMPILER_VERSION,
      '--num-of-optimizations',
      String(
        _numOfOptimization !== undefined
          ? _numOfOptimization
          : this.NUM_OF_OPTIMIZATIONS
      ),
    ]

    if (constructorArgs) {
      args.push('--constructor-args', constructorArgs)
    }
    args.push(
      contractAddress,
      sourceFile,
      '--etherscan-api-key',
      this.apiKey
    )

    const safeCommand = `forge ${args
      .map(arg => (arg === this.apiKey ? '[REDACTED]' : arg))
      .join(' ')}`

    await new Promise<void>((resolve, reject) => {
      const child = spawn('forge', args, { stdio: ['ignore', 'pipe', 'pipe'] })
      let stdout = ''
      let stderr = ''
      let timedOut = false
      let settled = false

      const finish = (handler: () => void) => {
        if (settled) return
        settled = true
        handler()
      }

      const timeout = setTimeout(() => {
        timedOut = true
        child.kill('SIGKILL')
      }, VERIFY_TIMEOUT_MS)

      child.stdout.on('data', chunk => {
        stdout += chunk.toString()
      })
      child.stderr.on('data', chunk => {
        stderr += chunk.toString()
      })

      child.once('error', err => {
        if (settled) return
        clearTimeout(timeout)
        console.log('-----------------')
        console.log(safeCommand)
        console.log('Failed to submit for verification', contractAddress, err)
        finish(() => reject(err))
      })

      child.once('close', code => {
        if (settled) return
        clearTimeout(timeout)
        console.log('-----------------')
        console.log(safeCommand)
        if (stdout.trim()) {
          console.log(stdout.trim())
        }
        if (timedOut) {
          const err = new Error(
            `Verification timed out after ${VERIFY_TIMEOUT_MS}ms for ${contractAddress}`
          )
          console.log(err.message)
          finish(() => reject(err))
          return
        }
        if (code !== 0) {
          const err = new Error(
            `Failed to submit for verification ${contractAddress}: ${stderr.trim()}`
          )
          console.log(err.message)
          finish(() => reject(err))
          return
        }

        console.log('Successfully submitted for verification', contractAddress)
        finish(() => resolve())
      })
    })
  }
}
