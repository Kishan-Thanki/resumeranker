<script lang="ts">
	import { goto } from '$app/navigation';
	import { FileText, Sparkles } from '@lucide/svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Label } from '$lib/components/ui/label';
	import FileUpload from '$lib/components/FileUpload.svelte';
	import { analyses, create } from '$lib/stores/analyses';
	import { ApiError } from '$lib/api';
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

	// Upload logic
	type JdMode = 'pdf' | 'text';
	let jdMode = $state<JdMode>('text');
	let jdFile = $state<File | null>(null);
	let jdText = $state('');
	let resumeFile = $state<File | null>(null);
	let showResumeTextNote = $state(false);
	let submitting = $state(false);
	let error = $state<string | null>(null);

	const jdReady = $derived(jdMode === 'pdf' ? jdFile !== null : jdText.trim().length >= 20);
	const ready = $derived(jdReady && resumeFile !== null);

	async function analyze() {
		if (!ready || !resumeFile) return;
		submitting = true;
		error = null;
		try {
			const result = await create({
				jdInputType: jdMode,
				jdText: jdMode === 'text' ? jdText : undefined,
				jdPdf: jdMode === 'pdf' ? (jdFile ?? undefined) : undefined,
				resume: resumeFile
			});
			await goto(`/app/analysis/${result.id}`);
		} catch (e) {
			if (e instanceof ApiError) {
				error = e.message || 'Could not create analysis.';
			} else {
				error = 'Could not reach the server.';
			}
		} finally {
			submitting = false;
		}
	}
</script>

<div class="space-y-12">
	<!-- Active Workspace (Upload Area) -->
	<section class="space-y-6">
		<div>
			<h1 class="text-xl font-semibold tracking-tight">New analysis</h1>
			<p class="text-muted-foreground mt-1 text-sm">
				Paste or upload the job description, then attach your resume to compare.
			</p>
		</div>

		<div class="grid gap-6 lg:grid-cols-2">
			<div class="bg-card border-border flex flex-col space-y-4 rounded-xl border p-5 shadow-sm">
				<div class="flex items-center justify-between">
					<Label class="text-base font-semibold">Job description</Label>
					<div class="bg-muted text-muted-foreground inline-flex rounded-md p-0.5 text-xs">
						<button
							type="button"
							class={['rounded-sm px-2.5 py-1 transition-colors', jdMode === 'text' ? 'bg-background text-foreground shadow-sm' : 'hover:text-foreground'].join(' ')}
							onclick={() => (jdMode = 'text')}
						>
							Paste text
						</button>
						<button
							type="button"
							class={['rounded-sm px-2.5 py-1 transition-colors', jdMode === 'pdf' ? 'bg-background text-foreground shadow-sm' : 'hover:text-foreground'].join(' ')}
							onclick={() => (jdMode = 'pdf')}
						>
							Upload PDF
						</button>
					</div>
				</div>

				{#if jdMode === 'text'}
					<div class="flex flex-1 flex-col">
						<Textarea
							bind:value={jdText}
							placeholder="Paste the full job description here. The first line will be used as the title."
							class="min-h-40 flex-1 font-mono text-sm"
						/>
						<p class="text-muted-foreground mt-2 text-right text-xs">
							{jdText.length} character{jdText.length === 1 ? '' : 's'}
						</p>
					</div>
				{:else}
					<div class="flex flex-1 flex-col justify-center">
						<FileUpload
							id="jd-file"
							accept="application/pdf,.pdf"
							acceptLabel="PDF only"
							file={jdFile}
							onChange={(f) => (jdFile = f)}
						/>
					</div>
				{/if}
			</div>

			<div class="bg-card border-border flex flex-col space-y-4 rounded-xl border p-5 shadow-sm">
				<Label class="text-base font-semibold">Resume</Label>
				<div class="flex flex-1 flex-col justify-center gap-4">
					<FileUpload
						id="resume-file"
						accept="application/pdf,.pdf"
						acceptLabel="PDF only"
						file={resumeFile}
						onChange={(f) => (resumeFile = f)}
					/>
					<div class="text-center">
						{#if !showResumeTextNote}
							<button
								type="button"
								class="text-muted-foreground hover:text-foreground focus-visible:ring-ring rounded-sm text-xs underline-offset-4 hover:underline focus-visible:ring-2 focus-visible:outline-hidden"
								onclick={() => (showResumeTextNote = true)}
							>
								Paste instead?
							</button>
						{:else}
							<p class="text-muted-foreground text-xs">
								Resume must be uploaded as PDF for v1.
							</p>
						{/if}
					</div>
				</div>
			</div>
		</div>

		{#if error}
			<p class="text-destructive text-sm font-medium">{error}</p>
		{/if}

		<div class="flex justify-end">
			<Button size="lg" onclick={analyze} disabled={!ready || submitting}>
				<Sparkles class="mr-2 size-4" />
				{submitting ? 'Analyzing...' : 'Analyze'}
			</Button>
		</div>
	</section>

	<!-- History Section -->
	<section class="space-y-4">
		<h2 class="text-lg font-semibold tracking-tight">Recent analyses</h2>
		{#if $analyses.length === 0}
			<div class="border-border bg-card/50 flex flex-col items-center justify-center rounded-xl border border-dashed p-8 text-center sm:p-12">
				<div class="bg-muted text-muted-foreground mb-4 flex size-12 items-center justify-center rounded-full">
					<FileText class="size-6" />
				</div>
				<h3 class="text-base font-medium">No history yet</h3>
				<p class="text-muted-foreground mt-1 max-w-sm text-sm">
					Your past analyses will appear here once you run your first check above.
				</p>
			</div>
		{:else}
			<div class="border-border overflow-hidden rounded-md border shadow-sm">
				<table class="w-full text-sm">
					<thead class="bg-muted/40 text-muted-foreground text-xs uppercase tracking-wide">
						<tr>
							<th class="px-4 py-2.5 text-left font-medium">Date</th>
							<th class="px-4 py-2.5 text-left font-medium">Job description</th>
							<th class="hidden px-4 py-2.5 text-left font-medium sm:table-cell">Resume</th>
							<th class="px-4 py-2.5 text-right font-medium">Status</th>
						</tr>
					</thead>
					<tbody class="divide-border divide-y">
						{#each $analyses as a (a.id)}
							<tr class="hover:bg-muted/30 focus-within:bg-muted/30 relative transition-colors">
								<td class="px-4 py-3 align-top whitespace-nowrap text-muted-foreground">
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
									<span class="truncate block max-w-[200px]">{a.resumeName}</span>
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
	</section>
</div>
