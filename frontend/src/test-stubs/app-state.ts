// Stub for `$app/state` — exposes a mutable `page` object that tests can
// rewrite via `setPage()` in `src/vitest-setup.ts`.
export const page: {
	url: URL;
	params: Record<string, string>;
	status: number;
	error: Error | null;
} = {
	url: new URL('http://localhost:5173/'),
	params: {},
	status: 200,
	error: null
};
