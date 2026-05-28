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
			return 'border-emerald-200 bg-emerald-50 text-emerald-800 dark:border-emerald-900/50 dark:bg-emerald-950/40 dark:text-emerald-300';
		}
		if (s === 'failed') {
			return 'border-red-200 bg-red-50 text-red-800 dark:border-red-900/50 dark:bg-red-950/40 dark:text-red-300';
		}
		return 'border-indigo-200 bg-indigo-50 text-indigo-800 dark:border-indigo-900/50 dark:bg-indigo-950/40 dark:text-indigo-300';
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
		<div class="border-border bg-card flex flex-col items-center justify-center rounded-xl border p-8 text-center sm:p-12">
			<div class="bg-muted text-muted-foreground mb-4 flex size-12 items-center justify-center rounded-full">
				<FileText class="size-6" />
			</div>
			<h2 class="text-xl font-semibold tracking-tight">No analyses yet</h2>
			<p class="text-muted-foreground mt-2 max-w-sm text-sm">
				Upload a resume and a job description to see how they line up, requirement by requirement.
			</p>
			
			<div class="mt-8 grid w-full max-w-2xl grid-cols-1 gap-4 sm:grid-cols-3">
				<div class="bg-background border-border flex flex-col items-center rounded-lg border p-4 text-center shadow-sm">
					<span class="bg-indigo-500/10 text-indigo-500 mb-2 flex size-6 items-center justify-center rounded-full text-xs font-bold">1</span>
					<span class="text-sm font-medium">Upload JD</span>
					<span class="text-muted-foreground mt-1 text-xs">Paste or upload the job description you want.</span>
				</div>
				<div class="bg-background border-border flex flex-col items-center rounded-lg border p-4 text-center shadow-sm">
					<span class="bg-indigo-500/10 text-indigo-500 mb-2 flex size-6 items-center justify-center rounded-full text-xs font-bold">2</span>
					<span class="text-sm font-medium">Upload Resume</span>
					<span class="text-muted-foreground mt-1 text-xs">Attach your PDF resume to compare.</span>
				</div>
				<div class="bg-background border-border flex flex-col items-center rounded-lg border p-4 text-center shadow-sm">
					<span class="bg-indigo-500/10 text-indigo-500 mb-2 flex size-6 items-center justify-center rounded-full text-xs font-bold">3</span>
					<span class="text-sm font-medium">Get Action Plan</span>
					<span class="text-muted-foreground mt-1 text-xs">Review gaps and tailor your resume.</span>
				</div>
			</div>

			<Button href="/app/new" class="mt-8">
				<Plus class="mr-2 size-4" />
				Start your first analysis
			</Button>
		</div>
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
									class="focus-visible:ring-ring rounded-sm font-medium transition-colors before:absolute before:inset-0 before:content-[''] hover:text-indigo-600 focus-visible:ring-2 focus-visible:outline-hidden dark:hover:text-indigo-400"
								>
									{a.jdTitle}
								</a>
							</td>
							<td class="text-muted-foreground hidden px-4 py-3 align-top sm:table-cell">
								<span class="truncate">{a.resumeName}</span>
							</td>
							<td class="px-4 py-3 text-right align-top">
								<Badge variant="outline" class={statusClasses(a.status) + ' inline-flex items-center gap-1.5'}>
									<span class="size-1.5 rounded-full {a.status === 'completed' ? 'bg-emerald-500' : a.status === 'failed' ? 'bg-red-500' : 'bg-indigo-500 animate-pulse'}"></span>
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
