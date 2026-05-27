<script lang="ts">
	// Static page — no data fetch, no auth required.
	// Update the "Last updated" line in tandem with any substantive content
	// change. Bump CURRENT_POLICY_VERSION in src/routes/auth/+page.svelte and
	// in backend `app/config.py:current_policy_version` whenever this date
	// changes, so existing users are prompted to re-accept.
	const lastUpdated = '2026-05-19';
</script>

<svelte:head>
	<title>Terms · Resume Ranker</title>
	<meta name="description" content="The terms under which you may use Resume Ranker." />
</svelte:head>

<div class="bg-background min-h-screen">
	<header class="mx-auto flex h-14 max-w-3xl items-center justify-between px-4 sm:px-6">
		<a href="/" class="text-sm font-semibold tracking-tight">Resume Ranker</a>
	</header>

	<main class="mx-auto max-w-2xl px-4 pt-8 pb-24 sm:px-6 sm:pt-12">
		<h1 class="text-2xl font-semibold tracking-tight sm:text-3xl">Terms of Service</h1>
		<p class="text-muted-foreground mt-2 text-sm">Last updated {lastUpdated}</p>

		<div class="prose prose-zinc dark:prose-invert mt-10 max-w-none space-y-8 text-sm leading-7">
			<section>
				<h2 class="text-base font-semibold">Summary</h2>
				<p class="text-muted-foreground mt-2">
					Resume Ranker is a free tool that compares a resume to a job description
					and returns a section-by-section match analysis. These terms describe
					the rules for using the service and the limits of our responsibility.
					Plain English; please read them. By signing in, you agree to all of
					what follows.
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">Acceptance of these terms</h2>
				<p class="text-muted-foreground mt-2">
					When you check the "I agree" box on the sign-in page and submit your
					email, you accept these terms and our
					<a class="text-foreground underline" href="/privacy">Privacy Policy</a>.
					We record which version you accepted and when. If you don't agree to
					these terms, do not sign in.
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">The service</h2>
				<p class="text-muted-foreground mt-2">
					Resume Ranker accepts a resume and a job description, then produces a
					section-by-section analysis with evidence quotes from both documents.
					The service is free, currently in beta, and provided "as is" without any
					warranty (see <em>Disclaimers</em> below). We may pause, change, or
					discontinue the service at any time.
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">Eligibility</h2>
				<p class="text-muted-foreground mt-2">
					You must be at least 13 years old (or the minimum digital-consent age
					in your jurisdiction, whichever is higher) to use Resume Ranker. You
					must have the legal capacity to enter into a binding agreement. If
					you're using the service on behalf of an organization, you confirm you
					have authority to bind that organization to these terms.
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">Your account</h2>
				<p class="text-muted-foreground mt-2">
					Sign-in is by magic-link to your email address. You are responsible for
					keeping access to that inbox secure. You may have one account per
					email address. Activity that happens through your signed-in session is
					attributed to you.
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">Acceptable use</h2>
				<p class="text-muted-foreground mt-2">When using Resume Ranker, you agree NOT to:</p>
				<ul class="text-muted-foreground mt-2 list-disc space-y-1 pl-6">
					<li>Submit content you don't own or have permission to upload (including someone else's resume without their consent).</li>
					<li>Use the service to harass, defame, or violate the rights of any third party.</li>
					<li>Attempt to access accounts, data, or features that are not yours.</li>
					<li>Scrape, crawl, or use automated systems to interact with the service beyond what the user interface permits.</li>
					<li>Reverse-engineer, decompile, or attempt to derive source code from any part of the service.</li>
					<li>Use the service to build a competing product or to train another machine-learning model.</li>
					<li>Submit content that is illegal, infringing, malicious, deceptive, or designed to harm us, other users, or third parties.</li>
					<li>Interfere with the service's normal operation (denial-of-service attempts, abuse of rate limits, etc.).</li>
				</ul>
				<p class="text-muted-foreground mt-3">
					We may suspend or terminate accounts that violate these rules without
					prior notice.
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">Your content</h2>
				<p class="text-muted-foreground mt-2">
					You retain ownership of everything you upload — your resume text, the
					job description, and any analyses you generate.
				</p>
				<p class="text-muted-foreground mt-3">
					You warrant that:
				</p>
				<ul class="text-muted-foreground mt-2 list-disc space-y-1 pl-6">
					<li>You own the resume content you upload, or you have explicit permission from its owner to upload and analyze it here.</li>
					<li>You have the right to share the job-description text you upload (most public job postings qualify).</li>
					<li>None of your uploaded content infringes someone else's intellectual property or privacy rights.</li>
				</ul>
				<p class="text-muted-foreground mt-3">
					You grant us a limited, non-exclusive, royalty-free license to store
					and process your content solely for the purpose of providing you the
					analysis. <strong class="text-foreground">We do not use your content
					to train any machine-learning model</strong>, and we do not sell,
					rent, or share it with parties other than the service providers
					described in our Privacy Policy.
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">AI output disclaimer</h2>
				<p class="text-muted-foreground mt-2">
					Resume Ranker uses a large language model to generate the analysis.
					Output from any large language model can be inaccurate, incomplete,
					out of date, biased, or otherwise unsuitable for your purpose. You
					should treat the analysis as a starting point for your own thinking,
					not as authoritative advice. The output is not a substitute for
					professional career counselling, recruiting expertise, legal advice,
					or any other licensed professional service.
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">No hiring guarantee</h2>
				<p class="text-muted-foreground mt-2">
					Using Resume Ranker does not guarantee any outcome — not interviews,
					not job offers, not improved match rates with employer tracking
					systems, not any specific employer's attention. The job market
					depends on many factors outside our service. We make no representation
					that following any suggestion in an analysis will improve your
					chances of being hired.
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">Intellectual property</h2>
				<p class="text-muted-foreground mt-2">
					The Resume Ranker brand, website, software, models, and analysis
					methodology are owned by us (or our licensors) and protected by
					copyright, trademark, and other laws. These terms grant you no
					rights in those, beyond the limited right to use the service
					according to these terms.
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">Privacy</h2>
				<p class="text-muted-foreground mt-2">
					Our handling of your personal data is described in our
					<a class="text-foreground underline" href="/privacy">Privacy Policy</a>,
					which is incorporated into these terms by reference. Acceptance of
					these terms includes acceptance of the Privacy Policy.
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">Service availability</h2>
				<p class="text-muted-foreground mt-2">
					Resume Ranker is a best-effort, free service with no service-level
					agreement. We may experience outages, planned maintenance, or
					capacity limits. We may rate-limit, queue, or temporarily refuse
					requests when our usage caps with third-party providers are reached.
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">Termination</h2>
				<p class="text-muted-foreground mt-2">
					You can stop using the service and delete your account at any time
					through the in-app account controls (or by contacting us). Deletion
					removes your stored content from our active systems within a
					reasonable period.
				</p>
				<p class="text-muted-foreground mt-3">
					We may suspend or terminate your access if you violate these terms,
					if your use threatens the service or other users, or if continuing
					to provide the service becomes legally or operationally impractical.
					Where reasonable, we will give you notice.
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">Disclaimers</h2>
				<p class="text-muted-foreground mt-2">
					The service is provided "AS IS" and "AS AVAILABLE," without warranties
					of any kind, express or implied, including the implied warranties of
					merchantability, fitness for a particular purpose, accuracy, and
					non-infringement. We do not warrant that the service will be
					uninterrupted, error-free, secure, or that the analyses will be
					accurate or useful for your purpose.
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">Limitation of liability</h2>
				<p class="text-muted-foreground mt-2">
					To the maximum extent permitted by applicable law, our total
					liability for any claim arising out of or relating to the service
					or these terms is limited to the greater of (a) the amount you have
					paid us in the 12 months before the claim arose (currently $0, as
					the service is free) or (b) USD $50.
				</p>
				<p class="text-muted-foreground mt-3">
					In no event shall we be liable for indirect, incidental, special,
					consequential, exemplary, or punitive damages — including but not
					limited to lost profits, lost data, lost opportunities, or
					reputational harm — even if we have been advised of the possibility
					of such damages.
				</p>
				<p class="text-muted-foreground mt-3">
					Some jurisdictions do not allow exclusion of certain warranties or
					limitations on certain damages; in those jurisdictions our liability
					is limited to the smallest extent permitted by law.
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">Indemnification</h2>
				<p class="text-muted-foreground mt-2">
					You agree to defend and indemnify us against any third-party claim
					arising out of (a) content you upload that violates someone else's
					rights, (b) your violation of these terms, or (c) your violation of
					any law in connection with your use of the service.
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">Changes to these terms</h2>
				<p class="text-muted-foreground mt-2">
					We may update these terms from time to time. If we make material
					changes, we will update the version date at the top of this page
					and, where practical, prompt you to re-accept the new version on
					your next sign-in. Continuing to use the service after such a
					prompt means you accept the updated terms.
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">Governing law and disputes</h2>
				<p class="text-muted-foreground mt-2">
					These terms are governed by the laws of India, without regard to
					conflict-of-law principles. Any dispute arising out of or relating
					to these terms or the service will be resolved in the courts located
					in India. (If you operate from a different jurisdiction and want a
					different governing law, please contact us before signing in.)
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">Contact</h2>
				<p class="text-muted-foreground mt-2">
					For questions about these terms or to make a request that can't
					be handled through the in-app controls, use our
					<a class="text-foreground underline" href="/contact">contact form</a>.
				</p>
			</section>

			<section>
				<h2 class="text-base font-semibold">Severability</h2>
				<p class="text-muted-foreground mt-2">
					If any provision of these terms is held unenforceable, that provision
					will be reformed to the minimum extent necessary to make it
					enforceable, and the remaining provisions will continue in full force.
				</p>
			</section>
		</div>
	</main>
</div>
