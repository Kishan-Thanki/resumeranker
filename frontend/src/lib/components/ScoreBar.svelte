<script lang="ts">
	interface Props {
		label: string;
		score: number;
	}

	const { label, score }: Props = $props();
	const clamped = $derived(Math.max(0, Math.min(100, Math.round(score))));

	// Color the fill by score, reusing the same semantic palette as
	// MatchBadge (emerald / amber / red) so the whole results page reads
	// as one coherent traffic-light system.
	const fill = $derived.by(() => {
		if (clamped >= 67) return 'bg-emerald-500';
		if (clamped >= 34) return 'bg-amber-500';
		return 'bg-red-500';
	});
	const text = $derived.by(() => {
		if (clamped >= 67) return 'text-emerald-700 dark:text-emerald-400';
		if (clamped >= 34) return 'text-amber-700 dark:text-amber-400';
		return 'text-red-700 dark:text-red-400';
	});
</script>

<div>
	<div class="mb-2 flex items-baseline justify-between gap-3">
		<span class="text-sm font-medium">{label}</span>
		<span class="font-mono text-sm font-semibold tabular-nums {text}">{clamped}%</span>
	</div>
	<div
		class="bg-muted h-2 w-full overflow-hidden rounded-full"
		role="progressbar"
		aria-valuenow={clamped}
		aria-valuemin={0}
		aria-valuemax={100}
		aria-label={`${label} score: ${clamped} out of 100`}
	>
		<div
			class="{fill} h-full rounded-full transition-[width] duration-500 ease-out"
			style="width: {clamped}%"
		></div>
	</div>
</div>
