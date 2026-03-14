# Changelog

## [1.1.0](https://github.com/andreistefanciprian/urlshortener/compare/api-gateway-v1.0.9...api-gateway-v1.1.0) (2026-03-14)


### Features

* **api-gateway:** add HTTPRoute template support ([42cf852](https://github.com/andreistefanciprian/urlshortener/commit/42cf852c34df42a8a411ca1069d32f71fdf3c68a))
* **api-gateway:** add HTTPRoute template support ([#190](https://github.com/andreistefanciprian/urlshortener/issues/190)) ([3b5178c](https://github.com/andreistefanciprian/urlshortener/commit/3b5178c9b522c4ce8aa388e42489e19c200e0bbb))

## [1.0.9](https://github.com/andreistefanciprian/urlshortener/compare/api-gateway-v1.0.8...api-gateway-v1.0.9) (2026-03-14)


### Bug Fixes

* **url-gen:** improve grpc health probes for kubernetes readiness ([2864c9a](https://github.com/andreistefanciprian/urlshortener/commit/2864c9a34e268983dfb15ce2780437d875f7d405))

## [1.0.8](https://github.com/andreistefanciprian/urlshortener/compare/api-gateway-v1.0.7...api-gateway-v1.0.8) (2026-03-08)


### Bug Fixes

* **api-gateway:** fix gofmt issue in fake code change ([9fb6a3c](https://github.com/andreistefanciprian/urlshortener/commit/9fb6a3c563c7ef72565c5a82064c6ff0766f712f))
* **api-gateway:** update GHCR workflow inputs and fake code change ([c7dcda5](https://github.com/andreistefanciprian/urlshortener/commit/c7dcda5978a8bd95b63106d255525590d4e5554d))

## [1.0.7](https://github.com/andreistefanciprian/urlshortener/compare/api-gateway-v1.0.6...api-gateway-v1.0.7) (2026-03-08)


### Bug Fixes

* **api-gateway:** simplify package comment ([84e6f63](https://github.com/andreistefanciprian/urlshortener/commit/84e6f63dd2efffc57c5893ee808045828922007b))

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
