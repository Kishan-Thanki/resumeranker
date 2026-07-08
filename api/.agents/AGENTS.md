## 1. API State Locked
As of the MVP completion, the backend API (`api/`) is considered feature-complete. Do not introduce any new features or endpoints. Modifications should be limited strictly to necessary bug fixes or structural updates that do not expand the product scope.

## 2. Coding Standards
- **Architecture**: Strict separation of concerns (Handlers -> Services -> Repositories). Interfaces are defined locally where they are used ("accept interfaces, return structs") to prevent tight coupling.
- **Database**: All SQL queries MUST be generated using `sqlc`. Do not write raw queries in the repository structs.
- **Configuration**: All configuration values (URLs, Limits, Durations, Secrets) MUST be read from environment variables via `api/internal/config/config.go`. No hardcoded strings.
- **Dependencies**: Use explicit Dependency Injection across all structs.
- **Errors**: Return meaningful domain errors from Services, which Handlers translate into appropriate HTTP status codes (400, 401, 403, 404, 500).
