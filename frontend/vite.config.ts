import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

// Vitest config is in `vitest.config.ts` — kept separate because the
// SvelteKit plugin forces SSR-condition resolution for `svelte`, which
// breaks @testing-library/svelte's `mount()` call. The test config uses
// the plain `@sveltejs/vite-plugin-svelte` with browser conditions.
export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	ssr: {
		noExternal: ['svelte-sonner']
	},
	build: {
		// Explicit OFF — production bundles must not ship .map files, which
		// would reveal the un-minified TypeScript sources of the app.
		// Vite 8 uses oxc-minify by default; we don't pin `minify` here
		// because forcing `esbuild` requires installing it as an explicit
		// dep in Vite 8+ (it's no longer bundled).
		sourcemap: false
	}
});
