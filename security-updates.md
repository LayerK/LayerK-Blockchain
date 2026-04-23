# Security Update Recommendations

Prepared for the LayerK Blockchain mono-repo to highlight near-term security work items grounded in the current code base.

## 1. Refresh Nitro contract dependencies and remove suppressed CVEs
- **Evidence:** `layerk-nitro/package.json` still pins `@openzeppelin/contracts` to `4.5.0` and `@openzeppelin/contracts-upgradeable` to `4.5.2` (`layerk-nitro/package.json:49-55`), while `layerk-nitro/audit-ci.jsonc` suppresses 20+ OpenZeppelin GitHub Security Advisories (for example `GHSA-4g63-c64m-25w9`, `GHSA-mx2q-35m2-x2rh`, `GHSA-9vx6-7xxf-x967`).
- **Risk:** Those versions pre-date the 2023–2024 fixes for SignatureChecker, TransparentUpgradeableProxy selector clashes, Base64 memory safety, etc. Keeping them plus suppressing the alerts means Nitro contracts continue to ship with known bugs that an attacker can exploit on-chain.
- **Actions:**
  - Upgrade the Nitro contracts workspace to `@openzeppelin/contracts@^4.9.5` and `@openzeppelin/contracts-upgradeable@^4.9.5`, keeping `openzeppelin-upgrades` tooling in sync.
  - Re-run `yarn build`, Foundry tests, and formal tools (Slither/Manticore) to validate storage compatibility.
  - Trim `layerk-nitro/audit-ci.jsonc` so that only unfixed vulnerabilities stay on the allowlist; the bulk of the OpenZeppelin advisories should disappear once the dependency bump lands.
  - Document any remaining suppressed advisories (if unavoidable) with justification per GHSA entry.

## 2. Synchronize token-bridge builds with the Nitro baseline
- **Evidence:** `layerk-bridge/package.json` depends on `@arbitrum/nitro-contracts@^1.1.1` (`layerk-bridge/package.json:43-48`), whereas the in-repo Nitro workspace advertises `version: "2.1.0"` (`layerk-nitro/package.json:2-4`).
- **Risk:** The L1/L2 token bridge can compile against a different contract set than the one getting deployed to LayerK. That mismatch leaves the bridge without the latest rollup core fixes (BOLD pre-verifier, governor hardening, etc.) and invalidates audit coverage because the versions no longer align.
- **Actions:**
  - Point `layerk-bridge` at the local Nitro package (e.g., Yarn workspace reference) or bump to `@arbitrum/nitro-contracts@^2.1.0`.
  - Rebuild bridge artifacts and re-run `hardhat test`, Foundry suites, and the `test-e2e` scenarios against the upgraded dependency.
  - Update deployment scripts plus `_deployments` manifests to reflect the new bytecode hashes.

## 3. Harden AbsInbox allow-list enforcement (landed – follow-ups only)
- **Status:** The single `onlyAllowed` modifier was split into `onlyAllowedOrigin` (used only by `sendL2MessageFromOrigin`, which already enforces `msg.sender == tx.origin` via `CallerChecker.isCallerCodelessOrigin()`) and `onlyAllowedSender` (used everywhere else, keying the allow list off `msg.sender`) in `layerk-nitro/src/bridge/AbsInbox.sol`. All 13 Inbox / ERC20Inbox call sites now use `onlyAllowedSender`, so allow-listed EOAs can no longer be phished through an arbitrary forwarding contract.
- **Remaining follow-ups:**
  - Add Foundry tests that exercise both an EOA caller and a contract caller against each allow-list gated entrypoint, plus a regression test that the `sendL2MessageFromOrigin` path still rejects non-codeless origins.
  - Before enabling the allow list in production, migrate any existing EOA entries that actually operate through a forwarding contract (e.g., multisig, account abstraction) onto their contract addresses — otherwise enabling the list will now block them.
  - Update integrator docs so they know the allow list is keyed on `msg.sender` (contracts must be listed explicitly); this is a behavior change for anyone relying on the previous `tx.origin` semantics.

## 4. Add finality and authenticated alerting to the address monitors
- **Evidence:** Both monitoring utilities stream every new LayerK block and immediately trust what they see (`layerk-api/address-monitor/monitor.js:3-50` and `layerk-api/address-monitor-go/monitor.go:20-210`). They only print to stdout; there is no confirmation depth, persistence, or signed notification.
- **Risk:** A reorg (or RPC spoofing) can feed false transactions into the monitor, causing alert fatigue or missing the actual exploit. Without webhook signatures or TLS pinning, an attacker who controls DNS/RPC can suppress or forge alerts.
- **Actions:**
  - Require ≥ N confirmations (configurable) before treating a block as final, and cache processed block hashes to resist replays.
  - Emit alerts through authenticated channels (e.g., signed webhooks, PagerDuty, SIEM ingestion) instead of stdout, and store events durably.
  - Add RPC failover plus certificate pinning (Go’s `crypto/x509` roots or Ethers custom `fetch` override) so the monitor cannot be silently pointed to a malicious endpoint.
  - Externalize the monitored address set to a signed JSON or contract registry so tampering is detectable.
