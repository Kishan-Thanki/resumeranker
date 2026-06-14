<script lang="ts">
	import '../app.css';
	import '@fontsource-variable/inter/index.css';
	import '@fontsource/jetbrains-mono/400.css';
	import '@fontsource/jetbrains-mono/500.css';
	import { ModeWatcher } from 'mode-watcher';
	import { page } from '$app/state';
	import { Toaster } from '$lib/components/ui/sonner';
	import Footer from '$lib/components/Footer.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import favicon from '$lib/assets/favicon.svg';

	let { children } = $props();

	// The /app shell has its own theme toggle in the TopBar. For every
	// other (public) route — landing, auth, legal pages — there's no top
	// bar, so render a single global toggle pinned to the top-right.
	const showGlobalToggle = $derived(!page.url.pathname.startsWith('/app'));
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>Resume Ranker</title>
</svelte:head>

<ModeWatcher defaultMode="system" track={true} />
<Toaster />

<!--
	Flex column makes the Footer stick to the bottom of the viewport on short
	pages, while still pushing down naturally on long ones. `min-h-screen` on
	the wrapper plus `flex-1` on the content slot is the standard pattern.
	The existing per-page `min-h-screen` containers still work — they just
	now grow to fill at least the viewport height of the wrapper.
-->
{#if showGlobalToggle}
	<ThemeToggle class="fixed top-3 right-3 z-50" />
{/if}

<div class="flex min-h-screen flex-col">
	<div class="flex flex-1 flex-col">
		{@render children()}
	</div>
	<Footer />
</div>
