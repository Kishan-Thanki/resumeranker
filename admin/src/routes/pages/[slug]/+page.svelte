<script lang="ts">
	import { enhance } from '$app/forms';
	let { data, form } = $props();
</script>

<div class="mb-8 flex items-center justify-between">
	<div>
		<a href="/" class="text-sm text-indigo-600 hover:underline mb-2 inline-block">&larr; Back to Pages</a>
		<h2 class="text-2xl font-bold tracking-tight">Edit: {data.page.title}</h2>
	</div>
</div>

{#if form?.success}
	<div class="mb-6 p-4 bg-emerald-50 text-emerald-700 rounded-md border border-emerald-200 text-sm font-medium">
		Content updated successfully!
	</div>
{/if}
{#if form?.message}
	<div class="mb-6 p-4 bg-destructive/10 text-destructive rounded-md border border-destructive/20 text-sm font-medium">
		{form.message}
	</div>
{/if}

<form method="POST" action="?/update" use:enhance class="space-y-8 max-w-3xl">
	<div class="bg-card border-border rounded-xl border shadow-sm p-6 space-y-6">
		<h3 class="text-lg font-semibold border-b border-border pb-2">SEO & Metadata</h3>
		
		<div class="space-y-2">
			<label for="title" class="text-sm font-medium block">Page Title</label>
			<input type="text" id="title" name="title" value={data.page.title} class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50" />
		</div>
		
		<div class="space-y-2">
			<label for="metaDescription" class="text-sm font-medium block">Meta Description</label>
			<textarea id="metaDescription" name="metaDescription" rows="3" class="flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50">{data.page.metaDescription}</textarea>
		</div>
	</div>

	{#if data.blocks.length > 0}
		<div class="bg-card border-border rounded-xl border shadow-sm p-6 space-y-6">
			<h3 class="text-lg font-semibold border-b border-border pb-2">Content Blocks</h3>
			
			{#each data.blocks as block}
				<div class="space-y-2">
					<label for="block_{block.blockKey}" class="text-sm font-medium block capitalize">{block.blockKey.replace(/_/g, ' ')}</label>
					<textarea id="block_{block.blockKey}" name="block_{block.blockKey}" rows="5" class="flex min-h-[120px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50">{block.contentValue}</textarea>
				</div>
			{/each}
		</div>
	{/if}

	<div class="flex justify-end">
		<button type="submit" class="inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 bg-primary text-primary-foreground hover:bg-primary/90 h-10 px-6 py-2">
			Save Changes
		</button>
	</div>
</form>
