<script lang="ts">
	import { Badge } from '$lib/components/ui/badge';
	import type { MatchStrength } from '$lib/types';

	interface Props {
		strength: MatchStrength;
		matched?: boolean;
	}

	const { strength, matched = true }: Props = $props();

	const label = $derived.by(() => {
		if (!matched || strength === 'none') return 'Missing';
		if (strength === 'strong') return 'Strong';
		if (strength === 'partial') return 'Partial';
		return 'Weak';
	});

	const classes = $derived.by(() => {
		if (!matched || strength === 'none') {
			return 'border-border bg-muted text-muted-foreground';
		}
		if (strength === 'strong') {
			return 'border-emerald-200 bg-emerald-50 text-emerald-800 dark:border-emerald-900 dark:bg-emerald-950/40 dark:text-emerald-300';
		}
		if (strength === 'partial') {
			return 'border-amber-200 bg-amber-50 text-amber-800 dark:border-amber-900 dark:bg-amber-950/40 dark:text-amber-300';
		}
		return 'border-red-200 bg-red-50 text-red-800 dark:border-red-900 dark:bg-red-950/40 dark:text-red-300';
	});
</script>

<Badge variant="outline" class={classes} aria-label={`Match: ${label}`}>
	{label}
</Badge>
