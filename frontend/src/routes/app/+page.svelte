<script lang="ts">
	import { Plus, FileText } from '@lucide/svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import { analyses } from '$lib/stores/analyses';
	import type { AnalysisStatus } from '$lib/types';

	const dateFmt = new Intl.DateTimeFormat(undefined, {
		year: 'numeric',
		month: 'short',
		day: 'numeric'
	});

	function statusLabel(s: AnalysisStatus): string {
		if (s === 'queued') return 'Queued';
		if (s === 'processing') return 'Processing';
		if (s === 'completed') return 'Completed';
		return 'Failed';
	}

	function statusClasses(s: AnalysisStatus): string {
		if (s === 'completed') {
			return 'border-emerald-200 bg-emerald-50 text-emerald-800 dark:border-emerald-900 dark:bg-emerald-950/40 dark:text-emerald-300';
		}
		if (s === 'failed') {
			return 'border-red-200 bg-red-50 text-red-800 dark:border-red-900 dark:bg-red-950/40 dark:text-red-300';
		}
		return 'border-border bg-muted text-muted-foreground';
	}
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between gap-3">
		<h1 class="text-xl font-semibold tracking-tight">Your analyses</h1>
		<Button href="/app/new">
			<Plus class="size-4" />
			New analysis
		</Button>
	</div>

	{#if $analyses.length === 0}
		<EmptyState
			icon={FileText}
			title="No analyses yet"
			description="Upload a resume and a job description to see how they line up, requirement by requirement."
		>
			{#snippet cta()}
				<Button href="/app/new">Start your first analysis</Button>
			{/snippet}
		</EmptyState>
	{:else}
		<div class="border-border overflow-hidden rounded-md border">
			<table class="w-full text-sm">
				<thead class="bg-muted/40 text-muted-foreground text-xs uppercase tracking-wide">
					<tr>
						<th class="px-4 py-2.5 text-left font-medium">Date</th>
						<th class="px-4 py-2.5 text-left font-medium">Job description</th>
						<th class="hidden px-4 py-2.5 text-left font-medium sm:table-cell">Resume</th>
						<th class="px-4 py-2.5 text-right font-medium">Status</th>
					</tr>
				</thead>
				<tbody>
					{#each $analyses as a (a.id)}
						<tr
							class="border-border hover:bg-muted/30 focus-within:bg-muted/30 relative border-t transition-colors"
						>
							<td class="px-4 py-3 align-top whitespace-nowrap">
								{dateFmt.format(new Date(a.createdAt))}
							</td>
							<td class="px-4 py-3 align-top">
								<a
									href={`/app/analysis/${a.id}`}
									aria-label={`Open analysis: ${a.jdTitle}`}
									class="focus-visible:ring-ring rounded-sm before:absolute before:inset-0 before:content-[''] focus-visible:ring-2 focus-visible:outline-hidden"
								>
									{a.jdTitle}
								</a>
							</td>
							<td class="text-muted-foreground hidden px-4 py-3 align-top sm:table-cell">
								<span class="truncate">{a.resumeName}</span>
							</td>
							<td class="px-4 py-3 text-right align-top">
								<Badge variant="outline" class={statusClasses(a.status)}>
									{statusLabel(a.status)}
								</Badge>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
