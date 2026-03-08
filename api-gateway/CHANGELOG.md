# Changelog

## [1.0.6](https://github.com/andreistefanciprian/urlshortener/compare/api-gateway-v1.0.5...api-gateway-v1.0.6) (2026-03-08)


### Bug Fixes

* **api-gateway:** update package comment ([000825f](https://github.com/andreistefanciprian/urlshortener/commit/000825f0b37dc2e03fae22a504b01a4d010688e1))

## [1.0.5](https://github.com/andreistefanciprian/urlshortener/compare/api-gateway-v1.0.4...api-gateway-v1.0.5) (2026-03-08)


### Bug Fixes

* **api-gateway:** hyphenate health-check in package comment ([2e9f77b](https://github.com/andreistefanciprian/urlshortener/commit/2e9f77b279821d1ac6c76b3df7cba961bb6b2bc8))

## [1.0.4](https://github.com/andreistefanciprian/urlshortener/compare/api-gateway-v1.0.3...api-gateway-v1.0.4) (2026-03-08)


### Bug Fixes

* **api-gateway:** update package comment ([7cdbcf1](https://github.com/andreistefanciprian/urlshortener/commit/7cdbcf1266c4340a10d1564e3219c03f11c8e073))

## [1.0.3](https://github.com/andreistefanciprian/urlshortener/compare/api-gateway-v1.0.2...api-gateway-v1.0.3) (2026-03-08)


### Bug Fixes

* **api-gateway:** update package comment in handlers ([2821b59](https://github.com/andreistefanciprian/urlshortener/commit/2821b5946f8d7a0065960890fe4a94cf7ca40384))

## [1.0.2](https://github.com/andreistefanciprian/urlshortener/compare/api-gateway-v1.0.1...api-gateway-v1.0.2) (2026-03-08)


### Bug Fixes

* merge changelogs and fix version-file path in release-please config ([ed92b20](https://github.com/andreistefanciprian/urlshortener/commit/ed92b2058c11b01387842f199319d280181c1a15))

## [1.0.1](https://github.com/andreistefanciprian/urlshortener/compare/api-gateway-v1.0.0...api-gateway-v1.0.1) (2026-03-08)


### Bug Fixes

* **api-gateway:** add package comment to handlers ([1362278](https://github.com/andreistefanciprian/urlshortener/commit/13622785a4971ac08d6adefa4b7a250ff82c730e))

## 1.0.0 (2026-03-08)


### Features

* add URL_SCHEME env var to api-gateway for http/https scheme selection ([00d4dd0](https://github.com/andreistefanciprian/urlshortener/commit/00d4dd023c98dbb223063a917af370ce05e31f72))


### Bug Fixes

* normalize and validate URL_SCHEME on startup ([3f10720](https://github.com/andreistefanciprian/urlshortener/commit/3f10720b4c54a16b440a1aaad4e35742883b385e))
* preserve original URL_SCHEME value in fatal log for easier misconfiguration diagnosis ([e3826bf](https://github.com/andreistefanciprian/urlshortener/commit/e3826bf45c5cc3c88d7ce69c2b92b71970079d75))
* use logrus instead of stdlib log in mustURLScheme for consistent structured logging ([3bc64b9](https://github.com/andreistefanciprian/urlshortener/commit/3bc64b93f79da47f946dcf6deed023d2e1f2ad93))
