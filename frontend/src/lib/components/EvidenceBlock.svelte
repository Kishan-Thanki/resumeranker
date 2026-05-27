<script lang="ts">
	import type { Evidence } from '$lib/types';

	interface Props {
		evidence: Evidence;
	}

	const { evidence }: Props = $props();
	const sourceLabel = $derived(evidence.source === 'jd' ? 'Job description' : 'Resume');
</script>

<figure class="bg-muted/40 border-border rounded-md border p-3">
	<blockquote class="font-mono text-sm leading-relaxed">
		{evidence.text}
	</blockquote>
	<figcaption class="text-muted-foreground mt-2 text-xs">
		{sourceLabel}{#if evidence.location}
			<span aria-hidden="true"> · </span>{evidence.location}{/if}
	</figcaption>
</figure>
