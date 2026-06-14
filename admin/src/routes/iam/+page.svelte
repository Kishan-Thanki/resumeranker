<script lang="ts">
	import { onMount } from 'svelte';
	import { api, type MeResponse } from '$lib/api';
	import { currentUser } from '$lib/stores/auth';

	let users = $state<MeResponse[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	onMount(async () => {
		try {
			users = await api.listUsers();
		} catch (e: any) {
			error = e.message || 'Failed to load users';
		} finally {
			loading = false;
		}
	});

	async function handleRoleChange(userId: string, newRole: 'user' | 'admin' | 'superadmin') {
		const oldUsers = [...users];
		try {
			// Optimistic UI update
			users = users.map((u) => (u.id === userId ? { ...u, role: newRole } : u));
			await api.updateUserRole(userId, newRole);
		} catch (e: any) {
			error = e.message || 'Failed to update role';
			users = oldUsers; // Revert
		}
	}
</script>

<div class="flex flex-col gap-6">
	<div>
		<h2 class="text-2xl font-bold tracking-tight">Identity and Access Management</h2>
		<p class="text-muted-foreground">Manage users and their roles across all services.</p>
	</div>

	{#if error}
		<div class="rounded-md bg-red-50 p-4 text-sm text-red-700">
			{error}
		</div>
	{/if}

	{#if loading}
		<p class="text-slate-500 animate-pulse text-sm">Loading users...</p>
	{:else}
		<div class="rounded-md border bg-card">
			<table class="w-full text-sm text-left">
				<thead class="border-b bg-muted/50">
					<tr>
						<th class="h-10 px-4 font-medium text-muted-foreground">ID</th>
						<th class="h-10 px-4 font-medium text-muted-foreground">Email</th>
						<th class="h-10 px-4 font-medium text-muted-foreground">Role</th>
						{#if $currentUser?.role === 'superadmin'}
							<th class="h-10 px-4 font-medium text-muted-foreground text-right">Actions</th>
						{/if}
					</tr>
				</thead>
				<tbody>
					{#each users as u (u.id)}
						<tr class="border-b last:border-0 hover:bg-muted/50">
							<td class="p-4 font-mono text-xs text-slate-500">{u.id}</td>
							<td class="p-4 font-medium">{u.email}</td>
							<td class="p-4">
								<span
									class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-semibold
                                    {u.role === 'superadmin'
										? 'bg-purple-100 text-purple-800'
										: u.role === 'admin'
											? 'bg-blue-100 text-blue-800'
											: 'bg-slate-100 text-slate-800'}"
								>
									{u.role}
								</span>
							</td>
							{#if $currentUser?.role === 'superadmin'}
								<td class="p-4 text-right">
									<select
										class="h-8 rounded-md border border-slate-300 bg-transparent px-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-400 disabled:opacity-50"
										value={u.role}
										disabled={u.id === $currentUser.id}
										onchange={(e) => handleRoleChange(u.id, e.currentTarget.value as any)}
									>
										<option value="user">User</option>
										<option value="admin">Admin</option>
										<option value="superadmin">Superadmin</option>
									</select>
								</td>
							{/if}
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
