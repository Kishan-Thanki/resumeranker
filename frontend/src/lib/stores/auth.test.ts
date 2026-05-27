import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { get } from 'svelte/store';

import { api, ApiError } from '$lib/api';

// Mock the api module so the store doesn't try to call a real backend.
vi.mock('$lib/api', async (importOriginal) => {
	const actual = await importOriginal<typeof import('$lib/api')>();
	return {
		...actual,
		api: {
			requestMagicLink: vi.fn(),
			verifyMagicLink: vi.fn(),
			me: vi.fn(),
			signOut: vi.fn(),
			acceptPolicy: vi.fn(),
			deleteAccount: vi.fn(),
			listAnalyses: vi.fn(),
			getAnalysis: vi.fn(),
			createAnalysis: vi.fn()
		}
	};
});

// Import AFTER the mock so the store picks up our mocked `api`.
const { isAuthed, currentUser, signInWithToken, signOut, refreshAuth, isCurrentlyAuthed } =
	await import('./auth');

describe('auth store', () => {
	beforeEach(() => {
		isAuthed.set(false);
		currentUser.set(null);
		vi.clearAllMocks();
	});

	afterEach(() => {
		vi.restoreAllMocks();
	});

	describe('signInWithToken', () => {
		it('populates currentUser and flips isAuthed', async () => {
			// /auth/verify no longer returns a session token to JS — the
			// backend sets the cookie via Set-Cookie. The response body
			// is just the user record.
			vi.mocked(api.verifyMagicLink).mockResolvedValueOnce({
				user: { id: 'u-1', email: 'a@b.com', acceptedPolicyVersion: '2026-05-19' }
			});

			await signInWithToken('raw-token');

			expect(api.verifyMagicLink).toHaveBeenCalledWith('raw-token');
			expect(get(isAuthed)).toBe(true);
			expect(isCurrentlyAuthed()).toBe(true);
			expect(get(currentUser)?.email).toBe('a@b.com');
		});

		it('propagates ApiError on bad token', async () => {
			vi.mocked(api.verifyMagicLink).mockRejectedValueOnce(new ApiError(400, 'invalid link'));

			await expect(signInWithToken('bad')).rejects.toBeInstanceOf(ApiError);
			expect(get(isAuthed)).toBe(false);
			expect(get(currentUser)).toBeNull();
		});
	});

	describe('signOut', () => {
		it('clears local state and flips isAuthed false', async () => {
			isAuthed.set(true);
			currentUser.set({ id: 'u-1', email: 'a@b.com', acceptedPolicyVersion: null });
			vi.mocked(api.signOut).mockResolvedValueOnce(undefined);

			await signOut();

			expect(api.signOut).toHaveBeenCalled();
			expect(get(isAuthed)).toBe(false);
			expect(get(currentUser)).toBeNull();
		});

		it('swallows backend errors and still clears locally', async () => {
			isAuthed.set(true);
			currentUser.set({ id: 'u-1', email: 'a@b.com', acceptedPolicyVersion: null });
			vi.mocked(api.signOut).mockRejectedValueOnce(new ApiError(500, 'boom'));

			await expect(signOut()).resolves.toBeUndefined();
			expect(get(isAuthed)).toBe(false);
			expect(get(currentUser)).toBeNull();
		});
	});

	describe('refreshAuth', () => {
		it('returns true and flips isAuthed when /me succeeds', async () => {
			vi.mocked(api.me).mockResolvedValueOnce({
				id: 'u-1',
				email: 'a@b.com',
				acceptedPolicyVersion: '2026-05-19'
			});

			const result = await refreshAuth();

			expect(result).toBe(true);
			expect(get(isAuthed)).toBe(true);
			expect(get(currentUser)?.email).toBe('a@b.com');
		});

		it('returns false and clears state when /me fails', async () => {
			vi.mocked(api.me).mockRejectedValueOnce(new ApiError(401, 'invalid or expired session'));

			const result = await refreshAuth();

			expect(result).toBe(false);
			expect(get(isAuthed)).toBe(false);
			expect(get(currentUser)).toBeNull();
		});
	});
});
