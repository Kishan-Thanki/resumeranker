import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { get } from 'svelte/store';

import { api, ApiError } from '$lib/api';
import type { AnalysisResult } from '$lib/types';

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

const { analyses, getAll, getById, refreshAll, refreshById, create, clearAnalyses } = await import(
	'./analyses'
);

function makeAnalysis(over: Partial<AnalysisResult> = {}): AnalysisResult {
	return {
		id: 'a-1',
		createdAt: '2026-05-01T12:00:00.000Z',
		jdTitle: 'Senior Backend Engineer',
		resumeName: 'resume.pdf',
		status: 'completed',
		sections: [],
		...over
	};
}

describe('analyses store', () => {
	beforeEach(() => {
		clearAnalyses();
		vi.clearAllMocks();
	});

	afterEach(() => {
		vi.restoreAllMocks();
	});

	describe('refreshAll', () => {
		it('replaces store contents and sorts by createdAt desc', async () => {
			const older = makeAnalysis({ id: 'a-old', createdAt: '2026-04-01T00:00:00Z' });
			const newer = makeAnalysis({ id: 'a-new', createdAt: '2026-05-15T00:00:00Z' });
			vi.mocked(api.listAnalyses).mockResolvedValueOnce([older, newer]);

			const sorted = await refreshAll();

			expect(sorted.map((a) => a.id)).toEqual(['a-new', 'a-old']);
			expect(getAll().map((a) => a.id)).toEqual(['a-new', 'a-old']);
		});

		it('propagates api errors', async () => {
			vi.mocked(api.listAnalyses).mockRejectedValueOnce(new ApiError(401, 'unauthorized'));
			await expect(refreshAll()).rejects.toBeInstanceOf(ApiError);
		});
	});

	describe('refreshById', () => {
		it('upserts a new analysis into an empty store', async () => {
			const a = makeAnalysis({ id: 'a-new', status: 'completed' });
			vi.mocked(api.getAnalysis).mockResolvedValueOnce(a);

			const result = await refreshById('a-new');

			expect(result?.id).toBe('a-new');
			expect(getById('a-new')?.status).toBe('completed');
		});

		it('replaces an existing entry in-place', async () => {
			const queued = makeAnalysis({ id: 'a-1', status: 'queued', sections: [] });
			const completed = makeAnalysis({
				id: 'a-1',
				status: 'completed',
				sections: [{ id: 'skills', label: 'Skills', score: 80, requirements: [] }]
			});
			vi.mocked(api.listAnalyses).mockResolvedValueOnce([queued]);
			vi.mocked(api.getAnalysis).mockResolvedValueOnce(completed);

			await refreshAll();
			expect(getById('a-1')?.status).toBe('queued');

			await refreshById('a-1');
			expect(getById('a-1')?.status).toBe('completed');
			expect(getById('a-1')?.sections).toHaveLength(1);
			// Store should not have duplicated.
			expect(get(analyses).filter((a) => a.id === 'a-1')).toHaveLength(1);
		});

		it('returns undefined on 404 instead of throwing', async () => {
			vi.mocked(api.getAnalysis).mockRejectedValueOnce(new ApiError(404, 'not found'));
			const result = await refreshById('missing');
			expect(result).toBeUndefined();
		});

		it('propagates non-404 ApiErrors', async () => {
			vi.mocked(api.getAnalysis).mockRejectedValueOnce(new ApiError(500, 'boom'));
			await expect(refreshById('a-1')).rejects.toBeInstanceOf(ApiError);
		});
	});

	describe('create', () => {
		it('POSTs and inserts the new analysis at the top', async () => {
			const existing = makeAnalysis({ id: 'a-old', createdAt: '2026-04-01T00:00:00Z' });
			vi.mocked(api.listAnalyses).mockResolvedValueOnce([existing]);
			await refreshAll();

			const created = makeAnalysis({
				id: 'a-new',
				createdAt: '2026-05-17T00:00:00Z',
				status: 'queued'
			});
			vi.mocked(api.createAnalysis).mockResolvedValueOnce(created);

			const resume = new File(['%PDF-1.4'], 'r.pdf', { type: 'application/pdf' });
			const result = await create({ jdInputType: 'text', jdText: 'JD', resume });

			expect(result.id).toBe('a-new');
			expect(getAll().map((a) => a.id)).toEqual(['a-new', 'a-old']);
		});
	});

	describe('clearAnalyses', () => {
		it('empties the store', async () => {
			vi.mocked(api.listAnalyses).mockResolvedValueOnce([makeAnalysis()]);
			await refreshAll();
			expect(getAll()).toHaveLength(1);

			clearAnalyses();
			expect(getAll()).toEqual([]);
		});
	});

	describe('analyses readable', () => {
		it('emits sorted snapshots when subscribed', async () => {
			const a = makeAnalysis({ id: 'a-1', createdAt: '2026-05-01T00:00:00Z' });
			const b = makeAnalysis({ id: 'a-2', createdAt: '2026-05-15T00:00:00Z' });
			vi.mocked(api.listAnalyses).mockResolvedValueOnce([a, b]);
			await refreshAll();

			const snapshots: string[][] = [];
			const unsub = analyses.subscribe((list) => {
				snapshots.push(list.map((item) => item.id));
			});
			unsub();

			// Last emission is the current sorted state.
			expect(snapshots.at(-1)).toEqual(['a-2', 'a-1']);
		});
	});
});
