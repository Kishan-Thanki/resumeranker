# ResumeRanker - Frontend Layer

This directory contains the entire frontend architecture for ResumeRanker.

## Architecture Overview

The frontend is built as a pure, lightweight static site served directly by a reverse proxy or web server on your server infrastructure.

- **Vanilla Frontend:** Pure ES6 JavaScript, HTML, and CSS (no heavy build steps, no framework bloat like React or Vite).

### Directory Structure

```text
web/
└── public/              # (Source Code) Pure Vanilla HTML/CSS/ES6 JS (No build steps)
```

## Architecture & Frontend Philosophy

1. **Zero-Build Stack**: The frontend utilizes pure ES6 Modules (`import`/`export`), native CSS Variables, and standard HTML5. There is no Webpack, Vite, or Tailwind.

2. **Strict Monochromatic Design**: The UI strictly enforces a black-and-white aesthetic with native Light/Dark mode toggling. No colors are permitted.

3. **DOM Safety**: To prevent XSS vulnerabilities, the codebase strictly avoids `innerHTML` and dynamically constructs all elements using `document.createElement`.

4. **Secure State Management**: CSRF tokens are securely maintained in JavaScript memory closures (inside `api.js`) to prevent XSS theft, and JWTs are handled entirely via backend `HttpOnly` cookies.

5. **Semantic Error Mapping**: The frontend (`api.js` and `ui.js`) is natively aware of the API's complex HTTP response states (e.g., catching `429 Too Many Requests` or `502 Bad Gateway` from the AI Engine) and elegantly displaying them to the user via toast notifications.

## License

This project is proprietary and confidential. All rights are reserved.
Please see the [LICENSE](LICENSE) file for more details.
