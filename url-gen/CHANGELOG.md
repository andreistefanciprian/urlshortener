# Changelog

## [1.1.5](https://github.com/andreistefanciprian/urlshortener/compare/url-gen-v1.1.4...url-gen-v1.1.5) (2026-04-23)


### Bug Fixes

* **deps:** bump Go to 1.26.2 and grpc to v1.79.3 across all services ([9059cc5](https://github.com/andreistefanciprian/urlshortener/commit/9059cc5bbd8a7d4d91b679faadeb33ec2a164286))

## [1.1.4](https://github.com/andreistefanciprian/urlshortener/compare/url-gen-v1.1.3...url-gen-v1.1.4) (2026-03-14)


### Bug Fixes

* **url-gen:** extract health check timeout into named constant ([7547596](https://github.com/andreistefanciprian/urlshortener/commit/75475968987b4ed151e31d3837f080f6d6bc6696))
* **url-gen:** extract health check timeout into named constant ([#188](https://github.com/andreistefanciprian/urlshortener/issues/188)) ([e80aa90](https://github.com/andreistefanciprian/urlshortener/commit/e80aa901a55696fb4d6397ea988993bef967f7c4))

## [1.1.3](https://github.com/andreistefanciprian/urlshortener/compare/url-gen-v1.1.2...url-gen-v1.1.3) (2026-03-14)


### Bug Fixes

* **url-gen:** add package doc comment to main ([01653ef](https://github.com/andreistefanciprian/urlshortener/commit/01653efbe50b8e564536393e109b10a6f45b5c61))
* **url-gen:** add package doc comment to main ([#177](https://github.com/andreistefanciprian/urlshortener/issues/177)) ([4ab5fe3](https://github.com/andreistefanciprian/urlshortener/commit/4ab5fe30bfb1584ed53213afc079caf4b8442c29))

## [1.1.2](https://github.com/andreistefanciprian/urlshortener/compare/url-gen-v1.1.1...url-gen-v1.1.2) (2026-03-11)


### Bug Fixes

* grpc health probe improvements for url-gen and url-read ([4de7f55](https://github.com/andreistefanciprian/urlshortener/commit/4de7f5570ba7c017877c4b55a7477f49f314d762))
* replace log.Fatal/log.Printf with logrus; use Error on Serve failure ([e76b2c8](https://github.com/andreistefanciprian/urlshortener/commit/e76b2c84031f75ca97c3c1f493a3e6e38cd8bd63))
* **url-gen:** grpc health probe improvements for url-gen and url-read ([#166](https://github.com/andreistefanciprian/urlshortener/issues/166)) ([b02a0b7](https://github.com/andreistefanciprian/urlshortener/commit/b02a0b7e26fed9fde2a3c139623f0c1be27407ea))
* use custom no-op logger for redis.SetLogger (v9 interface fix) ([ebd9516](https://github.com/andreistefanciprian/urlshortener/commit/ebd9516fcf6062c3ebcb0ae8fbfec6e512ad1428))

## [1.1.1](https://github.com/andreistefanciprian/urlshortener/compare/url-gen-v1.1.0...url-gen-v1.1.1) (2026-03-10)


### Bug Fixes

* **url-gen:** improve grpc health probes for kubernetes readiness ([2864c9a](https://github.com/andreistefanciprian/urlshortener/commit/2864c9a34e268983dfb15ce2780437d875f7d405))

## [1.1.0](https://github.com/andreistefanciprian/urlshortener/compare/url-gen-v1.0.0...url-gen-v1.1.0) (2026-03-08)


### Features

* **ci:** update url-read, url-gen, frontend workflows and add release-please config ([74c725a](https://github.com/andreistefanciprian/urlshortener/commit/74c725aa77c58b6cdec4c1a7d13ed8c319d47f1d))
