<script lang="ts">
	// Static "about" page. Intentionally light on personal identifiers —
	// who runs this and where to reach them stays vague until a real
	// support / contact channel is published.
</script>

<svelte:head>
	<title>About · Resume Ranker</title>
	<meta name="description" content="What Resume Ranker is and why it exists." />
</svelte:head>

<div class="bg-background min-h-screen">
	<header class="mx-auto flex h-14 max-w-3xl items-center justify-between px-4 sm:px-6">
		<a href="/" class="text-sm font-semibold tracking-tight">Resume Ranker</a>
	</header>

	<main class="mx-auto max-w-2xl px-4 pt-8 pb-24 sm:px-6 sm:pt-12">
		<h1 class="text-2xl font-semibold tracking-tight sm:text-3xl">About Resume Ranker</h1>

		<div class="mt-10 max-w-none space-y-8 text-sm leading-7">
			<section>
				<h2 class="text-base font-semibold">What this is</h2>
				<p class="text-muted-foreground mt-2">
					Resume Ranker compares a resume to a job description and produces
					a section-by-section match analysis with evidence quotes drawn
					from both documents. The goal is to make it easy to see what a
					job posting is actually asking for and whether the resume
					already says it &mdash; without falling back on a single opaque
					"match score."
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">Why no top-line score</h2>
				<p class="text-muted-foreground mt-2">
					A single percentage is satisfying but misleading. A resume can
					score 90% by keyword-overlap and still miss the one or two
					requirements that actually matter. Breaking the analysis down by
					section, with the exact quote each conclusion is based on, lets
					you reach your own judgment instead of trusting a number.
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">How it works</h2>
				<ul class="text-muted-foreground mt-2 list-disc space-y-2 pl-6">
					<li>You upload a resume (PDF) and the job description (paste or PDF).</li>
					<li>
						The text is extracted, and a large language model identifies the
						concrete requirements the JD asks for.
					</li>
					<li>
						For each requirement, the model looks for evidence in the resume
						and quotes the supporting sentence verbatim.
					</li>
					<li>
						The result is a section-by-section view of strengths, partial
						matches, and gaps &mdash; each backed by a quote you can verify.
					</li>
				</ul>
				<p class="text-muted-foreground mt-3">
					Suggestions in the analysis are factual, not prescriptive: "the
					JD asks for X, your resume doesn't mention X." Whether to add X
					&mdash; or whether it's even true that you have X &mdash; is your
					call.
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">Who runs Resume Ranker</h2>
				<p class="text-muted-foreground mt-2">
					Resume Ranker is built by an independent developer experimenting
					with AI-assisted tools for job seekers. It is not affiliated with
					any employer, recruiter, applicant-tracking-system vendor, or
					career-coaching service. It is currently in beta and free to use.
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">What we don't do</h2>
				<ul class="text-muted-foreground mt-2 list-disc space-y-2 pl-6">
					<li>We don't submit your resume to job postings on your behalf.</li>
					<li>We don't share your resume with employers or recruiters.</li>
					<li>We don't sell your data, ever.</li>
					<li>We don't use your uploaded content to train any AI model.</li>
					<li>We don't guarantee any specific job-search outcome.</li>
				</ul>
				<p class="text-muted-foreground mt-3">
					The full data-handling story lives in our
					<a class="text-foreground underline" href="/privacy">Privacy Policy</a>;
					the rules of use are in our
					<a class="text-foreground underline" href="/terms">Terms of Service</a>.
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">Contact</h2>
				<p class="text-muted-foreground mt-2">
					For questions about your data, your account, or these policies,
					use our
					<a class="text-foreground underline" href="/contact">contact form</a>.
					We'll reply to the email you provide.
				</p>
			</section>
		</div>
	</main>
</div>
