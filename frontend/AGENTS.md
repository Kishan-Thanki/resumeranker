# Rules for AI assistant on this project

Follow these rules when editing this repo. They come from the v1 frontend
brief; if `BRIEF.md` is present at the repo root it is the source of truth,
otherwise the brief lives in the conversation that created this scaffold.

1. **No `any` in TypeScript.** Anywhere. Use `unknown` and narrow, or define proper types.
2. **Strict mode in `tsconfig.json`.** No exceptions.
3. **No file deletion.** If a file becomes obsolete, leave it and ask. Never delete without explicit permission.
4. **Components live in component files.** A route `+page.svelte` is for composition and data flow, not markup beyond ~100 lines.
5. **Mock data only in `src/lib/mock/`.** No hardcoded mock arrays inline in components.
6. **shadcn-svelte is the only external component library.** Do not introduce Skeleton, Bits UI (directly), Flowbite, or any other component library alongside it. If a component you need isn't in shadcn-svelte, build it yourself in the shadcn style.
7. **Accessibility:** every interactive element has a label or `aria-label`. Focus states must be visible. Forms must be keyboard-navigable. No `div` with `onclick` — use real buttons.
8. **Mobile works.** Every page must render usably at 380px viewport. Mobile is not optional.
9. **`lucide-svelte` (or `@lucide/svelte`) is the only icon set.** It is shadcn-svelte's default. Do not introduce other icon libraries.
10. **Understand before using.** If a library or pattern is unfamiliar, explain it in a code comment or ask before adding it.
11. **No `console.log` left in committed code.** Remove or convert to comments.
12. **Prefer composition over abstraction.** Don't extract a component until it has two real callers.
13. **Docker-first workflow.** Never run `pnpm dev` directly on the host. All commands go through `docker compose exec web ...`. The host has no `node_modules` directory.
14. **Pin Node and pnpm versions.** Use `node:22-alpine` in Dockerfiles. The `"packageManager"` field in `package.json` pins pnpm so the lockfile and the image stay in sync.
