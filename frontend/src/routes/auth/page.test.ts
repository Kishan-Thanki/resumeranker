import { describe, expect, it, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';

import { api } from '$lib/api';

vi.mock('$lib/api', async (importOriginal) => {
	const actual = await importOriginal<typeof import('$lib/api')>();
	return {
		...actual,
		api: {
			requestMagicLink: vi.fn(),
			verifyMagicLink: vi.fn(),
			me: vi.fn(),
			signOut: vi.fn(),
			listAnalyses: vi.fn(),
			getAnalysis: vi.fn(),
			createAnalysis: vi.fn()
		}
	};
});

const Page = (await import('./+page.svelte')).default;

describe('/auth', () => {
	it('renders the email input and submit button', () => {
		render(Page);
		expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
		expect(screen.getByRole('button', { name: /send magic link/i })).toBeInTheDocument();
	});

	it('disables submit until an email is entered', () => {
		render(Page);
		const button = screen.getByRole('button', { name: /send magic link/i });
		expect(button).toBeDisabled();
	});

	it('calls api.requestMagicLink with the entered email + policy version on submit', async () => {
		vi.mocked(api.requestMagicLink).mockResolvedValueOnce(undefined);
		render(Page);

		const input = screen.getByLabelText(/email/i) as HTMLInputElement;
		await fireEvent.input(input, { target: { value: 'user@example.com' } });

		// The submit button is gated by the click-wrap acceptance checkbox.
		// Without ticking it, the button stays disabled and submit is a no-op.
		const checkbox = screen.getByLabelText(/i agree to the/i);
		await fireEvent.click(checkbox);

		const button = screen.getByRole('button', { name: /send magic link/i });
		await fireEvent.click(button);

		// Whatever string the auth page hard-codes for CURRENT_POLICY_VERSION
		// must be the second arg. We assert on the shape, not the value, so
		// bumping the version on /auth doesn't break the test.
		expect(api.requestMagicLink).toHaveBeenCalledWith(
			'user@example.com',
			expect.any(String)
		);
	});
});
