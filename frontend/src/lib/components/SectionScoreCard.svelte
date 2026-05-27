<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import type { SectionScore } from '$lib/types';
	import RequirementRow from './RequirementRow.svelte';
	import ScoreBar from './ScoreBar.svelte';

	interface Props {
		section: SectionScore;
	}

	const { section }: Props = $props();
</script>

<Card.Root>
	<Card.Header>
		<ScoreBar label={section.label} score={section.score} />
	</Card.Header>
	<Card.Content class="px-0 pb-0">
		<ul class="border-border border-t">
			{#each section.requirements as req (req.id)}
				<RequirementRow requirement={req} />
			{/each}
		</ul>
	</Card.Content>
</Card.Root>
