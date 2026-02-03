import { task } from "hardhat/config";
import { ethers } from "ethers";
import "@nomiclabs/hardhat-etherscan"
import { Bytecode } from "@nomiclabs/hardhat-etherscan/dist/src/solc/bytecode"
import { TASK_VERIFY_GET_CONTRACT_INFORMATION, TASK_VERIFY_GET_COMPILER_VERSIONS, TASK_VERIFY_GET_LIBRARIES } from "@nomiclabs/hardhat-etherscan/dist/src/constants"
import fs from "fs";

task("compareBytecode", "Compares deployed bytecode with local builds")
    .addParam("contractAddrs", "A comma-separated list of deployed contract addresses")
    .setAction(async ({ contractAddrs }, hre) => {
        const addresses = contractAddrs.split(',');

        // Build a lookup map of deployed bytecode hashes to contract names.
        const artifactPaths = await hre.artifacts.getArtifactPaths();
        const localCodeHashes = new Map<string, string[]>();
        for (const artifactPath of artifactPaths) {
            const artifact = JSON.parse(fs.readFileSync(artifactPath, "utf8"));
            const deployedBytecode = artifact.deployedBytecode;
            if (typeof deployedBytecode !== "string" || deployedBytecode.length <= 2) {
                continue;
            }
            const localCodeHash = ethers.utils.keccak256(deployedBytecode);
            const existing = localCodeHashes.get(localCodeHash);
            if (existing) {
                existing.push(artifact.contractName);
            } else {
                localCodeHashes.set(localCodeHash, [artifact.contractName]);
            }
        }

        let cachedCompilerVersions: string[] | undefined;
        let cachedLibraries: Record<string, string> | undefined;

        for (const contractAddr of addresses) {
            const trimmed = contractAddr.trim();
            if (!trimmed) {
                continue;
            }

            // Fetch deployed contract bytecode
            const deployedBytecode = await hre.ethers.provider.getCode(trimmed);
            if (deployedBytecode === "0x") {
                console.log(`No bytecode found at address ${trimmed}`);
                continue;
            }

            const deployedCodeHash = ethers.utils.keccak256(deployedBytecode);
            const matches = localCodeHashes.get(deployedCodeHash);
            if (matches && matches.length > 0) {
                console.log(
                    `Contract Address ${trimmed} matches with ${matches.join(", ")}`
                );
                continue;
            }

            const deployedBytecodeHex = deployedBytecode.startsWith("0x")
                ? deployedBytecode.slice(2)
                : deployedBytecode;
            try {
                if (!cachedCompilerVersions) {
                    cachedCompilerVersions = await hre.run(
                        TASK_VERIFY_GET_COMPILER_VERSIONS
                    );
                }
                if (!cachedLibraries) {
                    cachedLibraries = await hre.run(TASK_VERIFY_GET_LIBRARIES);
                }
                const info = await hre.run(TASK_VERIFY_GET_CONTRACT_INFORMATION, {
                    deployedBytecode: new Bytecode(deployedBytecodeHex),
                    matchingCompilerVersions: cachedCompilerVersions,
                    libraries: cachedLibraries,
                })
                console.log(
                    `Contract Address ${trimmed} matches with ${info.contractName} without checking constructor arguments`
                );
            } catch (error) {
                console.log(`No matching contract found for address ${trimmed}`);
            }
        }
    });

export default {};
