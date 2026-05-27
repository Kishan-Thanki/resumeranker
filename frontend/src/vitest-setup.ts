import '@testing-library/jest-dom/vitest';

import { page } from './test-stubs/app-state';

// Helpers for tests that need to mutate the global `page` state.
export function setPage(partial: Partial<typeof page>): void {
	Object.assign(page, partial);
}

export function resetPage(): void {
	page.url = new URL('http://localhost:5173/');
	page.params = {};
	page.status = 200;
	page.error = null;
}
