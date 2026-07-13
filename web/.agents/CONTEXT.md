# Web Frontend Context

## Architecture
The `@web` module is a zero-build, ultra-lightweight frontend stack served by an embedded Caddy server. 

## Key Technologies
- **Caddy**: Serves static HTML/CSS/JS and acts as a reverse proxy, seamlessly forwarding requests from `/api/*` to the internal Go `api-server:8080`.
- **Vanilla ES6**: No React, no Vue, no Webpack. Native modules only.
- **Native CSS Variables**: All theming is handled by root CSS variables to maintain a strict monochromatic design (Light/Dark mode).

## Javascript Modules (`public/assets/js/`)
1. `api.js`: The central nervous system for API calls. Holds the CSRF token in an ephemeral JS closure (never in localStorage) and intercepts generic HTTP fetch errors to translate `429 Too Many Requests`, `502 Bad Gateway`, and `504 Gateway Timeout` into user-friendly strings.
2. `auth.js`: Handles login, registration, password resets, and verification states. 
3. `ui.js`: Global DOM manipulations, toasts, modals, and dark mode toggling. Avoids `innerHTML` completely to prevent XSS.
4. `dashboard.js` & `account.js`: Page-specific controllers for the authenticated app layer.

## Project Structure
- `/public/`: Public landing, about, contact pages.
- `/public/auth/`: The unauthenticated flow (`login.html`, `register.html`).
- `/public/dashboard/`: The authenticated application flow (`dashboard.html`, `account.html`).

## Domain Specifics
The production domain is strictly **`resumeranker.kishanthanki.dev`**. All embedded "support" links or legal emails use this domain (e.g., `noreply@resumeranker.kishanthanki.dev`).
