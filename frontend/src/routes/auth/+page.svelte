<script lang="ts">
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { api } from '$lib/api';
	import { CURRENT_POLICY_VERSION } from '$lib/policy';

	let email = $state('');
	let acceptedPolicy = $state(false);
	let submitting = $state(false);
	let error = $state<string | null>(null);

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		const trimmed = email.trim();
		if (!trimmed || !acceptedPolicy) return;
		submitting = true;
		error = null;
		try {
			await api.requestMagicLink(trimmed, CURRENT_POLICY_VERSION);
			await goto(`/auth/sent?email=${encodeURIComponent(trimmed)}`);
		} catch {
			// Backend always returns 200 to prevent email enumeration. A real
			// failure here means the network is down or the API isn't running.
			error = "Couldn't reach the server. Is the backend running?";
		} finally {
			submitting = false;
		}
	}
</script>

<div class="bg-background flex min-h-screen items-center justify-center px-4 py-12">
	<form onsubmit={handleSubmit} class="w-full max-w-sm space-y-6">
		<div class="space-y-2 text-center">
			<a href="/" class="text-sm font-semibold tracking-tight">Resume Ranker</a>
			<h1 class="text-xl font-semibold">Sign in or create account</h1>
			<p class="text-muted-foreground text-sm">
				Enter your email &mdash; we'll send a one-time link. New accounts are
				created on first sign-in.
			</p>
		</div>

		<div class="space-y-2">
			<Label for="email">Email</Label>
			<Input
				id="email"
				type="email"
				autocomplete="email"
				required
				bind:value={email}
				placeholder="you@example.com"
			/>
		</div>

		<div class="flex items-start gap-3">
			<input
				id="accept-policy"
				type="checkbox"
				required
				bind:checked={acceptedPolicy}
				class="border-input ring-offset-background focus-visible:ring-ring data-[state=checked]:bg-primary data-[state=checked]:text-primary-foreground mt-0.5 h-4 w-4 shrink-0 rounded-sm border accent-zinc-900 outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 dark:accent-zinc-100"
			/>
			<Label for="accept-policy" class="text-muted-foreground text-xs leading-relaxed font-normal">
				I agree to the
				<a class="text-foreground underline hover:no-underline" href="/terms" target="_blank" rel="noopener">Terms of Service</a>
				and
				<a class="text-foreground underline hover:no-underline" href="/privacy" target="_blank" rel="noopener">Privacy Policy</a>.
			</Label>
		</div>

		{#if error}
			<p class="text-destructive text-sm">{error}</p>
		{/if}

		<Button
			type="submit"
			class="w-full"
			disabled={!email.trim() || !acceptedPolicy || submitting}
		>
			{submitting ? 'Sending...' : 'Send magic link'}
		</Button>
	</form>
</div>
