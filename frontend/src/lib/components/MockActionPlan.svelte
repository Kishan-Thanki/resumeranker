<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, slide } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	
	type ActionItem = {
		type: 'missing' | 'matched' | 'improvement';
		title: string;
		desc: string;
	};

	const items: ActionItem[] = [
		{ 
			type: 'missing', 
			title: 'Missing Keyword: Kubernetes', 
			desc: "The JD explicitly requires this, but it wasn't detected in your resume. Consider adding a bullet point demonstrating your experience with this." 
		},
		{ 
			type: 'matched', 
			title: 'Matched: Python (FastAPI)', 
			desc: 'Found in your Experience section: "Built scalable microservices using FastAPI and Python."' 
		},
		{ 
			type: 'improvement', 
			title: 'Action: Quantify Impact', 
			desc: 'You mention increasing revenue, but missed adding specific metrics (e.g. "by 25%"). Adding numbers helps ATS score you higher.' 
		}
	];

	let visibleItems = $state<ActionItem[]>([]);
	let isScanning = $state(true);
	let mounted = $state(false);

	onMount(() => {
		mounted = true;
		let index = 0;
		let timeoutId: ReturnType<typeof setTimeout>;
		
		const tick = () => {
			if (index < items.length) {
				visibleItems = [...visibleItems, items[index]];
				index++;
				timeoutId = setTimeout(tick, 1800);
			} else {
				isScanning = false;
				timeoutId = setTimeout(() => {
					visibleItems = [];
					isScanning = true;
					index = 0;
					timeoutId = setTimeout(tick, 1200);
				}, 4000);
			}
		};

		timeoutId = setTimeout(tick, 1000);
		
		return () => {
			clearTimeout(timeoutId);
		};
	});
</script>

<div class="border-border bg-card/50 mt-16 overflow-hidden rounded-2xl border shadow-2xl">
	<div class="bg-muted/40 border-border flex items-center justify-between border-b px-4 py-3">
		<div class="flex items-center gap-2">
			<div class="flex gap-1.5">
				<div class="bg-destructive/60 size-3 rounded-full"></div>
				<div class="bg-amber-500/60 size-3 rounded-full"></div>
				<div class="bg-emerald-500/60 size-3 rounded-full"></div>
			</div>
			<span class="text-muted-foreground ml-2 text-xs font-medium uppercase tracking-wider">Live Analysis</span>
		</div>
		{#if isScanning}
			<div class="flex items-center gap-2">
				<span class="bg-indigo-500 size-2 rounded-full animate-pulse"></span>
				<span class="text-indigo-500 text-xs font-medium animate-pulse">Scanning...</span>
			</div>
		{:else}
			<div class="flex items-center gap-1.5">
				<span class="bg-emerald-500 size-2 rounded-full"></span>
				<span class="text-emerald-600 dark:text-emerald-400 text-xs font-medium">Complete</span>
			</div>
		{/if}
	</div>
	
	<div class="p-6">
		<div class="border-border bg-card rounded-lg border shadow-sm overflow-hidden min-h-[310px]">
			{#if visibleItems.length === 0 && mounted}
				<div class="flex h-full min-h-[310px] flex-col items-center justify-center text-center p-6 text-muted-foreground" in:fade={{ duration: 300 }}>
					<div class="relative flex items-center justify-center size-12 mb-4">
						<span class="absolute inset-0 border-2 border-indigo-500/30 rounded-full animate-[ping_2s_cubic-bezier(0,0,0.2,1)_infinite]"></span>
						<span class="relative size-4 bg-indigo-500 rounded-full animate-pulse"></span>
					</div>
					<p class="text-sm font-medium animate-pulse">Extracting keywords from JD...</p>
				</div>
			{/if}

			<div class="divide-border divide-y">
				{#each visibleItems as item (item.title)}
					<div 
						class="flex items-start gap-4 p-4"
						in:slide={{ duration: 400, easing: cubicOut, axis: 'y' }}
					>
						{#if item.type === 'missing'}
							<div class="bg-destructive/10 text-destructive mt-0.5 flex h-6 w-6 shrink-0 items-center justify-center rounded-full">
								<span class="text-sm font-bold">!</span>
							</div>
						{:else if item.type === 'matched'}
							<div class="bg-emerald-500/10 text-emerald-500 mt-0.5 flex h-6 w-6 shrink-0 items-center justify-center rounded-full">
								<span class="text-sm font-bold">✓</span>
							</div>
						{:else}
							<div class="bg-amber-500/10 text-amber-500 mt-0.5 flex h-6 w-6 shrink-0 items-center justify-center rounded-full">
								<span class="text-sm font-bold">↑</span>
							</div>
						{/if}
						<div in:fade={{ delay: 200, duration: 400 }}>
							<p class="font-medium text-foreground">{item.title}</p>
							<p class="text-muted-foreground mt-1 text-sm leading-relaxed">
								{item.desc}
							</p>
						</div>
					</div>
				{/each}
			</div>
		</div>
	</div>
</div>
