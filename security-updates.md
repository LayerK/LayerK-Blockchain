# Security Update Recommendations

Prepared for the LayerK Blockchain mono-repo to highlight near-term security work items grounded in the current code base.

## 1. Contract dependency locks aligned (landed)
- **Status:** `layerk-nitro/package.json` and `layerk-bridge/package.json` both request `@openzeppelin/contracts@^4.9.5` and `@openzeppelin/contracts-upgradeable@^4.9.5`. Their Yarn locks now include direct `4.9.5` resolutions, and Nitro no longer carries unused OpenZeppelin `4.5.x` lock entries.
- **Remaining risk:** `layerk-bridge` still inherits `@openzeppelin/contracts@4.5.0` and `@openzeppelin/contracts-upgradeable@4.5.2` transitively from `@arbitrum/nitro-contracts@2.1.0`. That cannot be removed safely from the bridge lock until the upstream package stops declaring those exact transitive versions or the bridge is moved to the local Nitro package.
- **Follow-up actions:**
  - Re-run `yarn build`, Foundry tests, and formal tools before publishing new artifacts.
  - Track an upstream `@arbitrum/nitro-contracts` release that no longer pins OpenZeppelin `4.5.x`, or switch the bridge to a verified local package reference.
  - Keep `layerk-nitro/audit-ci.jsonc` limited to advisories that remain unavoidable after dependency refreshes.

## 2. Dependency update automation enabled (landed)
- **Status:** A root `.github/dependabot.yml` now covers active npm packages, Stylus Cargo packages, and bridge/Nitro submodules against `main`. The stale bridge-local Dependabot config was reduced to the package paths that exist if `layerk-bridge` is used standalone.
- **Risk reduced:** Dependency drift and security advisories are more likely to surface as normal pull requests instead of staying hidden in stale lockfiles or removed directories.
- **Follow-up actions:**
  - Review the first Dependabot PRs for grouping behavior and noise.
  - Add GitHub Actions coverage at the repo root if this mono-repo should run CI directly from GitHub.

## 3. Harden AbsInbox allow-list enforcement (landed - follow-ups only)
- **Status:** The single `onlyAllowed` modifier was split into `onlyAllowedOrigin` and `onlyAllowedSender` in `layerk-nitro/src/bridge/AbsInbox.sol`. Allow-listed EOAs can no longer be phished through an arbitrary forwarding contract on sender-gated entrypoints.
- **Remaining follow-ups:**
  - Add Foundry tests that exercise both an EOA caller and a contract caller against each allow-list gated entrypoint, plus a regression test that the `sendL2MessageFromOrigin` path still rejects non-codeless origins.
  - Before enabling the allow list in production, migrate any existing EOA entries that actually operate through a forwarding contract, such as multisig or account abstraction, onto their contract addresses.
  - Update integrator docs so they know the allow list is keyed on `msg.sender`; contracts must be listed explicitly.

## 4. Deprecated address monitors partially hardened
- **Status:** The monitoring utilities under `layerk-api` remain in the deprecated subtree, but the maintained implementations now use configurable confirmation depth and bounded queues. The Go monitor also caps successful JSON-RPC response bodies with `MAX_RPC_RESPONSE_BYTES`, and the JavaScript monitor drains queued blocks without repeated array shifts.
- **Risk:** If those monitors are still used operationally, they still alert through stdout without durable storage, authenticated delivery, or RPC failover.
- **Follow-up actions:**
  - Confirm whether the deprecated monitors are still deployed anywhere.
  - If active, move maintained monitor code out of `layerk-api` and add durable replay state, signed alerts, and RPC failover.

## 5. Merkle tree deserialization hardened (landed)
- **Status:** `NewMerkleTreeFromReader` now reads the serialized node-type byte with `io.ReadFull` instead of a single `Read` call, and the Merkle accumulator tests cover readers that temporarily report no progress.
- **Risk reduced:** A short or zero-progress read can no longer be interpreted as a valid leaf node type before the serialized stream has actually supplied the node discriminator.
- **Follow-up actions:**
  - Re-run the Merkle tree Go tests in CI before shipping, since local Go execution is intentionally skipped in this workspace.
