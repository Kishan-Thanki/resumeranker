import adapter from '@sveltejs/adapter-node';

// CSP `connect-src` must include the backend origin so the browser is
// allowed to fetch the API. Inlined here at build time (matches Vite's
// build-time inlining of VITE_API_BASE_URL into the client bundle).
const apiOrigin = process.env.VITE_API_BASE_URL ?? 'http://localhost:8000';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	compilerOptions: {
		runes: ({ filename }) => (filename.split(/[/\\]/).includes('node_modules') ? undefined : true)
	},
	kit: {
		adapter: adapter(),

		// SvelteKit emits a couple of inline `<script>` tags per page for
		// hydration bootstrap. A strict `script-src 'self'` blocks them
		// and silently kills hydration — typing/clicking does nothing.
		// Letting SvelteKit own the CSP means it auto-injects hashes for
		// exactly those inline scripts; everything else still requires
		// `'self'`.
		//
		// `mode: 'hash'` (not 'auto') is required because mode-watcher
		// injects an early-paint dark-mode script via `<svelte:head>`
		// using `{@html ...}`. SvelteKit's nonce-mode only nonces its
		// own scripts and misses this one; hash-mode walks every inline
		// `<script>` in the final HTML and adds an sha256 allow-entry.
		csp: {
			mode: 'hash',
			directives: {
				'default-src': ['self'],
				// `mode: 'hash'` only hashes SvelteKit's own emitted inline
				// scripts. mode-watcher injects an early-paint dark-mode
				// script via `<svelte:head>` + `{@html ...}` which sits
				// outside that pipeline. Its hash is hard-coded below; if
				// mode-watcher releases a new version that changes the
				// script body, the browser console will report the new
				// expected hash — replace this value with that.
				'script-src': [
					'self',
					'sha256-Cr3r+iKjDTUxJaxM3r/Iq0ow6clOB9AqoT6j0wMFMIM=' // mode-watcher setInitialMode IIFE
				],
				// Tailwind 4 + Vite emit inline styles; toggling this off
				// breaks the UI. Documented gap in frontend/SECURITY.md.
				'style-src': ['self', 'unsafe-inline'],
				'img-src': ['self', 'data:'],
				'font-src': ['self', 'data:'],
				'connect-src': ['self', apiOrigin],
				'frame-ancestors': ['none'],
				'form-action': ['self'],
				'base-uri': ['self'],
				'object-src': ['none']
			}
		}
	}
};

export default config;
