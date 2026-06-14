<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { isAuthed, currentUser, refreshAuth } from '$lib/stores/auth';
	import { ModeWatcher } from 'mode-watcher';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';

	let { children } = $props();
    let checking = $state(true);

	onMount(async () => {
		const isAuthPage = page.url.pathname.startsWith('/auth');
		const valid = await refreshAuth();
		checking = false;

		if (!valid && !isAuthPage) {
			goto('/auth');
		} else if (valid) {
			// Enforce role
			const role = $currentUser?.role;
			if (role !== 'admin' && role !== 'superadmin') {
				// Access Denied
				// If they are just a regular user, log them out or show an error
			}
		}
	});

    // Reactive effect for route changes after mount
    $effect(() => {
        const isAuthPage = page.url.pathname.startsWith('/auth');
        if (!checking && !$isAuthed && !isAuthPage) {
            goto('/auth');
        }
    });

    const isAccessDenied = $derived($isAuthed && $currentUser && $currentUser.role !== 'admin' && $currentUser.role !== 'superadmin');
    const showSidebar = $derived($isAuthed && !isAccessDenied && !page.url.pathname.startsWith('/auth'));
</script>

<ModeWatcher />

{#if checking}
    <div class="flex min-h-screen items-center justify-center bg-secondary/30">
        <p class="text-slate-500 animate-pulse">Loading admin...</p>
    </div>
{:else if isAccessDenied}
    <div class="flex min-h-screen items-center justify-center bg-secondary/30">
        <div class="text-center">
            <h1 class="text-2xl font-bold text-red-500">Access Denied</h1>
            <p class="text-slate-600 mt-2">You do not have permission to access the admin panel.</p>
            <a href="/auth" class="mt-4 inline-block text-sm text-slate-500 underline">Sign in with a different account</a>
        </div>
    </div>
{:else}
    <div class="flex min-h-screen bg-secondary/30">
        {#if showSidebar}
            <aside class="w-64 border-r border-border bg-card p-6 flex flex-col gap-6">
                <div>
                    <h1 class="text-xl font-bold">CMS Admin</h1>
                    <p class="text-xs text-muted-foreground">Resume Ranker</p>
                </div>
                <nav class="flex flex-col gap-2">
                    <a href="/" class="px-3 py-2 text-sm font-medium hover:bg-secondary rounded-md">Pages</a>
                    {#if $currentUser?.role === 'superadmin'}
                        <a href="/iam" class="px-3 py-2 text-sm font-medium hover:bg-secondary rounded-md text-indigo-600">IAM (Roles)</a>
                    {/if}
                </nav>
                <div class="mt-auto flex items-center justify-between">
                    <div class="overflow-hidden">
                        <p class="text-xs text-slate-500 truncate">{$currentUser?.email}</p>
                        <p class="text-xs text-slate-400 capitalize">{$currentUser?.role}</p>
                    </div>
                    <ThemeToggle />
                </div>
            </aside>
        {/if}
        <main class="flex-1 p-8">
            {@render children()}
        </main>
    </div>
{/if}
