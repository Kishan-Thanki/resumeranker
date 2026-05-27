<script lang="ts">
	import { Moon, Sun } from '@lucide/svelte';
	import { toggleMode, mode } from 'mode-watcher';
	import { Button } from '$lib/components/ui/button';

	// Optional extra classes for positioning (e.g. a fixed corner on
	// pages that have no top bar of their own).
	let { class: className = '' }: { class?: string } = $props();
</script>

<Button
	variant="ghost"
	size="icon"
	onclick={toggleMode}
	aria-label="Toggle color theme"
	class={className}
>
	{#if mode.current === 'dark'}
		<Sun class="size-4" />
	{:else}
		<Moon class="size-4" />
	{/if}
</Button>
