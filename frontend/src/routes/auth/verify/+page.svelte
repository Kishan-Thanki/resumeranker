<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button';
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
			await goto('/app');
		} catch {
			error = 'This link is invalid or has expired. Request a new one.';
		}
	});
</script>

<div class="bg-background flex flex-1 items-center justify-center px-4 py-12">
	<div class="text-center">
		{#if error}
			<p class="text-sm font-medium">{error}</p>
			<div class="mt-4">
				<Button href="/auth" variant="outline">Back to sign in</Button>
			</div>
		{:else}
			<div class="flex items-center justify-center gap-1.5 text-indigo-500">
				<span class="bg-current size-2 animate-pulse rounded-full"></span>
				<span class="bg-current size-2 animate-pulse rounded-full [animation-delay:120ms]"></span>
				<span class="bg-current size-2 animate-pulse rounded-full [animation-delay:240ms]"></span>
			</div>
			<p class="text-muted-foreground mt-4 text-sm">Verifying...</p>
		{/if}
	</div>
</div>
