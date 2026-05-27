<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import { acceptCurrentPolicy } from '$lib/stores/auth';
	import { CURRENT_POLICY_VERSION } from '$lib/policy';
	import { toast } from 'svelte-sonner';

	// Modal is shown by the /app layout when the cached user's stored
	// `acceptedPolicyVersion` is older than the bundle's CURRENT_POLICY_VERSION.
	// Two-way bound `open` lets the parent both show and react to dismissal.
	let { open = $bindable() }: { open: boolean } = $props();

	let submitting = $state(false);

	async function handleAccept() {
		submitting = true;
		try {
			await acceptCurrentPolicy(CURRENT_POLICY_VERSION);
			open = false;
		} catch {
			toast.error("Couldn't record acceptance", {
				description: 'Please try again, or sign out and back in.'
			});
		} finally {
			submitting = false;
		}
	}
</script>

<!--
	`onOpenChange` is intentionally a no-op: we never want the dialog to
	close from outside-click or ESC. The only exit path is the "I agree"
	button below, which writes the new acceptance to the backend before
	flipping `open`.
-->
<Dialog.Root bind:open onOpenChange={(next) => { if (!next && !submitting) open = true; }}>
	<Dialog.Content
		class="sm:max-w-md"
		showCloseButton={false}
		interactOutsideBehavior="ignore"
		escapeKeydownBehavior="ignore"
	>
		<Dialog.Header>
			<Dialog.Title>We've updated our terms</Dialog.Title>
			<Dialog.Description>
				Our Terms of Service and Privacy Policy have changed since you last
				signed in. Please review and accept to continue.
			</Dialog.Description>
		</Dialog.Header>

		<div class="text-muted-foreground space-y-2 text-sm">
			<p>
				Read the updated documents:
			</p>
			<ul class="list-disc space-y-1 pl-6">
				<li>
					<a
						href="/terms"
						target="_blank"
						rel="noopener"
						class="text-foreground underline hover:no-underline"
					>
						Terms of Service
					</a>
				</li>
				<li>
					<a
						href="/privacy"
						target="_blank"
						rel="noopener"
						class="text-foreground underline hover:no-underline"
					>
						Privacy Policy
					</a>
				</li>
			</ul>
			<p class="text-xs">
				Version <span class="font-mono">{CURRENT_POLICY_VERSION}</span>. By clicking
				"I agree", you accept both documents as currently published.
			</p>
		</div>

		<Dialog.Footer class="gap-2 sm:gap-2">
			<Button type="button" disabled={submitting} onclick={handleAccept}>
				{submitting ? 'Saving...' : 'I agree'}
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
