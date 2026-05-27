<script lang="ts">
	import { Upload, X } from '@lucide/svelte';
	import { cn } from '$lib/utils';

	interface Props {
		accept: string;
		acceptLabel: string;
		file: File | null;
		onChange: (file: File | null) => void;
		id?: string;
	}

	let { accept, acceptLabel, file, onChange, id = 'file-upload' }: Props = $props();

	let dragging = $state(false);
	let inputEl: HTMLInputElement | null = $state(null);

	function pickFromList(list: FileList | null | undefined): File | null {
		if (!list || list.length === 0) return null;
		const next = list.item(0);
		if (!next) return null;
		if (accept) {
			const ok = accept
				.split(',')
				.map((t) => t.trim())
				.some((t) => {
					if (t.startsWith('.')) return next.name.toLowerCase().endsWith(t.toLowerCase());
					if (t.endsWith('/*')) return next.type.startsWith(t.slice(0, -1));
					return next.type === t;
				});
			if (!ok) return null;
		}
		return next;
	}

	function handleInput(event: Event) {
		const target = event.currentTarget as HTMLInputElement;
		onChange(pickFromList(target.files));
	}

	function handleDrop(event: DragEvent) {
		event.preventDefault();
		dragging = false;
		onChange(pickFromList(event.dataTransfer?.files));
	}

	function handleDragOver(event: DragEvent) {
		event.preventDefault();
		dragging = true;
	}

	function handleDragLeave() {
		dragging = false;
	}

	function clear() {
		onChange(null);
		if (inputEl) inputEl.value = '';
	}

	function formatSize(bytes: number): string {
		if (bytes < 1024) return `${bytes} B`;
		if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
		return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
	}
</script>

{#if file}
	<div
		class="border-border bg-card flex items-start justify-between gap-3 rounded-md border p-4"
	>
		<div class="min-w-0">
			<p class="truncate text-sm font-medium">{file.name}</p>
			<p class="text-muted-foreground mt-0.5 text-xs">{formatSize(file.size)}</p>
		</div>
		<button
			type="button"
			class="hover:bg-muted text-muted-foreground hover:text-foreground rounded-md p-1.5 transition-colors"
			onclick={clear}
			aria-label="Remove file"
		>
			<X class="size-4" />
		</button>
	</div>
{:else}
	<label
		for={id}
		ondrop={handleDrop}
		ondragover={handleDragOver}
		ondragleave={handleDragLeave}
		class={cn(
			'border-border bg-card hover:bg-muted/40 flex cursor-pointer flex-col items-center justify-center gap-2 rounded-md border border-dashed px-4 py-10 text-center transition-colors',
			dragging && 'border-indigo-400 bg-indigo-50/60 dark:border-indigo-500 dark:bg-indigo-950/30'
		)}
	>
		<Upload class={cn('size-5 transition-colors', dragging ? 'text-indigo-500' : 'text-muted-foreground')} />
		<p class="text-sm">
			<span class="font-medium">Click to upload</span>
			<span class="text-muted-foreground">or drag and drop</span>
		</p>
		<p class="text-muted-foreground text-xs">{acceptLabel}</p>
		<input
			bind:this={inputEl}
			{id}
			type="file"
			class="sr-only"
			{accept}
			onchange={handleInput}
		/>
	</label>
{/if}
