# ResumeRanker API Context

## Overview
The `api/` directory contains the backend for ResumeRanker, a monolithic Go service providing RESTful APIs for the web platform.

## Architecture
The API follows a strict layered (Clean) architecture:
- **Handlers (`http`)**: Responsible for decoding JSON, extracting URL/Query params, invoking Services, and returning JSON responses with correct HTTP status codes.
- **Services**: Contains core business logic. Defines local interfaces for its dependencies (Repository, other services). Follows dependency injection.
- **Repositories (`database`)**: Interfaces with PostgreSQL using `sqlc` generated queries and `pgxpool`.

## Core Domains
- `users`: User registration, authentication, profile management, and terms of service (agreements). Handles naive bulk emails via goroutines.
- `apikey`: Creation, validation, and quota tracking of API keys.
- `analysis`: gRPC Client that forwards PDF bytes to the isolated Python AI Engine. It injects a `RequestID` for distributed tracing and gracefully parses gRPC error strings into semantic Go sentinel errors (`ErrRateLimit`, `ErrInvalidPDF`) to yield exact HTTP codes (429, 400).
- `audit`: Logging of critical system events (e.g., logins, key creation).
- `auth`: JWT session management and cookie issuance.
- `email`: Integration with Resend for transactional emails (Verification, Password Reset, Status updates).

## Key Technologies
- **RPC Communication**: `grpc` (connecting Go to the Python Engine via `proto/analysis.proto`)
- **Router**: `go-chi/chi`
- **Database**: PostgreSQL (driver: `pgx/v5`)
- **Query Builder**: `sqlc` (writes queries in `.sql`, generates type-safe Go)
- **Authentication**: JWT (JSON Web Tokens) stored in `HttpOnly` cookies.
- **Logging**: `slog` with custom redaction (`slogredact`) for sensitive fields.
- **Configuration**: Strictly environment-variable driven via `api/internal/config/config.go` (no hardcoded credentials or limits).
