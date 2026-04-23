# Changelog

## [1.2.3](https://github.com/andreistefanciprian/urlshortener/compare/url-read-v1.2.2...url-read-v1.2.3) (2026-04-23)


### Bug Fixes

* **deps:** bump Go to 1.26.2 and grpc to v1.79.3 across all services ([9059cc5](https://github.com/andreistefanciprian/urlshortener/commit/9059cc5bbd8a7d4d91b679faadeb33ec2a164286))

## [1.2.2](https://github.com/andreistefanciprian/urlshortener/compare/url-read-v1.2.1...url-read-v1.2.2) (2026-03-14)


### Bug Fixes

* **url-read:** add package doc comment to main ([f49d158](https://github.com/andreistefanciprian/urlshortener/commit/f49d1586cfa48bfac2a4613f9bd3c354f6fe64b3))
* **url-read:** add package doc comment to main ([#182](https://github.com/andreistefanciprian/urlshortener/issues/182)) ([5203929](https://github.com/andreistefanciprian/urlshortener/commit/52039290154cebad216811108752a529d914522c))

## [1.2.1](https://github.com/andreistefanciprian/urlshortener/compare/url-read-v1.2.0...url-read-v1.2.1) (2026-03-11)


### Bug Fixes

* grpc health probe improvements for url-gen and url-read ([4de7f55](https://github.com/andreistefanciprian/urlshortener/commit/4de7f5570ba7c017877c4b55a7477f49f314d762))
* replace log.Fatal/log.Printf with logrus; use Error on Serve failure ([e76b2c8](https://github.com/andreistefanciprian/urlshortener/commit/e76b2c84031f75ca97c3c1f493a3e6e38cd8bd63))
* **url-gen:** grpc health probe improvements for url-gen and url-read ([#166](https://github.com/andreistefanciprian/urlshortener/issues/166)) ([b02a0b7](https://github.com/andreistefanciprian/urlshortener/commit/b02a0b7e26fed9fde2a3c139623f0c1be27407ea))
* use custom no-op logger for redis.SetLogger (v9 interface fix) ([ebd9516](https://github.com/andreistefanciprian/urlshortener/commit/ebd9516fcf6062c3ebcb0ae8fbfec6e512ad1428))

## [1.2.0](https://github.com/andreistefanciprian/urlshortener/compare/url-read-v1.1.0...url-read-v1.2.0) (2026-03-10)


### Features

* **url-read:** improve grpc health probes for kubernetes readiness ([0a51710](https://github.com/andreistefanciprian/urlshortener/commit/0a51710c141a5c2efaa1e3fe397a4d10d2ce451b))
* **url-read:** improve grpc health probes for kubernetes readiness ([#160](https://github.com/andreistefanciprian/urlshortener/issues/160)) ([10caf53](https://github.com/andreistefanciprian/urlshortener/commit/10caf53770e302e5a74a6b3cd1d08e31d64e2158))


### Bug Fixes

* **url-gen:** improve grpc health probes for kubernetes readiness ([2864c9a](https://github.com/andreistefanciprian/urlshortener/commit/2864c9a34e268983dfb15ce2780437d875f7d405))

## [1.1.0](https://github.com/andreistefanciprian/urlshortener/compare/url-read-v1.0.0...url-read-v1.1.0) (2026-03-08)


### Features

* **ci:** update url-read, url-gen, frontend workflows and add release-please config ([74c725a](https://github.com/andreistefanciprian/urlshortener/commit/74c725aa77c58b6cdec4c1a7d13ed8c319d47f1d))
