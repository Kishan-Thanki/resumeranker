<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Textarea } from '$lib/components/ui/textarea';
	import { api } from '$lib/api';
	import Wordmark from '$lib/components/Wordmark.svelte';

	let name = $state('');
	let email = $state('');
	let message = $state('');
	// Honeypot field. NOT rendered as an input; bots autofilling all
	// form fields will set it, and the backend silently rejects.
	let website = $state('');
	let submitting = $state(false);
	let submitted = $state(false);
	let error = $state<string | null>(null);

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		if (!name.trim() || !email.trim() || message.trim().length < 10) return;
		submitting = true;
		error = null;
		try {
			await api.sendContactMessage({
				name: name.trim(),
				email: email.trim(),
				message: message.trim(),
				website
			});
			submitted = true;
		} catch {
			error = "Couldn't send your message. Please try again in a moment.";
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>Contact · Resume Ranker</title>
	<meta name="description" content="Send a message to Resume Ranker." />
</svelte:head>

<div class="bg-background flex-1">
	<header class="mx-auto flex h-14 max-w-3xl items-center justify-between px-4 sm:px-6">
		<Wordmark href="/" />
	</header>

	<main class="mx-auto max-w-md px-4 pt-8 pb-24 sm:px-6 sm:pt-12">
		<h1 class="text-2xl font-semibold tracking-tight sm:text-3xl">Contact</h1>
		<p class="text-muted-foreground mt-2 text-sm">
			For privacy questions, security reports, accessibility issues, or anything
			else, send us a message below.
		</p>

		{#if submitted}
			<div class="border-border mt-10 rounded-lg border p-6">
				<h2 class="text-base font-semibold">Thanks — we received your message.</h2>
				<p class="text-muted-foreground mt-2 text-sm">
					We'll get back to you at the email address you provided. If your message
					was urgent (security or abuse), we aim to respond within 72 hours.
				</p>
				<div class="mt-6">
					<Button variant="outline" href="/">Back to home</Button>
				</div>
			</div>
		{:else}
			<form onsubmit={handleSubmit} class="mt-10 space-y-5">
				<div class="space-y-2">
					<Label for="contact-name">Your name</Label>
					<Input
						id="contact-name"
						type="text"
						autocomplete="name"
						required
						maxlength={100}
						bind:value={name}
						placeholder="Jane Doe"
					/>
				</div>

				<div class="space-y-2">
					<Label for="contact-email">Your email</Label>
					<Input
						id="contact-email"
						type="email"
						autocomplete="email"
						required
						bind:value={email}
						placeholder="you@example.com"
					/>
					<p class="text-muted-foreground text-xs">
						We'll reply to this address. Not used for anything else.
					</p>
				</div>

				<div class="space-y-2">
					<Label for="contact-message">Message</Label>
					<Textarea
						id="contact-message"
						required
						minlength={10}
						maxlength={5000}
						rows={6}
						bind:value={message}
						placeholder="What's on your mind?"
					/>
					<p class="text-muted-foreground text-xs">
						Up to 5000 characters. Avoid pasting anything sensitive — once you
						send, the message lives in our admin inbox.
					</p>
				</div>

				<!--
					Honeypot field. Visually hidden + aria-hidden so screen readers
					skip it. Bots that autofill all form fields tip themselves off.
				-->
				<div
					aria-hidden="true"
					style="position:absolute;left:-9999px;width:1px;height:1px;overflow:hidden;"
				>
					<label for="contact-website">Leave this field empty</label>
					<input
						id="contact-website"
						type="text"
						tabindex="-1"
						autocomplete="off"
						bind:value={website}
					/>
				</div>

				{#if error}
					<p class="text-destructive text-sm">{error}</p>
				{/if}

				<Button
					type="submit"
					class="w-full"
					disabled={!name.trim() || !email.trim() || message.trim().length < 10 || submitting}
				>
					{submitting ? 'Sending...' : 'Send message'}
				</Button>

				<p class="text-muted-foreground text-center text-xs">
					By sending, you agree to our
					<a class="hover:text-foreground underline" href="/privacy">Privacy Policy</a>.
				</p>
			</form>
		{/if}
	</main>
</div>
