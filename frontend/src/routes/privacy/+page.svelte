<script lang="ts">
	// Static page — no data fetch, no auth required.
	// Update the "Last updated" line in tandem with substantive changes.
	const lastUpdated = '2026-05-19';
</script>

<svelte:head>
	<title>Privacy · Resume Ranker</title>
	<meta name="description" content="How Resume Ranker collects and handles your data." />
</svelte:head>

<div class="bg-background min-h-screen">
	<header class="mx-auto flex h-14 max-w-3xl items-center justify-between px-4 sm:px-6">
		<a href="/" class="text-sm font-semibold tracking-tight">Resume Ranker</a>
	</header>

	<main class="mx-auto max-w-2xl px-4 pt-8 pb-24 sm:px-6 sm:pt-12">
		<h1 class="text-2xl font-semibold tracking-tight sm:text-3xl">Privacy policy</h1>
		<p class="text-muted-foreground mt-2 text-sm">Last updated {lastUpdated}</p>

		<div class="prose prose-zinc dark:prose-invert mt-10 max-w-none space-y-8 text-sm leading-7">
			<section>
				<h2 class="text-base font-semibold">Summary</h2>
				<p class="text-muted-foreground mt-2">
					Resume Ranker is a tool that compares a resume to a job description and
					returns a section-by-section match analysis. To do that, we need a small
					amount of data from you. This page describes exactly what we collect, why,
					where it goes, and what control you have over it.
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">Data we collect</h2>
				<ul class="text-muted-foreground mt-2 list-disc space-y-2 pl-6">
					<li>
						<strong class="text-foreground">Your email address</strong> &mdash; the only
						identifier for your account. Used to send sign-in links and to scope your
						analyses to you.
					</li>
					<li>
						<strong class="text-foreground">Resume text</strong> &mdash; extracted from
						any PDF you upload. The PDF itself is discarded immediately after
						extraction; only the text is stored.
					</li>
					<li>
						<strong class="text-foreground">Job-description text</strong> &mdash;
						whatever you paste in or upload as PDF. Same handling as resume text.
					</li>
					<li>
						<strong class="text-foreground">Analysis results</strong> &mdash; the
						section scores and evidence quotes our pipeline produces for each
						resume / JD pair.
					</li>
					<li>
						<strong class="text-foreground">Technical metadata</strong> &mdash; IP
						addresses and timestamps tied to authentication requests, kept only for
						rate-limiting and abuse prevention.
					</li>
				</ul>
			</section>

			<section>
				<h2 class="text-base font-semibold">Why we collect it</h2>
				<p class="text-muted-foreground mt-2">
					Strictly to provide the analysis you asked for. We do not sell or share
					your data with advertisers. We do not use your data to train models.
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">Who sees your data</h2>
				<p class="text-muted-foreground mt-2">
					We use a small set of named third-party services ("sub-processors")
					to run Resume Ranker. Each receives only the data it strictly needs
					for its function. They process that data under their own
					privacy terms, which we've reviewed. The full list:
				</p>
				<div class="text-muted-foreground mt-4 overflow-x-auto">
					<table class="w-full border-collapse text-left text-xs">
						<thead>
							<tr class="border-border border-b">
								<th class="text-foreground py-2 pr-4 font-semibold">Provider</th>
								<th class="text-foreground py-2 pr-4 font-semibold">Purpose</th>
								<th class="text-foreground py-2 font-semibold">What it sees</th>
							</tr>
						</thead>
						<tbody>
							<tr class="border-border border-b">
								<td class="py-2 pr-4 align-top">
									<strong class="text-foreground">Resend</strong>
									<a
										class="text-muted-foreground block text-xs underline"
										href="https://resend.com/legal/privacy-policy"
										target="_blank"
										rel="noreferrer noopener">privacy</a
									>
								</td>
								<td class="py-2 pr-4 align-top">Transactional email (sign-in links)</td>
								<td class="py-2 align-top">Your email address; the magic-link URL</td>
							</tr>
							<tr class="border-border border-b">
								<td class="py-2 pr-4 align-top">
									<strong class="text-foreground">AI / LLM provider</strong>
									<span class="text-muted-foreground block text-xs">
										A third-party large-language-model API used to
										generate analyses. Specific provider details are available on request via the contact channel below.
									</span>
								</td>
								<td class="py-2 pr-4 align-top">Generate the analysis from your resume and JD</td>
								<td class="py-2 align-top">
									Your resume text and the JD text. Not your email, not your IP.
									We do not train any model on it, and we request that the
									provider not use it for training where their terms allow.
								</td>
							</tr>
							<tr class="border-border border-b">
								<td class="py-2 pr-4 align-top">
									<strong class="text-foreground">Cloudflare</strong>
									<a
										class="text-muted-foreground block text-xs underline"
										href="https://www.cloudflare.com/privacypolicy/"
										target="_blank"
										rel="noreferrer noopener">privacy</a
									>
								</td>
								<td class="py-2 pr-4 align-top">DNS resolution; future edge protection</td>
								<td class="py-2 align-top">Public DNS queries only (no user content)</td>
							</tr>
							<tr>
								<td class="py-2 pr-4 align-top">
									<strong class="text-foreground">Hosting provider</strong>
									<span class="text-muted-foreground block text-xs">
										Database and application server. Specific provider chosen at
										deploy time; named in the production version of this page.
									</span>
								</td>
								<td class="py-2 pr-4 align-top">Run the database and application processes</td>
								<td class="py-2 align-top">All stored data at rest, behind disk encryption</td>
							</tr>
						</tbody>
					</table>
				</div>
				<p class="text-muted-foreground mt-4">
					No human at Resume Ranker reads your resume or analysis content as
					part of normal operations. Engineers may access logs containing
					your email address (not content) when debugging issues.
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">How long we keep your data</h2>
				<ul class="text-muted-foreground mt-2 list-disc space-y-2 pl-6">
					<li>
						<strong class="text-foreground">Analyses:</strong> kept until you delete
						them or close your account.
					</li>
					<li>
						<strong class="text-foreground">Sessions:</strong> 30-day sliding expiry;
						revoked when you sign out.
					</li>
					<li>
						<strong class="text-foreground">Magic-link tokens:</strong> 15-minute
						expiry, single use; invalidated automatically after first use or
						expiry, and deleted outright when you close your account.
					</li>
					<li>
						<strong class="text-foreground">Rate-limit counters:</strong> kept for the
						length of their window (up to 24 hours), then automatically discarded.
					</li>
				</ul>
			</section>

			<section>
				<h2 class="text-base font-semibold">Your rights</h2>
				<p class="text-muted-foreground mt-2">
					You can sign in to your account at any time to view or delete your
					data. Deleting an analysis removes its content (resume text, JD text,
					scores) from our systems. Closing your account removes your email and
					any remaining analyses associated with it.
				</p>
				<p class="text-muted-foreground mt-3">
					If you're in a jurisdiction with formal data-subject rights (GDPR, UK
					DPA, India DPDP, CCPA, etc.), you also have rights of access,
					correction, deletion, and objection to processing. Most of these are
					exercised by signing in and using the in-app controls; for any request
					that can't be handled in-app, see the Contact section below.
				</p>
				<p class="text-muted-foreground mt-3">
					When you sign in for the first time, you affirmatively accept this
					Privacy Policy and the
					<a class="text-foreground underline" href="/terms">Terms of Service</a>
					by checking a box on the sign-in page. We record which policy version
					you accepted, and the timestamp of acceptance, on your user record.
					This consent log is part of the data we keep about you and is removed
					when you close your account.
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">Security</h2>
				<ul class="text-muted-foreground mt-2 list-disc space-y-2 pl-6">
					<li>
						All traffic between your browser and our servers is encrypted with TLS.
					</li>
					<li>
						Magic-link and session tokens are hashed at rest; we never store the raw
						tokens.
					</li>
					<li>
						Your resume and job-description text are encrypted at the application
						layer before being written to the database, on top of the hosting
						provider's disk encryption.
					</li>
					<li>
						The sign-in session lives in an HttpOnly cookie that JavaScript cannot
						read, reducing the risk of token theft.
					</li>
					<li>
						Per-IP and per-email rate limits are in place on authentication
						endpoints.
					</li>
					<li>
						We never log passwords (there are none &mdash; we use magic-link auth),
						API keys, or the body of your resume or job description.
					</li>
				</ul>
			</section>

			<section>
				<h2 class="text-base font-semibold">Cookies and tracking</h2>
				<p class="text-muted-foreground mt-2">
					We do not use marketing cookies, third-party analytics, or tracking
					pixels. We use a single strictly-necessary cookie to keep you signed
					in: a session cookie set after you verify a magic link. It is
					HttpOnly (not readable by JavaScript), and its only purpose is
					authentication. Because it is essential to providing a service you
					asked for, no consent banner is required for it.
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">Children's privacy</h2>
				<p class="text-muted-foreground mt-2">
					Resume Ranker is not directed at children under 13 (or the equivalent
					age in your jurisdiction). We do not knowingly collect data from
					children.
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">Changes to this policy</h2>
				<p class="text-muted-foreground mt-2">
					If we make material changes to this policy, we'll update the date at the
					top of the page and, where appropriate, notify signed-in users by email.
					Continuing to use the service after such notice constitutes acceptance.
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">Contact</h2>
				<p class="text-muted-foreground mt-2">
					For privacy questions or data-deletion requests that can't be
					handled through the in-app account controls, use our
					<a class="text-foreground underline" href="/contact">contact form</a>.
					We aim to acknowledge privacy requests within 72 hours.
				</p>
			</section>
		</div>
	</main>
</div>
