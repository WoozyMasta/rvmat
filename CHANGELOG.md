# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog][],
and this project adheres to [Semantic Versioning][].

<!--
## Unreleased

### Added
### Changed
### Removed
-->

## [0.3.0][] - 2026-03-04

### Added

* TexGen inheritance resolution APIs for stage-effective UV access:
  `ResolveStageTexGen`, `EffectiveUVSource`, `EffectiveUVTransform`
* Package-level normalization API:
  `Normalize`, `NormalizeOptions`, `NormalizeResult`.
* Material generation APIs: `Generate`, `GenerateDamage`,
  `GenerateDestruct`, `GenerateSet`, and `WriteGenerateSet`.
* High-level generation options for base-material profiles, stage overrides,
  texture auto-fill, compact TexGen output, and damage/destruct variants.
* Generator material profile catalog for common surface families, used as
  baseline seeds for generated `Super` materials.

### Changed

* Validation resolves TexGen inheritance, reports broken references/cycles,
  and can apply optional shader profile checks (`Super`, `Multi`, `Glass`).
* Parser and writer enforce canonical `.rvmat` key spelling `emmisive[]`,
  while API field naming remains `Emissive`.
* Writer output normalizes texture paths to game-style form and serializes
  float values without long IEEE tails.

[0.3.0]: https://github.com/WoozyMasta/rvmat/compare/v0.2.0...v0.3.0

## [0.2.0][] - 2026-02-26

### Changed

* json and yaml tags are now snake_case instead of camelCase

[0.2.0]: https://github.com/WoozyMasta/rvmat/compare/v0.1.0...v0.2.0

## [0.1.0][] - 2026-02-07

### Added

* First public release

[0.1.0]: https://github.com/WoozyMasta/rvmat/tree/v0.1.0

<!--links-->
[Keep a Changelog]: https://keepachangelog.com/en/1.1.0/
[Semantic Versioning]: https://semver.org/spec/v2.0.0.html
