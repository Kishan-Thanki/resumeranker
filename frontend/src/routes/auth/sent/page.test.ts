import { describe, expect, it } from 'vitest';
import { render, screen } from '@testing-library/svelte';

import { setPage, resetPage } from '../../../vitest-setup';

const Page = (await import('./+page.svelte')).default;

describe('/auth/sent', () => {
	it('shows the email back to the user when present in the URL', () => {
		setPage({
			url: new URL('http://localhost:5173/auth/sent?email=jane%40example.com')
		});
		render(Page);
		expect(screen.getByText(/jane@example\.com/)).toBeInTheDocument();
		resetPage();
	});

	it('falls back gracefully when email is missing', () => {
		setPage({ url: new URL('http://localhost:5173/auth/sent') });
		render(Page);
		expect(screen.getByText(/check your email/i)).toBeInTheDocument();
		resetPage();
	});
});
