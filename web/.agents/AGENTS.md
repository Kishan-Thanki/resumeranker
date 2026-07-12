# Web Frontend Agents Rules

These rules apply strictly to any AI agent modifying the `@web` directory.

## Core Directives
1. **No Frameworks, No Build Steps**: The frontend must remain pure Vanilla HTML, CSS, and ES6 Javascript Modules. Do not introduce React, Vue, Webpack, Vite, or Tailwind.
2. **Absolute Security**: 
   - Never use `innerHTML` or `insertAdjacentHTML` under any circumstances. Always use `document.createElement()` and `textContent` to construct the DOM securely.
   - Cross-Site Request Forgery (CSRF) tokens must never be stored in `localStorage` or `sessionStorage`. They must be held in memory (closure variables) via `api.js`.
   - Authentication relies purely on `HttpOnly` cookies. Never attempt to read or store JWTs in JavaScript.
3. **Rigid Design Aesthetics**:
   - The UI is strictly black and white. 
   - No colors are allowed. Only monochromatic shades (white, blacks, grays) driven by native CSS variables.
   - Theme must support only light and dark mode toggles.
4. **Decoupled Architecture**: 
   - The web container uses Caddy for routing and static file serving.
   - The web container is completely stateless. It communicates with the backend exclusively via relative paths (e.g., `/api/v1/...`) routed through Caddy's reverse proxy.
