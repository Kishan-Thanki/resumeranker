import type { AnalysisResult } from '$lib/types';

const BASE = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8000';

// Every request must include credentials so the browser sends the
// HttpOnly `session` cookie set by /auth/verify. Set-Cookie headers from
// authenticated responses (verify, sign-out, delete) are stored by the
// browser the same way, transparent to this client.
const CREDENTIALS: RequestCredentials = 'include';

export class ApiError extends Error {
	constructor(
		public status: number,
		message: string
	) {
		super(message);
	}
}

async function handle<T>(res: Response): Promise<T> {
	if (!res.ok) {
		let detail = res.statusText;
		try {
			const body = await res.json();
			if (body && typeof body.detail === 'string') detail = body.detail;
		} catch {
			/* not JSON */
		}
		throw new ApiError(res.status, detail);
	}
	if (res.status === 204) return undefined as T;
	return (await res.json()) as T;
}

export interface MeResponse {
	id: string;
	email: string;
	// Nullable for backfilled pre-policy users.
	acceptedPolicyVersion: string | null;
	role: 'user' | 'admin' | 'superadmin';
}

export interface VerifyResponse {
	// No `sessionToken` field — the backend sets an HttpOnly cookie via
	// Set-Cookie. JS can't read it, which is the whole point.
	user: MeResponse;
}

export const api = {
	async requestMagicLink(email: string, acceptedPolicyVersion: string): Promise<void> {
		await handle<{ ok: boolean }>(
			await fetch(`${BASE}/auth/request-link`, {
				method: 'POST',
				credentials: CREDENTIALS,
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ email, acceptedPolicyVersion })
			})
		);
	},

	async verifyMagicLink(token: string): Promise<VerifyResponse> {
		return handle<VerifyResponse>(
			await fetch(`${BASE}/auth/verify`, {
				method: 'POST',
				credentials: CREDENTIALS,
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ token })
			})
		);
	},

	async me(): Promise<MeResponse> {
		return handle<MeResponse>(
			await fetch(`${BASE}/me`, { credentials: CREDENTIALS })
		);
	},

	async sendContactMessage(input: {
		name: string;
		email: string;
		message: string;
		website?: string;
	}): Promise<void> {
		// Backend always returns 200 (success / rate-limited / honeypot-tripped
		// all look identical to the caller). Spam-prevention by-design.
		await handle<{ ok: boolean }>(
			await fetch(`${BASE}/contact`, {
				method: 'POST',
				credentials: CREDENTIALS,
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					name: input.name,
					email: input.email,
					message: input.message,
					website: input.website ?? ''
				})
			})
		);
	},

	async signOut(): Promise<void> {
		await handle<void>(
			await fetch(`${BASE}/auth/sign-out`, {
				method: 'POST',
				credentials: CREDENTIALS
			})
		);
	},

	async acceptPolicy(version: string): Promise<MeResponse> {
		// Re-acceptance flow. Returns the updated /me payload so callers
		// can refresh their cached user record without a separate GET.
		return handle<MeResponse>(
			await fetch(`${BASE}/me/accept-policy`, {
				method: 'POST',
				credentials: CREDENTIALS,
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ acceptedPolicyVersion: version })
			})
		);
	},

	async deleteAccount(): Promise<void> {
		// Hard-delete on the backend cascades to sessions, magic_links, and
		// analyses. The session cookie is cleared by the backend on the
		// same response; callers just need to refresh local UI state.
		await handle<void>(
			await fetch(`${BASE}/me`, {
				method: 'DELETE',
				credentials: CREDENTIALS
			})
		);
	},

	async listAnalyses(): Promise<AnalysisResult[]> {
		return handle<AnalysisResult[]>(
			await fetch(`${BASE}/analyses`, { credentials: CREDENTIALS })
		);
	},

	async getAnalysis(id: string): Promise<AnalysisResult> {
		return handle<AnalysisResult>(
			await fetch(`${BASE}/analyses/${encodeURIComponent(id)}`, {
				credentials: CREDENTIALS
			})
		);
	},

	async deleteAnalysis(id: string): Promise<void> {
		// Removes the analysis' encrypted content and scores. Backend
		// returns 404 for a missing / non-owned id (no existence leak).
		await handle<void>(
			await fetch(`${BASE}/analyses/${encodeURIComponent(id)}`, {
				method: 'DELETE',
				credentials: CREDENTIALS
			})
		);
	},

	async createAnalysis(input: {
		jdInputType: 'pdf' | 'text';
		jdText?: string;
		jdPdf?: File;
		resume: File;
	}): Promise<AnalysisResult> {
		const fd = new FormData();
		fd.append('jd_input_type', input.jdInputType);
		fd.append('resume', input.resume);
		if (input.jdInputType === 'text') {
			fd.append('jd_text', input.jdText ?? '');
		} else if (input.jdPdf) {
			fd.append('jd_pdf', input.jdPdf);
		}
		return handle<AnalysisResult>(
			await fetch(`${BASE}/analyses`, {
				method: 'POST',
				credentials: CREDENTIALS,
				body: fd
			})
		);
	}
};
