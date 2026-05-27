import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { api, ApiError } from './api';

const BASE = 'http://localhost:8000';

interface FetchMock {
	calls: Array<{ url: string; init?: RequestInit }>;
	queue: Array<Response | Error>;
	enqueueJson: (status: number, body: unknown) => void;
	enqueueText: (status: number, body: string) => void;
	enqueueEmpty: (status: number) => void;
	enqueueError: (err: Error) => void;
}

function makeFetchMock(): FetchMock {
	const calls: FetchMock['calls'] = [];
	const queue: FetchMock['queue'] = [];

	globalThis.fetch = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
		calls.push({ url: String(input), init });
		const next = queue.shift();
		if (next === undefined) throw new Error(`fetch called with no queued response: ${input}`);
		if (next instanceof Error) throw next;
		return next;
	}) as unknown as typeof fetch;

	return {
		calls,
		queue,
		enqueueJson(status, body) {
			queue.push(
				new Response(JSON.stringify(body), {
					status,
					headers: { 'Content-Type': 'application/json' }
				})
			);
		},
		enqueueText(status, body) {
			queue.push(new Response(body, { status }));
		},
		enqueueEmpty(status) {
			queue.push(new Response(null, { status }));
		},
		enqueueError(err) {
			queue.push(err);
		}
	};
}

describe('api client', () => {
	let mock: FetchMock;

	beforeEach(() => {
		mock = makeFetchMock();
	});

	afterEach(() => {
		vi.restoreAllMocks();
	});

	describe('requestMagicLink', () => {
		it('POSTs JSON to /auth/request-link with credentials', async () => {
			mock.enqueueJson(200, { ok: true });
			await api.requestMagicLink('a@b.com', '2026-05-19');
			expect(mock.calls).toHaveLength(1);
			expect(mock.calls[0].url).toBe(`${BASE}/auth/request-link`);
			expect(mock.calls[0].init?.method).toBe('POST');
			expect(mock.calls[0].init?.credentials).toBe('include');
			expect(JSON.parse(mock.calls[0].init?.body as string)).toEqual({
				email: 'a@b.com',
				acceptedPolicyVersion: '2026-05-19'
			});
		});

		it('throws ApiError on non-2xx', async () => {
			mock.enqueueJson(500, { detail: 'kaboom' });
			await expect(api.requestMagicLink('a@b.com', '2026-05-19')).rejects.toBeInstanceOf(ApiError);
		});
	});

	describe('verifyMagicLink', () => {
		it('returns the parsed user payload (no session token in body)', async () => {
			// The backend sets an HttpOnly cookie via Set-Cookie — there's
			// no `sessionToken` in the JSON. The browser stores the cookie
			// transparently.
			mock.enqueueJson(200, {
				user: { id: 'u-1', email: 'a@b.com', acceptedPolicyVersion: '2026-05-19' }
			});
			const result = await api.verifyMagicLink('raw-token');
			expect(result.user.email).toBe('a@b.com');
			expect(mock.calls[0].init?.credentials).toBe('include');
		});

		it('throws ApiError with detail on 400', async () => {
			mock.enqueueJson(400, { detail: 'invalid link' });
			try {
				await api.verifyMagicLink('bad-token');
				expect.fail('should have thrown');
			} catch (e) {
				expect(e).toBeInstanceOf(ApiError);
				expect((e as ApiError).status).toBe(400);
				expect((e as ApiError).message).toBe('invalid link');
			}
		});
	});

	describe('me', () => {
		it('sends credentials so the session cookie travels', async () => {
			mock.enqueueJson(200, { id: 'u-1', email: 'a@b.com', acceptedPolicyVersion: null });
			await api.me();
			expect(mock.calls[0].init?.credentials).toBe('include');
			// No Authorization header — auth is via cookie now.
			const headers = (mock.calls[0].init?.headers as Record<string, string>) ?? {};
			expect(headers['Authorization']).toBeUndefined();
		});

		it('throws ApiError on 401', async () => {
			mock.enqueueJson(401, { detail: 'invalid or expired session' });
			await expect(api.me()).rejects.toBeInstanceOf(ApiError);
		});
	});

	describe('signOut', () => {
		it('POSTs to /auth/sign-out with credentials', async () => {
			mock.enqueueEmpty(204);
			await api.signOut();
			expect(mock.calls[0].url).toBe(`${BASE}/auth/sign-out`);
			expect(mock.calls[0].init?.method).toBe('POST');
			expect(mock.calls[0].init?.credentials).toBe('include');
		});

		it('handles 204 with no body', async () => {
			mock.enqueueEmpty(204);
			await expect(api.signOut()).resolves.toBeUndefined();
		});
	});

	describe('listAnalyses', () => {
		it('returns array', async () => {
			mock.enqueueJson(200, [{ id: 'a-1', jdTitle: 'Foo' }]);
			const result = await api.listAnalyses();
			expect(Array.isArray(result)).toBe(true);
			expect(result[0].id).toBe('a-1');
			expect(mock.calls[0].init?.credentials).toBe('include');
		});
	});

	describe('getAnalysis', () => {
		it('encodes the id in the URL', async () => {
			mock.enqueueJson(200, { id: 'a%20b', jdTitle: 'X' });
			await api.getAnalysis('a b');
			expect(mock.calls[0].url).toBe(`${BASE}/analyses/a%20b`);
			expect(mock.calls[0].init?.credentials).toBe('include');
		});

		it('throws 404 ApiError when missing', async () => {
			mock.enqueueJson(404, { detail: 'not found' });
			try {
				await api.getAnalysis('missing');
				expect.fail('should have thrown');
			} catch (e) {
				expect((e as ApiError).status).toBe(404);
			}
		});
	});

	describe('createAnalysis', () => {
		it('sends multipart form with text JD', async () => {
			mock.enqueueJson(201, { id: 'a-new', status: 'queued' });
			const resume = new File(['%PDF-1.4'], 'resume.pdf', { type: 'application/pdf' });
			await api.createAnalysis({ jdInputType: 'text', jdText: 'hello', resume });
			const fd = mock.calls[0].init?.body as FormData;
			expect(fd).toBeInstanceOf(FormData);
			expect(fd.get('jd_input_type')).toBe('text');
			expect(fd.get('jd_text')).toBe('hello');
			expect(fd.get('resume')).toBeInstanceOf(File);
			expect(mock.calls[0].init?.credentials).toBe('include');
		});

		it('sends multipart form with PDF JD', async () => {
			mock.enqueueJson(201, { id: 'a-pdf', status: 'queued' });
			const resume = new File(['%PDF-1.4'], 'r.pdf', { type: 'application/pdf' });
			const jdPdf = new File(['%PDF-1.4'], 'jd.pdf', { type: 'application/pdf' });
			await api.createAnalysis({ jdInputType: 'pdf', jdPdf, resume });
			const fd = mock.calls[0].init?.body as FormData;
			expect(fd.get('jd_input_type')).toBe('pdf');
			expect(fd.get('jd_pdf')).toBeInstanceOf(File);
			expect(fd.get('jd_text')).toBeNull();
		});

		it('surfaces 422 as ApiError', async () => {
			mock.enqueueJson(422, { detail: 'PDF text too short' });
			const resume = new File([''], 'bad.pdf', { type: 'application/pdf' });
			try {
				await api.createAnalysis({ jdInputType: 'text', jdText: 'x', resume });
				expect.fail('should have thrown');
			} catch (e) {
				expect((e as ApiError).status).toBe(422);
				expect((e as ApiError).message).toBe('PDF text too short');
			}
		});
	});
});
