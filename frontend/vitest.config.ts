import { svelte } from '@sveltejs/vite-plugin-svelte';
import { fileURLToPath } from 'node:url';
import { defineConfig } from 'vitest/config';

// Standalone Vitest config. Uses the plain Svelte plugin instead of the
// SvelteKit plugin so that `svelte` resolves to its browser/client exports
// (testing-library calls `mount()` which only exists client-side).
//
// SvelteKit's `$lib` and `$app/*` aliases are not provided here because we
// load the plain Svelte plugin instead of `sveltekit()`. We add `$lib`
// manually; `$app/*` modules are stubbed in `src/vitest-setup.ts`.
export default defineConfig({
	plugins: [svelte()],
	resolve: {
		conditions: ['browser'],
		alias: {
			$lib: fileURLToPath(new URL('./src/lib', import.meta.url)),
			'$app/environment': fileURLToPath(
				new URL('./src/test-stubs/app-environment.ts', import.meta.url)
			),
			'$app/navigation': fileURLToPath(
				new URL('./src/test-stubs/app-navigation.ts', import.meta.url)
			),
			'$app/state': fileURLToPath(new URL('./src/test-stubs/app-state.ts', import.meta.url))
		}
	},
	test: {
		environment: 'jsdom',
		globals: true,
		include: ['src/**/*.{test,spec}.ts'],
		setupFiles: ['src/vitest-setup.ts']
	}
});
