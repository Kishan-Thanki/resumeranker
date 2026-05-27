<script lang="ts">
	import { onDestroy } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { ArrowLeft, Plus, Trash2 } from '@lucide/svelte';
	import { toast } from 'svelte-sonner';
	import { Button } from '$lib/components/ui/button';
	import { Card } from '$lib/components/ui/card';
	import * as Dialog from '$lib/components/ui/dialog';
	import SectionScoreCard from '$lib/components/SectionScoreCard.svelte';
	import { analyses, refreshById, remove } from '$lib/stores/analyses';
	import type { AnalysisResult, RequirementMatch } from '$lib/types';

	const POLL_INTERVAL_MS = 2000;
	const POLL_TIMEOUT_MS = 2 * 60 * 1000; // give up after 2 minutes

	// Param is always present in this route; narrow for TypeScript.
	const id = $derived(page.params.id ?? '');
	const analysis = $derived<AnalysisResult | undefined>($analyses.find((a) => a.id === id));
	const isPending = $derived(
		analysis !== undefined && (analysis.status === 'queued' || analysis.status === 'processing')
	);

	let initialFetchDone = $state(false);
	let initialFetchFailed = $state(false);
	let pollingStalled = $state(false);
	let pollTimer: ReturnType<typeof setTimeout> | null = null;
	let pollStartedAt = 0;

	let deleteDialogOpen = $state(false);
	let deleting = $state(false);

	async function handleDelete() {
		deleting = true;
		try {
			await remove(id);
			deleteDialogOpen = false;
			toast('Analysis deleted', { description: 'Its content has been removed.' });
			await goto('/app');
		} catch {
			toast.error("Couldn't delete analysis", { description: 'Please try again.' });
		} finally {
			deleting = false;
		}
	}

	const dateFmt = new Intl.DateTimeFormat(undefined, {
		year: 'numeric',
		month: 'short',
		day: 'numeric',
		hour: '2-digit',
		minute: '2-digit'
	});

	const unmatched = $derived.by<RequirementMatch[]>(() => {
		if (!analysis || analysis.status !== 'completed') return [];
		return analysis.sections.flatMap((s) => s.requirements.filter((r) => !r.matched));
	});

	function stopPolling() {
		if (pollTimer !== null) {
			clearTimeout(pollTimer);
			pollTimer = null;
		}
	}

	function schedulePoll() {
		stopPolling();
		pollTimer = setTimeout(async () => {
			pollTimer = null;
			try {
				await refreshById(id);
			} catch {
				// transient; will retry on the next tick if still pending
			}
			if (Date.now() - pollStartedAt > POLL_TIMEOUT_MS) {
				pollingStalled = true;
				return;
			}
			// `analysis` is reactive — re-check via the latest store snapshot
			const latest = $analyses.find((a) => a.id === id);
			if (latest && (latest.status === 'queued' || latest.status === 'processing')) {
				schedulePoll();
			}
		}, POLL_INTERVAL_MS);
	}

	// Run once when the id is first known; refresh from backend in case the
	// store cache is stale or this is a direct hit on the URL.
	$effect(() => {
		if (!id || initialFetchDone) return;
		initialFetchDone = true;
		(async () => {
			try {
				await refreshById(id);
			} catch {
				initialFetchFailed = true;
			}
		})();
	});

	// Drive the polling loop only while the analysis is queued/processing.
	$effect(() => {
		if (isPending && pollTimer === null && !pollingStalled) {
			pollStartedAt = Date.now();
			schedulePoll();
		}
		if (!isPending) {
			stopPolling();
		}
	});

	onDestroy(stopPolling);
</script>

<!--
	Back link is rendered before the conditional content so it's available in
	every state (loading, not-found, pending, failed, completed). The TopBar
	gives a route to /app via the brand wordmark, but a labelled "Back to
	analyses" link is clearer for users who don't know the wordmark is a link.
-->
<a
	href="/app"
	class="text-muted-foreground hover:text-foreground mb-4 inline-flex items-center gap-1 text-sm"
>
	<ArrowLeft class="size-3.5" />
	Back to analyses
</a>

{#if !analysis && !initialFetchDone}
	<p class="text-muted-foreground text-sm">Loading...</p>
{:else if !analysis}
	<div class="space-y-4">
		<h1 class="text-xl font-semibold tracking-tight">Analysis not found</h1>
		<p class="text-muted-foreground text-sm">
			{initialFetchFailed
				? "We couldn't reach the server. Try again in a moment."
				: "We couldn't find that analysis. It may have been removed or the link is wrong."}
		</p>
		<Button href="/app" variant="outline">Back to analyses</Button>
	</div>
{:else}
	<div class="space-y-6">
		<header class="flex items-start justify-between gap-3">
			<div class="space-y-1">
				<h1 class="text-xl font-semibold tracking-tight">{analysis.jdTitle}</h1>
				<p class="text-muted-foreground text-sm">
					{analysis.resumeName} · {dateFmt.format(new Date(analysis.createdAt))}
				</p>
			</div>
			<Button
				variant="ghost"
				size="icon"
				class="text-muted-foreground hover:text-destructive shrink-0"
				aria-label="Delete this analysis"
				onclick={() => (deleteDialogOpen = true)}
			>
				<Trash2 class="size-4" />
			</Button>
		</header>

		{#if analysis.status === 'queued' || analysis.status === 'processing'}
			<div class="space-y-3">
				<p class="text-muted-foreground inline-flex items-center gap-2 text-sm">
					<span class="text-current inline-flex items-center gap-1">
						<span class="bg-current size-1.5 animate-pulse rounded-full"></span>
						<span class="bg-current size-1.5 animate-pulse rounded-full [animation-delay:120ms]"
						></span>
						<span class="bg-current size-1.5 animate-pulse rounded-full [animation-delay:240ms]"
						></span>
					</span>
					{analysis.status === 'queued'
						? 'Queued'
						: 'Analyzing your resume against the job description'}
				</p>
				{#if pollingStalled}
					<p class="text-muted-foreground text-xs">
						Still working in the background. Refresh in a moment to check progress.
					</p>
				{/if}
				<div class="space-y-4">
					{#each ['Skills', 'Experience', 'Education', 'Leadership Signals'] as label, i (i)}
						<Card class="animate-pulse">
							<div class="px-6 pt-6">
								<div class="mb-2 flex items-baseline justify-between gap-3">
									<span class="text-sm font-medium">{label}</span>
									<span class="text-muted-foreground text-sm">—</span>
								</div>
								<div class="bg-muted h-1.5 rounded-full"></div>
							</div>
							<div class="border-border mt-4 space-y-3 border-t px-6 py-6">
								<div class="bg-muted h-3 w-3/4 rounded"></div>
								<div class="bg-muted h-3 w-2/3 rounded"></div>
								<div class="bg-muted h-3 w-1/2 rounded"></div>
							</div>
						</Card>
					{/each}
				</div>
			</div>
		{:else if analysis.status === 'failed'}
			<div class="space-y-4">
				<p class="text-sm">
					{analysis.errorMessage ??
						'Something went wrong while analyzing this resume. Nothing was saved.'}
				</p>
				<Button href="/app/new">Try again</Button>
			</div>
		{:else}
			<div class="space-y-4">
				{#each analysis.sections as section (section.id)}
					<SectionScoreCard {section} />
				{/each}
			</div>

			{#if unmatched.length > 0}
				<section class="space-y-3">
					<h2 class="text-base font-semibold">Gaps</h2>
					<p class="text-muted-foreground text-sm">
						Requirements in the job description that this resume doesn't address.
					</p>
					<ul class="border-border space-y-2 rounded-md border p-4 text-sm">
						{#each unmatched as r (r.id)}
							<li class="flex gap-2">
								<span class="text-muted-foreground select-none">—</span>
								<span>
									The JD requires <span class="font-medium">{r.requirement}</span>;
									your resume doesn't mention it.
								</span>
							</li>
						{/each}
					</ul>
				</section>
			{/if}

			<!--
				Footer CTAs so the page isn't a dead end. Primary action is
				"start another analysis" (most common next step after reading
				one). Secondary is "back to all analyses" for users who want
				to browse history.
			-->
			<footer class="border-border flex flex-col gap-3 border-t pt-6 sm:flex-row sm:items-center sm:justify-between">
				<Button href="/app/new">
					<Plus class="size-4" />
					New analysis
				</Button>
				<Button href="/app" variant="outline">
					<ArrowLeft class="size-4" />
					All analyses
				</Button>
			</footer>
		{/if}
	</div>
{/if}

<Dialog.Root bind:open={deleteDialogOpen}>
	<Dialog.Content class="sm:max-w-md">
		<Dialog.Header>
			<Dialog.Title>Delete this analysis?</Dialog.Title>
			<Dialog.Description>
				This permanently removes the analysis and its content — the resume text,
				job-description text, and scores. This cannot be undone.
			</Dialog.Description>
		</Dialog.Header>
		<Dialog.Footer class="gap-2 sm:gap-2">
			<Button
				type="button"
				variant="outline"
				disabled={deleting}
				onclick={() => (deleteDialogOpen = false)}
			>
				Cancel
			</Button>
			<Button type="button" variant="destructive" disabled={deleting} onclick={handleDelete}>
				{deleting ? 'Deleting...' : 'Delete analysis'}
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
