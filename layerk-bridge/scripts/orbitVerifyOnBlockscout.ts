import { ethers } from 'hardhat'
import { run } from 'hardhat'
import {
  AeWETH__factory,
  BeaconProxyFactory__factory,
  L1AtomicTokenBridgeCreator__factory,
  UpgradeableBeacon__factory,
} from '../build/types'
import { Provider } from '@ethersproject/providers'
import {
  abi as UpgradeExecutorABI,
  bytecode as UpgradeExecutorBytecode,
} from '@offchainlabs/upgrade-executor/build/contracts/src/UpgradeExecutor.sol/UpgradeExecutor.json'

const IMPLEMENTATION_STORAGE_SLOT =
  '0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc'

type VerificationOptions = {
  address: string
  constructorArguments: readonly unknown[]
  contract?: string
}

main()
  .then(() => console.log('Done.'))
  .catch((error: unknown) => {
    console.error(getErrorMessage(error))
    process.exitCode = 1
  })

function requireEnv(name: string): string {
  const value = process.env[name]?.trim()
  if (!value) {
    throw new Error(`Missing required env var ${name}`)
  }
  return value
}

function optionalEnv(name: string): string | undefined {
  const value = process.env[name]?.trim()
  return value ? value : undefined
}

function getErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message
  }
  return String(error)
}

async function main() {
  const parentRpcUrl = requireEnv('BASECHAIN_RPC')
  const tokenBridgeCreatorAddress = requireEnv('L1_TOKEN_BRIDGE_CREATOR')
  const inboxAddress = requireEnv('INBOX_ADDRESS')
  const deployerKey = optionalEnv('DEPLOYER_KEY')

  if (!deployerKey) {
    console.log(
      'DEPLOYER_KEY is missing. Deployer key is required if you want to have aeWETH and UpgradeExecutor verified.'
    )
  }

  const parentProvider = new ethers.providers.JsonRpcProvider(parentRpcUrl)
  const orbitProvider = ethers.provider
  const deployerOnOrbit = deployerKey
    ? new ethers.Wallet(deployerKey, orbitProvider)
    : undefined

  /// collect addresses
  const tokenBridgeCreator = L1AtomicTokenBridgeCreator__factory.connect(
    tokenBridgeCreatorAddress,
    parentProvider
  )
  const l2Factory = await tokenBridgeCreator.canonicalL2FactoryAddress()
  const l2Deployment = await tokenBridgeCreator.inboxToL2Deployment(
    inboxAddress
  )
  const beaconProxyFactory = BeaconProxyFactory__factory.connect(
    l2Deployment.beaconProxyFactory,
    orbitProvider
  )
  const upgradeableBeacon = UpgradeableBeacon__factory.connect(
    await beaconProxyFactory.beacon(),
    orbitProvider
  )
  const standardArbERC20 = await upgradeableBeacon.implementation()

  console.log(
    'Start verification of token bridge contracts deployed to chain',
    (await orbitProvider.getNetwork()).chainId
  )

  // verify L2 factory
  await _verifyContract('L2AtomicTokenBridgeFactory', l2Factory, [])

  // verify single TUP, others TUPs will be verified by bytecode match
  await _verifyContract('TransparentUpgradeableProxy', l2Deployment.router, [
    l2Factory,
    l2Deployment.proxyAdmin,
    '0x',
  ])

  // verify orbit contracts
  await _verifyContract(
    'L2GatewayRouter',
    await _getLogicAddress(l2Deployment.router, orbitProvider),
    []
  )
  await _verifyContract(
    'L2ERC20Gateway',
    await _getLogicAddress(l2Deployment.standardGateway, orbitProvider),
    []
  )
  await _verifyContract(
    'L2CustomGateway',
    await _getLogicAddress(l2Deployment.customGateway, orbitProvider),
    []
  )
  await _verifyContract(
    'L2WethGateway',
    await _getLogicAddress(l2Deployment.wethGateway, orbitProvider),
    []
  )
  await _verifyContract('BeaconProxyFactory', beaconProxyFactory.address, [])
  await _verifyContract('UpgradeableBeacon', upgradeableBeacon.address, [
    standardArbERC20,
  ])
  await _verifyContract('StandardArbERC20', standardArbERC20, [])
  await _verifyContract('ArbMulticall2', l2Deployment.multicall, [])
  await _verifyContract('ProxyAdmin', l2Deployment.proxyAdmin, [])

  /// special cases - aeWETH and UpgradeExecutor

  if (deployerOnOrbit) {
    // deploy dummy aeWETH and verify it. Its deployed bytecode will match the actual aeWETH bytecode
    const dummyAeWethFac = await new AeWETH__factory(deployerOnOrbit).deploy()
    const dummyAeWeth = await dummyAeWethFac.deployed()
    await _verifyContract('aeWETH', dummyAeWeth.address, [])

    // deploy dummy UpgradeExecutor and verify it. Its deployed bytecode will match the actual UpgradeExecutor bytecode
    const dummyUpgradeExecutorFac = new ethers.ContractFactory(
      UpgradeExecutorABI,
      UpgradeExecutorBytecode,
      deployerOnOrbit
    )
    const dummyUpgradeExecutor = await dummyUpgradeExecutorFac.deploy()
    await dummyUpgradeExecutor.deployed()
    await _verifyContract('UpgradeExecutor', dummyUpgradeExecutor.address, [])
  }
}

async function _verifyContract(
  contractName: string,
  contractAddress: string,
  constructorArguments: readonly unknown[] = [],
  contractPathAndName?: string // optional
): Promise<void> {
  try {
    const verificationOptions: VerificationOptions = {
      address: contractAddress,
      constructorArguments,
    }

    if (contractPathAndName) {
      verificationOptions.contract = contractPathAndName
    }

    await run('verify:verify', verificationOptions)
    console.log(`Verified contract ${contractName} successfully.`)
  } catch (error: unknown) {
    const message = getErrorMessage(error)
    if (message.includes('Already Verified')) {
      console.log(`Contract ${contractName} is already verified.`)
    } else {
      console.error(
        `Verification for ${contractName} failed with the following error: ${message}`
      )
    }
  }
}

async function _getLogicAddress(
  contractAddress: string,
  provider: Provider
): Promise<string> {
  return (
    await _getAddressAtStorageSlot(
      contractAddress,
      provider,
      IMPLEMENTATION_STORAGE_SLOT
    )
  ).toLowerCase()
}

async function _getAddressAtStorageSlot(
  contractAddress: string,
  provider: Provider,
  storageSlotBytes: string
): Promise<string> {
  const storageValue = await provider.getStorageAt(
    contractAddress,
    storageSlotBytes
  )

  if (storageValue === ethers.constants.HashZero) {
    throw new Error(
      `Storage slot ${storageSlotBytes} is empty for contract ${contractAddress}`
    )
  }

  const formatAddress = `0x${storageValue.slice(-40)}`

  return ethers.utils.getAddress(formatAddress)
}
