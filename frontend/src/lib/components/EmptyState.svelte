<script lang="ts">
	import type { Component, Snippet } from 'svelte';
	import type { LucideProps } from '@lucide/svelte';

	interface Props {
		icon?: Component<LucideProps>;
		title: string;
		description?: string;
		cta?: Snippet;
	}

	const { icon: Icon, title, description, cta }: Props = $props();
</script>

<div
	class="border-border bg-card flex flex-col items-center gap-3 rounded-md border border-dashed px-6 py-12 text-center"
>
	{#if Icon}
		<div
			class="flex size-12 items-center justify-center rounded-xl bg-indigo-50 text-indigo-600 dark:bg-indigo-950/50 dark:text-indigo-400"
		>
			<Icon class="size-6" />
		</div>
	{/if}
	<div class="space-y-1">
		<p class="text-sm font-medium">{title}</p>
		{#if description}
			<p class="text-muted-foreground max-w-md text-sm">{description}</p>
		{/if}
	</div>
	{#if cta}
		<div class="mt-2">
			{@render cta()}
		</div>
	{/if}
</div>
