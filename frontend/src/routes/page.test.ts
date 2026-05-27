import { describe, expect, it } from 'vitest';
import { render, screen } from '@testing-library/svelte';

import Page from './+page.svelte';
import Footer from '$lib/components/Footer.svelte';

describe('/ landing', () => {
	it('renders the hero copy and CTA to /auth', () => {
		render(Page);

		expect(
			screen.getByText(/See exactly how your resume matches a job description/i)
		).toBeInTheDocument();

		const cta = screen.getByRole('link', { name: /get started/i });
		expect(cta).toHaveAttribute('href', '/auth');
	});

	it('renders the three explainer block headings', () => {
		render(Page);
		expect(screen.getByRole('heading', { name: /upload/i })).toBeInTheDocument();
		expect(screen.getByRole('heading', { name: /analyze/i })).toBeInTheDocument();
		expect(screen.getByRole('heading', { name: /see evidence/i })).toBeInTheDocument();
	});
});

// Footer was extracted to a shared component and is now rendered by the
// root layout, not the landing page. Tested in isolation so we keep the
// "every page has the same policy links" guarantee.
describe('site Footer', () => {
	it('links to all five policy/info pages', () => {
		render(Footer);
		const aboutLink = screen.getByRole('link', { name: /^about$/i });
		const contactLink = screen.getByRole('link', { name: /^contact$/i });
		const termsLink = screen.getByRole('link', { name: /^terms$/i });
		const privacyLink = screen.getByRole('link', { name: /^privacy$/i });
		const a11yLink = screen.getByRole('link', { name: /^accessibility$/i });
		expect(aboutLink).toHaveAttribute('href', '/about');
		expect(contactLink).toHaveAttribute('href', '/contact');
		expect(termsLink).toHaveAttribute('href', '/terms');
		expect(privacyLink).toHaveAttribute('href', '/privacy');
		expect(a11yLink).toHaveAttribute('href', '/accessibility');
	});
});
