# Copilot Instructions

This repository implements a microservices architecture in Go, with the following services:

- **api-gateway**: REST API and gRPC client in ./api-gateway
- **url-gen**: gRPC server (handles short url generation logic) in ./url-gen
- **url-read**: gRPC server (handles short url reader logic) in ./url-read
- **redis**: Cache layer for fast URL lookups and reduced database load
- **postgresql**: Primary database for persistent URL storage and metadata

## General Guidelines
- Do not write complete functions or snippets unless I specifically ask for them.
- When writing code for me, help me understand it rather than just giving me the answer.
- My purpose is to get better at coding not just to get the code and follow your instructions blindly.
- Follow idiomatic Go practices and conventions.
- All services use context propagation.

## Architecture Decisions

- LongURL expiration time validation is performed upstream (at the API Gateway level) rather than at the database or cache layer.
