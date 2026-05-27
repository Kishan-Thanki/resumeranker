import { writable, get, type Readable } from 'svelte/store';
import { browser } from '$app/environment';
import { api, ApiError } from '$lib/api';
import type { AnalysisResult } from '$lib/types';

const internal = writable<AnalysisResult[]>([]);

function sortDesc(list: AnalysisResult[]): AnalysisResult[] {
	return [...list].sort((a, b) => (a.createdAt < b.createdAt ? 1 : -1));
}

export const analyses: Readable<AnalysisResult[]> = {
	subscribe: (run, invalidate) =>
		internal.subscribe((value) => run(sortDesc(value)), invalidate)
};

export function getAll(): AnalysisResult[] {
	return sortDesc(get(internal));
}

export function getById(id: string): AnalysisResult | undefined {
	return get(internal).find((a) => a.id === id);
}

function upsert(updated: AnalysisResult): void {
	internal.update((list) => {
		const idx = list.findIndex((a) => a.id === updated.id);
		if (idx >= 0) {
			const copy = [...list];
			copy[idx] = updated;
			return copy;
		}
		return [updated, ...list];
	});
}

/**
 * Fetch all analyses for the current user and replace the store contents.
 * Called from the app shell on mount and after sign-in.
 */
export async function refreshAll(): Promise<AnalysisResult[]> {
	if (!browser) return [];
	const list = await api.listAnalyses();
	internal.set(list);
	return sortDesc(list);
}

/**
 * Fetch one analysis from the backend and update the store. Returns
 * `undefined` if not found / not owned (backend returns 404 either way).
 * Used by the detail page on mount and inside the polling loop.
 */
export async function refreshById(id: string): Promise<AnalysisResult | undefined> {
	if (!browser) return undefined;
	try {
		const updated = await api.getAnalysis(id);
		upsert(updated);
		return updated;
	} catch (e) {
		if (e instanceof ApiError && e.status === 404) return undefined;
		throw e;
	}
}

/**
 * POST a new analysis. Returns the freshly-created (queued) result.
 * The detail page picks up polling from here.
 */
export async function create(input: {
	jdInputType: 'pdf' | 'text';
	jdText?: string;
	jdPdf?: File;
	resume: File;
}): Promise<AnalysisResult> {
	const created = await api.createAnalysis(input);
	upsert(created);
	return created;
}

/**
 * Delete one analysis on the backend and drop it from the local cache.
 * Backend hard-deletes the row (encrypted content + scores). A 404
 * (already gone / not owned) is treated as success — the end state the
 * caller wanted is reached either way.
 */
export async function remove(id: string): Promise<void> {
	if (!browser) return;
	try {
		await api.deleteAnalysis(id);
	} catch (e) {
		if (!(e instanceof ApiError && e.status === 404)) throw e;
	}
	internal.update((list) => list.filter((a) => a.id !== id));
}

/**
 * Clear the local cache. Called on sign-out so a different user's
 * analyses don't leak into the next session.
 */
export function clearAnalyses(): void {
	internal.set([]);
}
