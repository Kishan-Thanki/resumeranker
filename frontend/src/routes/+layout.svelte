<script lang="ts">
	import '../app.css';
	import '@fontsource-variable/inter/index.css';
	import '@fontsource/jetbrains-mono/400.css';
	import '@fontsource/jetbrains-mono/500.css';
	import { ModeWatcher } from 'mode-watcher';
	import { Toaster } from '$lib/components/ui/sonner';
	import Footer from '$lib/components/Footer.svelte';
	import favicon from '$lib/assets/favicon.svg';

	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>Resume Ranker</title>
</svelte:head>

<ModeWatcher />
<Toaster />

<!--
	Flex column makes the Footer stick to the bottom of the viewport on short
	pages, while still pushing down naturally on long ones. `min-h-screen` on
	the wrapper plus `flex-1` on the content slot is the standard pattern.
	The existing per-page `min-h-screen` containers still work — they just
	now grow to fill at least the viewport height of the wrapper.
-->
<div class="flex min-h-screen flex-col">
	<div class="flex-1">
		{@render children()}
	</div>
	<Footer />
</div>
