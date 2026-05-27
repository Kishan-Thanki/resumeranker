<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { currentUser, refreshAuth } from '$lib/stores/auth';
	import { refreshAll } from '$lib/stores/analyses';
	import { CURRENT_POLICY_VERSION } from '$lib/policy';
	import TopBar from '$lib/components/TopBar.svelte';
	import ReacceptanceDialog from '$lib/components/ReacceptanceDialog.svelte';

	let { children } = $props();

	let ready = $state(false);
	let reacceptOpen = $state(false);

	// Re-acceptance is required whenever the stored version is strictly older
	// than the bundle's CURRENT_POLICY_VERSION. `null` (pre-policy backfilled
	// users) does NOT trigger the modal — they were grandfathered by the
	// initial migration.
	$effect(() => {
		const user = $currentUser;
		if (!user) {
			reacceptOpen = false;
			return;
		}
		const stored = user.acceptedPolicyVersion;
		reacceptOpen = stored !== null && stored < CURRENT_POLICY_VERSION;
	});

	onMount(async () => {
		// Validate the session against the backend before rendering the app
		// shell. Cleans up stale localStorage tokens too.
		const authed = await refreshAuth();
		if (!authed) {
			await goto('/');
			return;
		}
		try {
			await refreshAll();
		} catch {
			// Non-fatal: routes that need the list will retry.
		}
		ready = true;
	});
</script>

{#if ready}
	<div class="bg-background min-h-screen">
		<TopBar />
		<main class="mx-auto max-w-5xl px-4 py-8 sm:px-6">
			{@render children()}
		</main>
	</div>
	<ReacceptanceDialog bind:open={reacceptOpen} />
{/if}
