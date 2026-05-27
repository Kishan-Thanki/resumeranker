import type { Config } from 'tailwindcss';

// Tailwind v4: the bulk of configuration lives in src/app.css via the
// `@theme inline { ... }` block (colors, radii, font stacks) and the
// `@tailwindcss/vite` plugin is the loader. This file exists so the
// repo file structure matches BRIEF.md and so any tool that looks for
// a config file finds one. Content paths are auto-detected by v4 from
// the Vite plugin; we list them here for tooling that still reads them.
export default {
	content: ['./src/**/*.{html,js,svelte,ts}']
} satisfies Config;
