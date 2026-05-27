<script lang="ts">
	import { goto } from '$app/navigation';
	import { Moon, Sun, LogOut, Trash2, User } from '@lucide/svelte';
	import { toggleMode, mode } from 'mode-watcher';
	import { toast } from 'svelte-sonner';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import { deleteAccount, signOut } from '$lib/stores/auth';
	import { clearAnalyses } from '$lib/stores/analyses';

	let deleteDialogOpen = $state(false);
	let deleting = $state(false);

	async function handleSignOut() {
		await signOut();
		clearAnalyses();
		await goto('/');
	}

	async function handleDeleteAccount() {
		// Hard-delete on the backend. Cascades wipe sessions, magic-links,
		// and analyses. We clear local state and route home with a toast
		// so the user sees confirmation that data is gone.
		deleting = true;
		try {
			await deleteAccount();
			clearAnalyses();
			deleteDialogOpen = false;
			toast('Account deleted', {
				description: 'Your data has been removed.'
			});
			await goto('/');
		} catch {
			toast.error("Couldn't delete account", {
				description: 'Please try again, or contact us if it keeps failing.'
			});
		} finally {
			deleting = false;
		}
	}
</script>

<header class="border-border bg-background sticky top-0 z-10 border-b">
	<div class="mx-auto flex h-14 max-w-5xl items-center justify-between px-4 sm:px-6">
		<a href="/app" class="text-sm font-semibold tracking-tight">Resume Ranker</a>
		<div class="flex items-center gap-1">
			<Button
				variant="ghost"
				size="icon"
				onclick={toggleMode}
				aria-label="Toggle color theme"
			>
				{#if mode.current === 'dark'}
					<Sun class="size-4" />
				{:else}
					<Moon class="size-4" />
				{/if}
			</Button>
			<DropdownMenu.Root>
				<DropdownMenu.Trigger>
					{#snippet child({ props })}
						<Button variant="ghost" size="icon" aria-label="Account menu" {...props}>
							<User class="size-4" />
						</Button>
					{/snippet}
				</DropdownMenu.Trigger>
				<DropdownMenu.Content align="end" class="w-48">
					<DropdownMenu.Item onSelect={handleSignOut}>
						<LogOut class="size-4" />
						Sign out
					</DropdownMenu.Item>
					<DropdownMenu.Separator />
					<DropdownMenu.Item
						class="text-destructive focus:text-destructive"
						onSelect={() => (deleteDialogOpen = true)}
					>
						<Trash2 class="size-4" />
						Delete account
					</DropdownMenu.Item>
				</DropdownMenu.Content>
			</DropdownMenu.Root>
		</div>
	</div>
</header>

<Dialog.Root bind:open={deleteDialogOpen}>
	<Dialog.Content class="sm:max-w-md">
		<Dialog.Header>
			<Dialog.Title>Delete your account?</Dialog.Title>
			<Dialog.Description>
				This is permanent. We'll immediately and permanently remove:
			</Dialog.Description>
		</Dialog.Header>

		<ul class="text-muted-foreground list-disc space-y-1 pl-6 text-sm">
			<li>Your account and email address</li>
			<li>All of your analyses and their content</li>
			<li>Your active sign-in session</li>
		</ul>

		<p class="text-muted-foreground text-xs">
			This cannot be undone. If you sign up again later with the same email,
			you'll start from an empty account.
		</p>

		<Dialog.Footer class="gap-2 sm:gap-2">
			<Button
				type="button"
				variant="outline"
				disabled={deleting}
				onclick={() => (deleteDialogOpen = false)}
			>
				Cancel
			</Button>
			<Button
				type="button"
				variant="destructive"
				disabled={deleting}
				onclick={handleDeleteAccount}
			>
				{deleting ? 'Deleting...' : 'Delete account'}
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
