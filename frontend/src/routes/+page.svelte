<script lang="ts">
	import { FileText, Sparkles, Quote } from '@lucide/svelte';
	import { Button } from '$lib/components/ui/button';
</script>

<div class="bg-background min-h-screen">
	<header class="mx-auto flex h-14 max-w-5xl items-center justify-between px-4 sm:px-6">
		<span class="text-sm font-semibold tracking-tight">Resume Ranker</span>
	</header>

	<main class="mx-auto max-w-3xl px-4 pt-12 pb-24 sm:px-6 sm:pt-20">
		<h1 class="text-3xl font-semibold tracking-tight sm:text-4xl">
			See exactly how your resume matches a job description.
		</h1>
		<p class="text-muted-foreground mt-4 max-w-2xl text-base sm:text-lg">
			Section-by-section scoring with evidence from your own resume. No magic match
			percentage &mdash; only what the JD asks for and what your resume already says.
		</p>

		<div class="mt-8">
			<Button size="lg" href="/auth">Get started</Button>
		</div>

		<section class="mt-20 grid gap-6 sm:mt-24 sm:grid-cols-3">
			<div>
				<FileText class="text-muted-foreground size-5" />
				<h2 class="mt-3 text-sm font-semibold">Upload</h2>
				<p class="text-muted-foreground mt-1 text-sm">
					Drop in a PDF resume and the job description you're applying to.
				</p>
			</div>
			<div>
				<Sparkles class="text-muted-foreground size-5" />
				<h2 class="mt-3 text-sm font-semibold">Analyze</h2>
				<p class="text-muted-foreground mt-1 text-sm">
					Each requirement in the JD is matched against the actual text of your resume.
				</p>
			</div>
			<div>
				<Quote class="text-muted-foreground size-5" />
				<h2 class="mt-3 text-sm font-semibold">See evidence</h2>
				<p class="text-muted-foreground mt-1 text-sm">
					Direct quotes &mdash; what the JD asks for, what your resume says, side by side.
				</p>
			</div>
		</section>
	</main>
</div>
