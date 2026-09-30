# IoT Rule Engine Plugin SDK public-release readiness audit

- Audit date: 2026-09-09
- Scope: public Go API, RPC transport, contract governance, documentation, CI, and repository metadata
- Release verdict: **technically close, administratively blocked**

## Executive assessment

The SDK is the strongest of the three repositories. It has a machine-readable v1 contract, exported API
baseline, compatibility policy, race tests, transport tests, examples, and generated-document gates. The
remaining blockers are public repository/tag sequencing, release provenance, AGPL ecosystem guidance, and
security automation rather than a known correctness defect.

## Findings

| ID | Severity | Finding | Required action before public `v1.0.0` |
| --- | --- | --- | --- |
| SDK-OS-001 | Resolved | The SDK license was finalized on 2026-09-09. | `LICENSE` and both README languages now identify `AGPL-3.0-only`; publish clear integration guidance because SDK copyleft obligations can affect adopters. |
| SDK-OS-002 | P1 | The remote still references a local legacy clone, while engine/plugin CI expects `punk-one/iot-rule-plugin-sdk@v1.0.0` on GitHub. | Create the public SDK repository first, protect the default branch, then publish a signed `v1.0.0` tag after clean-clone verification. |
| SDK-OS-003 | P1 | A three-year compatibility promise creates an operational support obligation but no public support-window calendar or deprecation ledger exists. | Publish support dates, supported Go versions, security-fix policy, and a deprecation registry. |
| SDK-OS-004 | P1 | CI lacks dependency review, secret scanning, CodeQL/static security analysis, and release artifact attestation. | Add least-privilege security workflows and signed provenance for tags. |
| SDK-OS-005 | P2 | Compatibility generation is strong, but contributor changes could still alter semantics without changing signatures. | Require behavioral contract tests and a compatibility checklist in pull requests. |
| SDK-OS-006 | P2 | The worktree contains substantial uncommitted v1 migration work. | Split into reviewable commits and generate the final baseline only from a clean release commit. |

## Verified strengths

- Only one Go module and a deliberately small direct dependency surface.
- CI runs unit, contract, race, vet, example build, and generated-baseline checks.
- `contract_v1.json`, `public_api.json`, and `PUBLIC_API.md` make accidental API drift visible.
- Event ownership and connector resource behavior have black-box tests.
- No known deployment credential literal or secret file type was found in the working tree scan.
- Public API is separated from internal transport and runtime implementation.

## Release order

1. Verify copyright ownership and dependency compatibility; publish SDK-specific AGPL integration guidance.
2. Run full-history secret and author/provenance review.
3. Complete clean-clone tests on every supported Go version.
4. Publish repository governance, support window, security policy, and contribution terms.
5. Create a signed `v1.0.0` tag and immutable release notes.
6. Update engine and plugin CI to consume that public tag, then publish the dependent repositories.

## Verification snapshot

The 2026-09-09 working tree passed `go test ./...`, `go test -race ./...`, `go vet ./...`, and
`go build ./examples/...`. These results validate the current tree only; repeat them from a clean clone on
every Go version declared in the final support policy before signing `v1.0.0`.

## Licensing note

The SDK now uses `AGPL-3.0-only`, consistently with the engine and official plugins. This maximizes license
consistency but is intentionally not a permissive SDK model. Plugin and application authors should review
AGPL obligations before incorporating, modifying, or distributing the SDK; the AGPL permits commercial use.
