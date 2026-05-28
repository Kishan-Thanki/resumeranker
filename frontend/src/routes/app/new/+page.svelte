<script lang="ts">
	import { goto } from '$app/navigation';
	import { ArrowLeft } from '@lucide/svelte';
	import { Button } from '$lib/components/ui/button';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Label } from '$lib/components/ui/label';
	import { Separator } from '$lib/components/ui/separator';
	import FileUpload from '$lib/components/FileUpload.svelte';
	import { create } from '$lib/stores/analyses';
	import { ApiError } from '$lib/api';

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

<div class="mx-auto max-w-5xl space-y-8">
	<a
		href="/app"
		class="text-muted-foreground hover:text-foreground inline-flex items-center gap-1 text-sm"
	>
		<ArrowLeft class="size-3.5" />
		Back to analyses
	</a>

	<div>
		<h1 class="text-xl font-semibold tracking-tight">New analysis</h1>
		<p class="text-muted-foreground mt-1 text-sm">
			Paste or upload the job description, then attach the resume to compare.
		</p>
	</div>

	<div class="grid gap-8 lg:grid-cols-2 lg:gap-12">
		<section class="bg-card border-border flex flex-col space-y-4 rounded-xl border p-6 shadow-sm">
			<div class="flex items-center justify-between">
				<Label class="text-base font-semibold">Job description</Label>
				<div class="bg-muted text-muted-foreground inline-flex rounded-md p-0.5 text-xs">
					<button
						type="button"
						class={[
							'rounded-sm px-2.5 py-1 transition-colors',
							jdMode === 'text' ? 'bg-background text-foreground shadow-sm' : 'hover:text-foreground'
						].join(' ')}
						onclick={() => (jdMode = 'text')}
					>
						Paste text
					</button>
					<button
						type="button"
						class={[
							'rounded-sm px-2.5 py-1 transition-colors',
							jdMode === 'pdf' ? 'bg-background text-foreground shadow-sm' : 'hover:text-foreground'
						].join(' ')}
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
						class="min-h-48 flex-1 font-mono text-sm"
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
		</section>

		<section class="bg-card border-border flex flex-col space-y-4 rounded-xl border p-6 shadow-sm">
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
							Resume must be uploaded as PDF for v1. Plain text input may be added later.
						</p>
					{/if}
				</div>
			</div>
		</section>
	</div>

	{#if error}
		<p class="text-destructive text-sm">{error}</p>
	{/if}

	<div class="flex justify-end">
		<Button onclick={analyze} disabled={!ready || submitting}>
			{submitting ? 'Uploading...' : 'Analyze'}
		</Button>
	</div>
</div>
