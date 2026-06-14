<script lang="ts">
	import { page } from '$app/state';
	import { api } from '$lib/api';
	import { CURRENT_POLICY_VERSION } from '$lib/policy';

	const email = $derived(page.url.searchParams.get('email') ?? '');
    let resendMessage = $state<string | null>(null);

	async function handleResend() {
		try {
			await api.requestMagicLink(email, CURRENT_POLICY_VERSION);
            resendMessage = `We sent another link to ${email || 'your inbox'}.`;
		} catch {
            resendMessage = "Couldn't reach the server. Is the backend running?";
		}
	}
</script>

<div class="bg-background flex flex-1 items-center justify-center px-4 py-12 text-slate-800">
	<div class="w-full max-w-sm space-y-6">
		<div class="flex flex-col items-center space-y-2 text-center">
			<h1 class="text-xl font-semibold">Check your email</h1>
			<p class="text-slate-500 text-sm">
				We sent an admin sign-in link to
				{#if email}
					<span class="text-slate-900 font-medium">{email}</span>
				{:else}
					your inbox
				{/if}.
			</p>
		</div>

        {#if resendMessage}
            <div class="rounded bg-slate-100 p-4 text-sm text-slate-700 text-center">
                {resendMessage}
            </div>
        {/if}

		<div class="flex flex-col items-center gap-3">
			<button
				type="button"
				onclick={handleResend}
				class="text-slate-500 hover:text-slate-900 text-sm underline-offset-4 hover:underline focus-visible:outline-none"
			>
				Didn't get it? Resend
			</button>
			<a
				href="/auth"
				class="text-slate-500 hover:text-slate-900 text-xs underline-offset-4 hover:underline focus-visible:outline-none"
			>
				Use a different email
			</a>
		</div>
	</div>
</div>
