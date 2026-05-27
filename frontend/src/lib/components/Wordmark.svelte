<script lang="ts">
	// Brand wordmark used in every page header. Plain text — no logo glyph.
	// Pass `href` to render it as a link (varies per page: `/`, `/app`).
	interface Props {
		href?: string;
		class?: string;
	}

	const { href, class: className = '' }: Props = $props();
</script>

{#if href}
	<a {href} class="text-sm font-semibold tracking-tight {className}">Resume Ranker</a>
{:else}
	<span class="text-sm font-semibold tracking-tight {className}">Resume Ranker</span>
{/if}
