# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [1.0.0] - 2026-03-08

Community-ready release with pluggable data providers, data portability, CI pipeline, and full documentation.

### Added

- Pluggable data provider system with registry, priority ordering, and automatic fallback ([#45](https://github.com/lugassawan/panen/issues/45), [#121](https://github.com/lugassawan/panen/pull/121))
- IDX website as secondary data provider alongside Yahoo Finance
- Provider health monitoring in Settings > Data Providers
- Export and import for full data portability -- backup, restore, and device migration ([#46](https://github.com/lugassawan/panen/issues/46), [#123](https://github.com/lugassawan/panen/pull/123))
- GitHub Actions CI pipeline with linting and testing workflows ([#49](https://github.com/lugassawan/panen/issues/49), [#119](https://github.com/lugassawan/panen/pull/119))
- CONTRIBUTING guide with development workflow, code style, testing, i18n, broker fees, and custom provider instructions ([#48](https://github.com/lugassawan/panen/issues/48), [#122](https://github.com/lugassawan/panen/pull/122))
- Calculate totalDeployed from buy transactions in cash flow summary ([#115](https://github.com/lugassawan/panen/issues/115), [#120](https://github.com/lugassawan/panen/pull/120))
- Comprehensive README, CHANGELOG, and user guides ([#47](https://github.com/lugassawan/panen/issues/47))

### Changed

- Refactored codebase for consistency and maintainability ([#117](https://github.com/lugassawan/panen/pull/117), [#118](https://github.com/lugassawan/panen/pull/118))
- Extracted shared scan helpers for database repositories ([#114](https://github.com/lugassawan/panen/pull/114))
- Replaced inline empty states with shared EmptyState component ([#113](https://github.com/lugassawan/panen/pull/113))

## [0.6.0] - 2026-02-22

Robustness and polish release with backup/recovery, logging, live config, and broker fee picker.

### Added

- Automated daily backup with startup recovery flow ([#100](https://github.com/lugassawan/panen/pull/100))
- Persistent structured logging with rotation and debug mode ([#101](https://github.com/lugassawan/panen/pull/101))
- Unified live config system with three-layer fallback (remote, cache, bundled) ([#102](https://github.com/lugassawan/panen/pull/102))
- Searchable broker picker with automatic fee sync from live config ([#103](https://github.com/lugassawan/panen/pull/103))
- In-app self-update system with GitHub Releases integration ([#104](https://github.com/lugassawan/panen/pull/104))
- Skeleton loader components for loading states ([#112](https://github.com/lugassawan/panen/pull/112))
- Shared LoadingState component ([#109](https://github.com/lugassawan/panen/pull/109))
- Shared EmptyState component ([#110](https://github.com/lugassawan/panen/pull/110))
- Shared Modal base component ([#108](https://github.com/lugassawan/panen/pull/108))

### Changed

- Extracted shared applog package ([#98](https://github.com/lugassawan/panen/pull/98))
- Updated documentation with i18n, new components, and conventions ([#99](https://github.com/lugassawan/panen/pull/99), [#116](https://github.com/lugassawan/panen/pull/116))

## [0.5.0] - 2026-02-08

Feature-complete release with portfolio dashboard, charts, comparison, transaction history, and dividend calendar.

### Added

- Portfolio dashboard as landing page ([#107](https://github.com/lugassawan/panen/pull/107))
- Stock comparison view ([#106](https://github.com/lugassawan/panen/pull/106))
- Transaction history view ([#105](https://github.com/lugassawan/panen/pull/105))

## [0.4.0] - 2026-01-25

Charts, dividend calendar, internationalization, and theme support.

### Added

- Price history chart ([#84](https://github.com/lugassawan/panen/pull/84))
- Charts and visualizations for portfolio detail ([#81](https://github.com/lugassawan/panen/pull/81))
- Fundamental change alerts ([#88](https://github.com/lugassawan/panen/pull/88))
- Internationalization with English and Bahasa Indonesia ([#87](https://github.com/lugassawan/panen/pull/87))
- Dividend calendar ([#86](https://github.com/lugassawan/panen/pull/86))
- Valuation zone indicators ([#85](https://github.com/lugassawan/panen/pull/85))

### Changed

- UX polish, accessibility, and user guidance improvements ([#79](https://github.com/lugassawan/panen/pull/79), [#80](https://github.com/lugassawan/panen/pull/80))
- Refactored codebase and improved test coverage ([#74](https://github.com/lugassawan/panen/pull/74))

## [0.3.0] - 2026-01-11

Monthly payday assistant, crash playbook, screener, DCA logic, and trailing stop.

### Added

- Monthly payday assistant ([#66](https://github.com/lugassawan/panen/pull/66))
- Crash playbook for proactive crash preparedness ([#67](https://github.com/lugassawan/panen/pull/67))
- Stock screener with filtering and ranking ([#71](https://github.com/lugassawan/panen/pull/71))
- DCA and average-up logic for Dividend Mode ([#72](https://github.com/lugassawan/panen/pull/72))
- Trailing stop suggestion engine for Value Mode ([#69](https://github.com/lugassawan/panen/pull/69))
- Auto update check with GitHub Releases ([#65](https://github.com/lugassawan/panen/pull/65))

### Fixed

- Pre-commit hook linter failure handling ([#73](https://github.com/lugassawan/panen/pull/73))
- Select component styling and warning suppression ([#70](https://github.com/lugassawan/panen/pull/70))

## [0.2.0] - 2025-12-28

Brokerage accounts, portfolio management, watchlist, background refresh, and conviction checklists.

### Added

- Brokerage account management ([#52](https://github.com/lugassawan/panen/pull/52))
- Portfolio management ([#54](https://github.com/lugassawan/panen/pull/54))
- Watchlist ([#55](https://github.com/lugassawan/panen/pull/55))
- Background auto-refresh for stock data ([#56](https://github.com/lugassawan/panen/pull/56))
- Conviction checklist system ([#60](https://github.com/lugassawan/panen/pull/60))
- Design system with tokens, components, and theme support ([#50](https://github.com/lugassawan/panen/pull/50))
- Release workflow and install script ([#53](https://github.com/lugassawan/panen/pull/53))
- Cross-platform build and distribution ([#22](https://github.com/lugassawan/panen/pull/22))
- nolateconst analyzer for panenlint ([#57](https://github.com/lugassawan/panen/pull/57))

### Changed

- Self-hosted fonts and migrated to lucide-svelte icons ([#51](https://github.com/lugassawan/panen/pull/51))
- Refactored backend structure and conventions ([#62](https://github.com/lugassawan/panen/pull/62))

## [0.1.0] - 2025-12-14

Initial release with single stock lookup, Yahoo Finance scraper, and valuation engine.

### Added

- Project scaffold with Wails v2, Svelte 5, and Tailwind CSS ([#1](https://github.com/lugassawan/panen/pull/1))
- Brand assets, app icon, and favicon ([#2](https://github.com/lugassawan/panen/pull/2))
- Database foundation with SQLite ([#11](https://github.com/lugassawan/panen/pull/11))
- Data provider interface and Yahoo Finance scraper ([#12](https://github.com/lugassawan/panen/pull/12))
- Valuation engine with Graham Number and PBV/PER bands ([#13](https://github.com/lugassawan/panen/pull/13))
- Core backend with use cases, presenter, and platform detection ([#14](https://github.com/lugassawan/panen/pull/14))
- Single stock lookup UI ([#16](https://github.com/lugassawan/panen/pull/16))
- Basic portfolio and holdings UI ([#20](https://github.com/lugassawan/panen/pull/20))
- App shell with sidebar navigation and page routing ([#19](https://github.com/lugassawan/panen/pull/19))

### Fixed

- Cookie and crumb authentication for Yahoo Finance scraper ([#17](https://github.com/lugassawan/panen/pull/17))

[1.0.0]: https://github.com/lugassawan/panen/compare/v0.6.0...v1.0.0
[0.6.0]: https://github.com/lugassawan/panen/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/lugassawan/panen/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/lugassawan/panen/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/lugassawan/panen/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/lugassawan/panen/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/lugassawan/panen/releases/tag/v0.1.0
