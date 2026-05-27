<script lang="ts">
	import { Progress } from '$lib/components/ui/progress';

	interface Props {
		label: string;
		score: number;
	}

	const { label, score }: Props = $props();
	const clamped = $derived(Math.max(0, Math.min(100, Math.round(score))));
</script>

<div>
	<div class="mb-2 flex items-baseline justify-between gap-3">
		<span class="text-sm font-medium">{label}</span>
		<span class="text-muted-foreground font-mono text-sm tabular-nums">{clamped}%</span>
	</div>
	<Progress value={clamped} aria-label={`${label} score: ${clamped} out of 100`} />
</div>
