import type { Handle } from '@sveltejs/kit';

/**
 * Attach defense-in-depth security headers to every response.
 *
 * Content-Security-Policy is NOT set here — it lives in svelte.config.js
 * (`kit.csp`) so SvelteKit can auto-hash its own bootstrap inline scripts.
 * Setting it here as well would collide with SvelteKit's injection and
 * silently break hydration (we hit that bug; don't reintroduce it).
 */
export const handle: Handle = async ({ event, resolve }) => {
	const response = await resolve(event);

	// Block sniffing, clickjacking, referrer leakage, and feature abuse.
	response.headers.set('X-Content-Type-Options', 'nosniff');
	response.headers.set('X-Frame-Options', 'DENY');
	response.headers.set('Referrer-Policy', 'strict-origin-when-cross-origin');
	response.headers.set(
		'Permissions-Policy',
		'camera=(), microphone=(), geolocation=(), payment=(), usb=()'
	);

	// HSTS is harmless on http://localhost (browsers ignore it) and
	// effective the moment we're behind a TLS-terminating proxy.
	response.headers.set('Strict-Transport-Security', 'max-age=63072000; includeSubDomains; preload');

	// Strip framework-fingerprintable headers SvelteKit sets by default.
	response.headers.delete('x-sveltekit-page');
	response.headers.delete('x-powered-by');
	// Remove the build-stable ETag (lets an adversary detect deploys).
	response.headers.delete('etag');

	return response;
};
