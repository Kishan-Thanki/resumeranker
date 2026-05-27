<script lang="ts">
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button';
	import { isAuthed } from '$lib/stores/auth';

	const home = $derived($isAuthed ? '/app' : '/');
	const title = $derived(page.status === 404 ? 'Page not found' : 'Something went wrong');
	const detail = $derived(
		page.status === 404
			? "The page you're looking for doesn't exist."
			: (page.error?.message ?? 'An unexpected error occurred.')
	);
</script>

<div class="bg-background flex min-h-screen items-center justify-center px-4 py-12">
	<div class="w-full max-w-md space-y-4 text-center">
		<p class="text-muted-foreground font-mono text-xs tracking-wider">
			HTTP {page.status}
		</p>
		<h1 class="text-2xl font-semibold tracking-tight">{title}</h1>
		<p class="text-muted-foreground text-sm">{detail}</p>
		<div class="pt-2">
			<Button href={home} variant="outline">Back to {$isAuthed ? 'analyses' : 'home'}</Button>
		</div>
	</div>
</div>
