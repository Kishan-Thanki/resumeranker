<script lang="ts">
	import { ChevronDown } from '@lucide/svelte';
	import type { RequirementMatch } from '$lib/types';
	import { cn } from '$lib/utils';
	import EvidenceBlock from './EvidenceBlock.svelte';
	import MatchBadge from './MatchBadge.svelte';

	interface Props {
		requirement: RequirementMatch;
	}

	const { requirement }: Props = $props();
	let open = $state(false);
</script>

<li class="border-border border-b last:border-b-0">
	<button
		type="button"
		class="hover:bg-muted/40 focus-visible:ring-ring flex w-full items-center justify-between gap-3 px-4 py-3 text-left text-sm transition-colors focus-visible:ring-2 focus-visible:outline-hidden"
		aria-expanded={open}
		onclick={() => (open = !open)}
	>
		<span class="min-w-0 flex-1">{requirement.requirement}</span>
		<span class="flex shrink-0 items-center gap-2">
			<MatchBadge strength={requirement.matchStrength} matched={requirement.matched} />
			<ChevronDown
				class={cn('text-muted-foreground size-4 transition-transform', open && 'rotate-180')}
			/>
		</span>
	</button>

	{#if open}
		<div class="space-y-3 px-4 pb-4 pt-1">
			<div class="grid gap-3 md:grid-cols-2">
				<div>
					<p class="text-muted-foreground mb-1.5 text-xs font-medium uppercase tracking-wide">
						From the job description
					</p>
					<EvidenceBlock evidence={requirement.jdEvidence} />
				</div>
				<div>
					<p class="text-muted-foreground mb-1.5 text-xs font-medium uppercase tracking-wide">
						{requirement.matched ? 'From the resume' : 'In the resume'}
					</p>
					{#if requirement.resumeEvidence.length > 0}
						<div class="space-y-2">
							{#each requirement.resumeEvidence as ev (ev.text)}
								<EvidenceBlock evidence={ev} />
							{/each}
						</div>
					{:else}
						<div
							class="border-border bg-muted/30 text-muted-foreground rounded-md border p-3 text-sm italic"
						>
							Not found in resume.
						</div>
					{/if}
				</div>
			</div>
			{#if requirement.note}
				<p class="text-muted-foreground text-xs">{requirement.note}</p>
			{/if}
		</div>
	{/if}
</li>
