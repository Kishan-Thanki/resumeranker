<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { signInWithToken } from '$lib/stores/auth';

	let error = $state<string | null>(null);

	onMount(async () => {
		const token = page.url.searchParams.get('token');
		if (!token) {
			error = 'No token in URL. Open the magic link from your email again.';
			return;
		}
		try {
			await signInWithToken(token);
			// Redirect to admin IAM dashboard root
			await goto('/');
		} catch {
			error = 'This link is invalid or has expired. Request a new one.';
		}
	});
</script>

<div class="bg-background flex flex-1 items-center justify-center px-4 py-12 text-slate-800">
	<div class="text-center">
		{#if error}
			<p class="text-sm font-medium text-red-500">{error}</p>
			<div class="mt-4">
				<a href="/auth" class="inline-flex h-10 items-center justify-center rounded-md border border-slate-300 bg-white px-4 py-2 text-sm font-medium hover:bg-slate-100 text-slate-900">
                    Back to sign in
                </a>
			</div>
		{:else}
			<div class="flex items-center justify-center gap-1.5 text-slate-500">
				<span class="bg-current size-2 animate-pulse rounded-full"></span>
				<span class="bg-current size-2 animate-pulse rounded-full [animation-delay:120ms]"></span>
				<span class="bg-current size-2 animate-pulse rounded-full [animation-delay:240ms]"></span>
			</div>
			<p class="text-slate-500 mt-4 text-sm">Verifying admin session...</p>
		{/if}
	</div>
</div>
