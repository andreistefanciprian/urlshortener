# Changelog

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
