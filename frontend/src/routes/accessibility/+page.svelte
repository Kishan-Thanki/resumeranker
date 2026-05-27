<script lang="ts">
	// Static page. Update lastUpdated whenever accessibility-related changes
	// ship (new audit pass, fixed issue, known limitation resolved).
	const lastUpdated = '2026-05-19';
</script>

<svelte:head>
	<title>Accessibility · Resume Ranker</title>
	<meta name="description" content="Resume Ranker's accessibility commitment and known limitations." />
</svelte:head>

<div class="bg-background min-h-screen">
	<header class="mx-auto flex h-14 max-w-3xl items-center justify-between px-4 sm:px-6">
		<a href="/" class="text-sm font-semibold tracking-tight">Resume Ranker</a>
	</header>

	<main class="mx-auto max-w-2xl px-4 pt-8 pb-24 sm:px-6 sm:pt-12">
		<h1 class="text-2xl font-semibold tracking-tight sm:text-3xl">Accessibility statement</h1>
		<p class="text-muted-foreground mt-2 text-sm">Last updated {lastUpdated}</p>

		<div class="mt-10 max-w-none space-y-8 text-sm leading-7">
			<section>
				<h2 class="text-base font-semibold">Our commitment</h2>
				<p class="text-muted-foreground mt-2">
					We want Resume Ranker to be usable by as many people as possible,
					including users who navigate the web with assistive technology
					(screen readers, voice control, keyboard-only navigation) or who
					have visual, motor, cognitive, or hearing differences.
				</p>
				<p class="text-muted-foreground mt-3">
					We target conformance with
					<a
						class="text-foreground underline"
						href="https://www.w3.org/WAI/WCAG21/quickref/?currentsidebar=%23col_overview&levels=aaa"
						target="_blank"
						rel="noreferrer noopener">WCAG 2.1 Level AA</a
					>. This is the practical standard most legal frameworks
					(EU EAA, US ADA, India RPwD Act) point to.
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">What works well</h2>
				<ul class="text-muted-foreground mt-2 list-disc space-y-2 pl-6">
					<li>
						<strong class="text-foreground">Keyboard navigation.</strong> Every
						interactive element on every page is reachable and activatable
						with keyboard alone. Tab order follows visual flow.
					</li>
					<li>
						<strong class="text-foreground">Visible focus.</strong> All focusable
						elements show a focus ring when keyboard-focused.
					</li>
					<li>
						<strong class="text-foreground">Semantic HTML.</strong> Real
						<code>&lt;button&gt;</code>, <code>&lt;a&gt;</code>,
						<code>&lt;label&gt;</code>, and <code>&lt;h1&gt;–&lt;h6&gt;</code>
						elements throughout. No <code>&lt;div&gt;</code>s pretending to
						be buttons.
					</li>
					<li>
						<strong class="text-foreground">Form labels.</strong> Every
						input has an associated <code>&lt;label&gt;</code> or
						<code>aria-label</code>.
					</li>
					<li>
						<strong class="text-foreground">Dark mode.</strong> System-preference
						aware via <code>mode-watcher</code>; both themes meet WCAG AA
						contrast for body text.
					</li>
					<li>
						<strong class="text-foreground">Responsive at 380px.</strong>
						Every page is designed to render usably on small mobile viewports
						without horizontal scrolling.
					</li>
					<li>
						<strong class="text-foreground">No motion-required interactions.</strong>
						We don't use auto-playing video or motion-triggered UI.
					</li>
				</ul>
			</section>

			<section>
				<h2 class="text-base font-semibold">Known limitations</h2>
				<p class="text-muted-foreground mt-2">
					We try to be honest about gaps. The following are areas where we
					know our coverage is incomplete:
				</p>
				<ul class="text-muted-foreground mt-2 list-disc space-y-2 pl-6">
					<li>
						<strong class="text-foreground">Screen-reader testing on
						analysis output.</strong> The section-by-section analysis page
						uses nested cards and evidence quotes. We have not yet completed
						a structured screen-reader audit of this view. Heading levels
						and ARIA landmarks may need refinement.
					</li>
					<li>
						<strong class="text-foreground">PDF upload UX for assistive
						tech.</strong> The file-upload control is a styled native
						<code>&lt;input type="file"&gt;</code>; it works with screen
						readers but the styled drag-and-drop affordances are visual-only.
					</li>
					<li>
						<strong class="text-foreground">Color as the sole indicator</strong>
						in some places — match-strength badges convey meaning primarily
						through color. Each badge has a text label alongside, but we may
						strengthen the visual signal with icons in a future pass.
					</li>
					<li>
						<strong class="text-foreground">No formal third-party audit yet.</strong>
						This statement reflects developer self-testing, not a certified
						external accessibility audit. Treat it accordingly.
					</li>
				</ul>
			</section>

			<section>
				<h2 class="text-base font-semibold">Browsers and assistive technology we test against</h2>
				<p class="text-muted-foreground mt-2">
					Development testing happens on:
				</p>
				<ul class="text-muted-foreground mt-2 list-disc space-y-2 pl-6">
					<li>Latest Chrome, Safari, and Firefox on macOS.</li>
					<li>Safari on iOS (mobile testing).</li>
					<li>Keyboard-only navigation across all flows.</li>
					<li>
						Limited spot-checks with VoiceOver on macOS / iOS. Full audits
						with NVDA (Windows) and TalkBack (Android) are not yet routine.
					</li>
				</ul>
			</section>

			<section>
				<h2 class="text-base font-semibold">Reporting an accessibility issue</h2>
				<p class="text-muted-foreground mt-2">
					If you encounter an accessibility barrier on Resume Ranker, please
					tell us. We treat accessibility issues with the same priority as
					functional bugs. Use our
					<a class="text-foreground underline" href="/contact">contact form</a>
					to report one.
				</p>
				<p class="text-muted-foreground mt-3">
					When reporting, please include: the page or feature, the assistive
					technology and browser you were using, what you tried to do, and
					what happened (or didn't).
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">Continued improvement</h2>
				<p class="text-muted-foreground mt-2">
					Accessibility isn't a checkbox — it's ongoing work. We update this
					page when meaningful gaps are closed (or new ones discovered). The
					"Last updated" date at the top of this page reflects the most
					recent material change.
				</p>
			</section>
		</div>
	</main>
</div>
