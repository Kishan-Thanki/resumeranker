<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import { CURRENT_POLICY_VERSION } from '$lib/policy';
    import ThemeToggle from '$lib/components/ThemeToggle.svelte';

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
			error = "Couldn't reach the server. Is the backend running?";
		} finally {
			submitting = false;
		}
	}
</script>

<div class="absolute right-4 top-4">
    <ThemeToggle />
</div>

<div class="bg-background flex flex-1 items-center justify-center px-4 py-12">
	<form onsubmit={handleSubmit} class="w-full max-w-sm space-y-6">
		<div class="flex flex-col items-center space-y-2 text-center">
			<div class="mb-4 flex items-center justify-center h-12 w-12 rounded-full bg-primary/10">
				<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="h-6 w-6 text-primary">
				  <path d="M20 13c0 5-3.5 7.5-7.66 8.95a1 1 0 0 1-.67-.01C7.5 20.5 4 18 4 13V6a1 1 0 0 1 1-1c2-1 4-2 7-2 3 0 5 1 7 2a1 1 0 0 1 1 1v7z"/>
				  <path d="m9 12 2 2 4-4"/>
				</svg>
			</div>
			<h1 class="text-2xl font-semibold tracking-tight">Admin Sign In</h1>
			<p class="text-muted-foreground text-sm">
				Enter your email to access the CMS admin panel.
			</p>
		</div>

		<div class="space-y-2">
			<label for="email" class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">Email</label>
			<input
				id="email"
				type="email"
				autocomplete="email"
				required
				bind:value={email}
				placeholder="admin@example.com"
                class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
			/>
		</div>

		<div class="flex items-start gap-3">
			<input
				id="accept-policy"
				type="checkbox"
				required
				bind:checked={acceptedPolicy}
				class="border-input ring-offset-background focus-visible:ring-ring data-[state=checked]:bg-primary data-[state=checked]:text-primary-foreground mt-0.5 h-4 w-4 shrink-0 rounded-sm border accent-zinc-900 outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
			/>
			<label for="accept-policy" class="text-muted-foreground text-sm leading-relaxed font-normal">
				I agree to the Terms of Service and Privacy Policy.
			</label>
		</div>

		{#if error}
			<p class="text-destructive text-sm font-medium">{error}</p>
		{/if}

		<button
			type="submit"
			class="inline-flex h-10 w-full items-center justify-center whitespace-nowrap rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground ring-offset-background transition-colors hover:bg-primary/90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50"
			disabled={!email.trim() || !acceptedPolicy || submitting}
		>
			{submitting ? 'Sending...' : 'Send magic link'}
		</button>
	</form>
</div>
