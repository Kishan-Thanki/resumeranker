<script lang="ts">
	let { data } = $props();
</script>

<div class="mb-8">
	<h2 class="text-2xl font-bold tracking-tight">Pages</h2>
	<p class="text-muted-foreground text-sm mt-1">Manage marketing and legal page content.</p>
</div>

<div class="bg-card border-border rounded-xl border shadow-sm">
	<div class="overflow-x-auto">
		<table class="w-full text-sm text-left">
			<thead class="text-muted-foreground bg-secondary/50 text-xs uppercase border-b border-border">
				<tr>
					<th class="px-6 py-4 font-medium">Page</th>
					<th class="px-6 py-4 font-medium">Slug</th>
					<th class="px-6 py-4 font-medium">Actions</th>
				</tr>
			</thead>
			<tbody class="divide-y divide-border">
				{#each data.pages as page}
					<tr class="hover:bg-secondary/20 transition-colors">
						<td class="px-6 py-4 font-medium text-foreground">{page.title}</td>
						<td class="px-6 py-4 text-muted-foreground">/{page.slug === 'landing' ? '' : page.slug}</td>
						<td class="px-6 py-4">
							<a href="/pages/{page.slug}" class="text-indigo-600 hover:text-indigo-700 font-medium">Edit Content</a>
						</td>
					</tr>
				{:else}
					<tr>
						<td colspan="3" class="px-6 py-12 text-center text-muted-foreground">
							No pages created yet.
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
</div>

