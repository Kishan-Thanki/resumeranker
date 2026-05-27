<script lang="ts">
	import { FileText, Sparkles, Quote } from '@lucide/svelte';
	import { Button } from '$lib/components/ui/button';
	import Wordmark from '$lib/components/Wordmark.svelte';
</script>

<div class="bg-background">
	<header class="mx-auto flex h-14 max-w-5xl items-center justify-between px-4 sm:px-6">
		<Wordmark />
	</header>

	<main class="mx-auto max-w-3xl px-4 pt-12 pb-24 sm:px-6 sm:pt-20">
		<div
			class="border-border bg-muted/40 text-muted-foreground mb-5 inline-flex items-center gap-1.5 rounded-full border px-3 py-1 text-xs font-medium"
		>
			<Sparkles class="size-3.5 text-indigo-500" />
			AI-powered, evidence-based
		</div>

		<h1 class="text-3xl font-semibold tracking-tight text-balance sm:text-4xl">
			See exactly how your resume matches a
			<span class="text-indigo-600 dark:text-indigo-400">job description</span>.
		</h1>
		<p class="text-muted-foreground mt-4 max-w-2xl text-base text-pretty sm:text-lg">
			Section-by-section scoring with evidence from your own resume. No magic match
			percentage &mdash; only what the JD asks for and what your resume already says.
		</p>

		<div class="mt-8">
			<Button size="lg" href="/auth">Get started</Button>
		</div>

		<section class="mt-20 grid gap-8 sm:mt-24 sm:grid-cols-3">
			<div>
				<div
					class="flex size-10 items-center justify-center rounded-xl bg-indigo-50 text-indigo-600 dark:bg-indigo-950/50 dark:text-indigo-400"
				>
					<FileText class="size-5" />
				</div>
				<h2 class="mt-4 text-sm font-semibold">Upload</h2>
				<p class="text-muted-foreground mt-1 text-sm">
					Drop in a PDF resume and the job description you're applying to.
				</p>
			</div>
			<div>
				<div
					class="flex size-10 items-center justify-center rounded-xl bg-violet-50 text-violet-600 dark:bg-violet-950/50 dark:text-violet-400"
				>
					<Sparkles class="size-5" />
				</div>
				<h2 class="mt-4 text-sm font-semibold">Analyze</h2>
				<p class="text-muted-foreground mt-1 text-sm">
					Each requirement in the JD is matched against the actual text of your resume.
				</p>
			</div>
			<div>
				<div
					class="flex size-10 items-center justify-center rounded-xl bg-sky-50 text-sky-600 dark:bg-sky-950/50 dark:text-sky-400"
				>
					<Quote class="size-5" />
				</div>
				<h2 class="mt-4 text-sm font-semibold">See evidence</h2>
				<p class="text-muted-foreground mt-1 text-sm">
					Direct quotes &mdash; what the JD asks for, what your resume says, side by side.
				</p>
			</div>
		</section>
	</main>
</div>
