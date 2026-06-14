<script lang="ts">
	import { page } from '$app/state';
	import { toast } from 'svelte-sonner';
	import { api } from '$lib/api';
	import { CURRENT_POLICY_VERSION } from '$lib/policy';
	import Wordmark from '$lib/components/Wordmark.svelte';

	const email = $derived(page.url.searchParams.get('email') ?? '');

	async function handleResend() {
		try {
			await api.requestMagicLink(email, CURRENT_POLICY_VERSION);
			toast('Magic link resent', {
				description: `We sent another link to ${email || 'your inbox'}.`
			});
		} catch {
			toast.error("Couldn't reach the server", {
				description: 'Is the backend running?'
			});
		}
	}
</script>

<div class="bg-background flex flex-1 items-center justify-center px-4 py-12">
	<div class="w-full max-w-sm space-y-6">
		<div class="flex flex-col items-center space-y-2 text-center">
			<Wordmark href="/" />
			<h1 class="text-xl font-semibold">Check your email</h1>
			<p class="text-muted-foreground text-sm">
				We sent a sign-in link to
				{#if email}
					<span class="text-foreground font-medium">{email}</span>
				{:else}
					your inbox
				{/if}.
			</p>
		</div>

		<div class="flex flex-col items-center gap-3">
			<button
				type="button"
				onclick={handleResend}
				class="text-muted-foreground hover:text-foreground focus-visible:ring-ring rounded-sm text-sm underline-offset-4 hover:underline focus-visible:ring-2 focus-visible:outline-hidden"
			>
				Didn't get it? Resend
			</button>
			<a
				href="/auth"
				class="text-muted-foreground hover:text-foreground focus-visible:ring-ring rounded-sm text-xs underline-offset-4 hover:underline focus-visible:ring-2 focus-visible:outline-hidden"
			>
				Use a different email
			</a>
		</div>


	</div>
</div>
