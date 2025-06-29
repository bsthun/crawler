<script lang="ts">
	import * as Pagination from '$/lib/shadcn/components/ui/pagination'
	import type { HTMLAttributes } from 'svelte/elements'
	import { Input } from '$/lib/shadcn/components/ui/input/index.js'

	type props = HTMLAttributes<HTMLParagraphElement>
	let className: props['class'] = undefined

	export let currentPage: number
	export let count: number
	export let perPage: number
	export { className as class }

	const onChange = (event: Event) => {
		const input = event.target as HTMLInputElement
		const value = parseInt(input.value, 10)
		if (!isNaN(value) && value >= 1 && value <= Math.ceil(count / perPage)) {
			currentPage = value
		} else {
			input.value = currentPage.toString()
		}
	}
</script>

<Pagination.Root class={className} {count} {perPage}>
	{#snippet children({ pages, currentPage: _ })}
		<Pagination.Content>
			<Pagination.Item>
				<Pagination.PrevButton
					class={currentPage === 1 ? 'pointer-events-none opacity-50' : 'cursor-pointer'}
					onclick={() => {
						if (currentPage > 1) {
							currentPage = currentPage - 1
						}
					}}
				/>
			</Pagination.Item>
			{#each pages as page (page.key)}
				{#if page.type === 'ellipsis'}
					<Pagination.Item>
						<Pagination.Ellipsis />
					</Pagination.Item>
					<Pagination.Item>
						<Input
							type="number"
							value={currentPage}
							min={1}
							max={Math.ceil(count / perPage)}
							onblur={onChange}
							class="w-16 text-center"
							placeholder="Page"
						/>
					</Pagination.Item>
					<Pagination.Item>
						<Pagination.Ellipsis />
					</Pagination.Item>
				{:else}
					<Pagination.Item>
						<Pagination.Link
							{page}
							isActive={currentPage === page.value}
							onclick={() => {
								currentPage = page.value
							}}
						>
							{page.value}
						</Pagination.Link>
					</Pagination.Item>
				{/if}
			{/each}
			<Pagination.Item>
				<Pagination.NextButton
					class={currentPage * perPage >= count ? 'pointer-events-none opacity-50' : 'cursor-pointer'}
					onclick={() => {
						if (currentPage * perPage < count) {
							currentPage = currentPage + 1
						}
					}}
				/>
			</Pagination.Item>
		</Pagination.Content>
	{/snippet}
</Pagination.Root>
