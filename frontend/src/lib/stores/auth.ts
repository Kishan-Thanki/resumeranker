import { writable, get } from 'svelte/store';
import { browser } from '$app/environment';
import { api, type MeResponse } from '$lib/api';

// Auth state can no longer be inferred from a JS-readable storage slot —
// the session token lives in an HttpOnly cookie that's deliberately
// invisible to this code. Initial value is `false`; the layout calls
// `refreshAuth()` on mount which probes /me and flips this true when
// the cookie is valid.
export const isAuthed = writable<boolean>(false);

// Cached current user record. Populated by refreshAuth() and updated by
// acceptCurrentPolicy() so the re-acceptance dialog can read/write without
// re-fetching /me.
export const currentUser = writable<MeResponse | null>(null);

/**
 * Exchange a magic-link token for a session. The backend sets an
 * HttpOnly `session` cookie on the response; there is no token returned
 * to JS. After this resolves the browser is authenticated for any
 * subsequent fetch with `credentials: 'include'`.
 *
 * The verify response also carries the user record so we can populate
 * `currentUser` without an extra /me round-trip.
 */
export async function signInWithToken(token: string): Promise<void> {
	if (!browser) return;
	const resp = await api.verifyMagicLink(token);
	currentUser.set(resp.user);
	isAuthed.set(true);
}

/**
 * Revoke the current session on the backend and clear local state.
 * Backend errors are swallowed — the user is signing out anyway, a
 * stale server-side session is harmless once it expires.
 */
export async function signOut(): Promise<void> {
	if (!browser) return;
	try {
		await api.signOut();
	} catch {
		/* ignore */
	}
	currentUser.set(null);
	isAuthed.set(false);
}

/**
 * Hard-delete the current user's account on the backend and clear local
 * state. Backend cascades to sessions, magic_links, and analyses, and
 * also clears the session cookie on the same response. Caller is
 * responsible for navigating away after this resolves.
 *
 * Errors are NOT swallowed — the caller needs to know if the delete
 * actually succeeded, since the in-app "your data is gone" message
 * depends on it.
 */
export async function deleteAccount(): Promise<void> {
	if (!browser) return;
	await api.deleteAccount();
	currentUser.set(null);
	isAuthed.set(false);
}

/**
 * Probe the backend for the current user. Used by the app shell to
 * determine whether the session cookie (if any) is still valid. Returns
 * true if /me succeeded, false otherwise.
 *
 * Unlike the localStorage era, we can't short-circuit on "no token
 * present" — JS can't see the HttpOnly cookie. The single /me call is
 * always-fire-once: it either succeeds (auth'd) or returns 401 (not
 * auth'd, including the missing-cookie case).
 */
export async function refreshAuth(): Promise<boolean> {
	if (!browser) return false;
	try {
		const me = await api.me();
		currentUser.set(me);
		isAuthed.set(true);
		return true;
	} catch {
		currentUser.set(null);
		isAuthed.set(false);
		return false;
	}
}

/**
 * Re-record the user's acceptance of the latest policy version. Updates
 * the cached `currentUser` so the gating dialog dismisses without a
 * separate /me round-trip.
 */
export async function acceptCurrentPolicy(version: string): Promise<void> {
	if (!browser) return;
	const updated = await api.acceptPolicy(version);
	currentUser.set(updated);
}

export function isCurrentlyAuthed(): boolean {
	return get(isAuthed);
}
