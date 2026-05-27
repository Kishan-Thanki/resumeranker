<!--
	Site-wide footer. Rendered by the root layout (src/routes/+layout.svelte)
	so it appears on every route — marketing pages, auth flow, and the
	signed-in /app shell. The links are intentionally minimal: only the
	pages every visitor should be able to reach (terms, privacy, the
	about/contact context). The TopBar handles in-app navigation; this
	handles "what's this site, where do I read the policies."
-->
<footer class="border-border border-t">
	<div
		class="text-muted-foreground mx-auto flex max-w-5xl flex-col gap-3 px-4 py-6 text-xs sm:flex-row sm:items-center sm:justify-between sm:px-6"
	>
		<span>Resume Ranker</span>
		<nav class="flex flex-wrap items-center gap-x-4 gap-y-2">
			<a class="hover:text-foreground" href="/about">About</a>
			<a class="hover:text-foreground" href="/contact">Contact</a>
			<a class="hover:text-foreground" href="/terms">Terms</a>
			<a class="hover:text-foreground" href="/privacy">Privacy</a>
			<a class="hover:text-foreground" href="/accessibility">Accessibility</a>
		</nav>
	</div>
</footer>
