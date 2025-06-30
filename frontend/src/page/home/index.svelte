<script lang="ts">
	import { onMount, getContext } from 'svelte'
	import { backend } from '$/util/backend.ts'
	import type { Writable } from 'svelte/store'
	import type { Setup } from '$/util/type/setup'
	import type { PayloadOverview, PayloadUserListResponse } from '$/util/backend/backend'
	import * as Card from '$/lib/shadcn/components/ui/card'
	import * as Chart from '$/lib/shadcn/components/ui/chart'
	import { scaleBand } from 'd3-scale'
	import { BarChart, type ChartContextValue } from 'layerchart'
	import { cubicInOut } from 'svelte/easing'
	import { BarChart3 } from 'lucide-svelte'
	import { catcher } from '$/util/backend.js'
	import { Link } from 'svelte-navigator'
	import Container from '$/component/layout/Container.svelte'

	const setup = getContext<Writable<Setup>>('setup')
	let overviewData = $state<PayloadOverview | null>(null)
	let context = $state<ChartContextValue>()
	let users = $state<PayloadUserListResponse | null>(null)
	let selectedUserId = $state<string | null>(null)

	const loadFilters = () => {
		if ($setup?.profile?.isAdmin) {
			backend.admin.userList()
				.then((res) => {
					if (res.success && res.data) {
						users = res.data
					}
				})
				.catch((err) => {
					catcher(err)
				})
		}
	}

	const mount = () => {
		const userId = $setup?.profile?.isAdmin ? selectedUserId : null
		backend.state
			.stateOverview({
				userId: userId as any,
			})
			.then((res) => {
				overviewData = res.data
			})
			.catch((err) => {
				catcher(err)
			})
	}

	$effect(() => {
		if (selectedUserId === null && $setup?.profile?.userId) {
			selectedUserId = $setup.profile.userId
		}
	})

	$effect(() => {
		if (selectedUserId !== undefined) {
			mount()
		}
	})

	onMount(() => {
		loadFilters()
		mount()
	})

	const chartConfig = {
		completed: { label: 'Completed', color: '#A78BFA' },
		failed: { label: 'Failed', color: '#C4B5FD' },
		pending: { label: 'Pending', color: '#DDD6FE' },
	} satisfies Chart.ChartConfig
</script>

<Container class="flex flex-col gap-6">
	{#if $setup?.profile?.userId}
		<Card.Root>
			<Card.Content>
				<div class="flex items-center space-x-4">
					<div
						class="flex h-12 w-12 items-center justify-center rounded-full bg-gradient-to-r from-blue-500 to-purple-600 text-lg font-semibold text-white"
					>
						{$setup.profile.name?.charAt(0)?.toUpperCase() ||
						$setup.profile.email?.charAt(0)?.toUpperCase() ||
						'U'}
					</div>
					<div class="flex-1">
						<h2 class="text-xl font-semibold text-gray-900">{$setup.profile.name || 'User'}</h2>
						<p class="text-gray-600">{$setup.profile.email}</p>
					</div>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}


	<Card.Root>
		<Card.Header>
			<Card.Title>Activity History</Card.Title>
			<Card.Description>Task status breakdown over the last 7 days</Card.Description>
		</Card.Header>
		<Card.Content>
			<div class="flex items-center gap-6">
				<div class="flex-[1]">
					<Chart.Container class="chart h-96 py-12" config={chartConfig}>
						<BarChart
							axis="x"
							bind:context
							data={overviewData?.histories?.map((history, index) => ({
								date: `Day ${index - 6}`,
								completed: history.completed,
								failed: history.failed,
								pending: history.pending,
							})) || []}
							grid={false}
							highlight={false}
							props={{
								bars: {
									stroke: 'white',
									initialY: context?.height,
									initialHeight: 0,
									motion: {
										y: { type: 'tween', duration: 500, easing: cubicInOut },
										height: { type: 'tween', duration: 500, easing: cubicInOut },
									},
								},
							}}
							rule={false}
							series={[
								{
									key: 'completed',
									label: 'Completed',
									color: chartConfig.completed.color,
								},
								{
									key: 'failed',
									label: 'Failed',
									color: chartConfig.failed.color,
								},
								{
									key: 'pending',
									label: 'Pending',
									color: chartConfig.pending.color,
								},
							]}
							seriesLayout="stack"
							x="date"
							xScale={scaleBand().padding(0.4)}
						>
							{#snippet tooltip()}
								<Chart.Tooltip />
							{/snippet}
						</BarChart>
					</Chart.Container>
				</div>

				<div class="flex-[1] space-y-4">
					{#if $setup?.profile?.isAdmin}
						<div class="flex flex-col gap-2">
							<label for="user-select" class="text-sm font-medium">View User</label>
							<select id="user-select" bind:value={selectedUserId} class="rounded-md border px-3 py-2 w-full">
								{#if users?.users}
									{#each users.users as user}
										<option value={user.id}>
											{user.firstname} {user.lastname} ({user.email})
										</option>
									{/each}
								{/if}
							</select>
						</div>
					{/if}
					<div class="grid grid-cols-3 gap-4">
						<div class="text-center">
							<div class="text-2xl font-bold text-purple-500">
								{overviewData?.histories?.reduce((sum, history) => sum + history.completed, 0) || 0}
							</div>
							<div class="text-sm text-gray-500">Completed</div>
						</div>
						<div class="text-center">
							<div class="text-2xl font-bold text-purple-400">
								{overviewData?.histories?.reduce((sum, history) => sum + history.failed, 0) || 0}
							</div>
							<div class="text-sm text-gray-500">Failed</div>
						</div>
						<div class="text-center">
							<div class="text-2xl font-bold text-purple-300">
								{overviewData?.histories?.reduce((sum, history) => sum + history.pending, 0) || 0}
							</div>
							<div class="text-sm text-gray-500">Pending</div>
						</div>
					</div>

					<div class="rounded-lg bg-purple-50 p-4 text-center">
						<div class="text-3xl font-bold text-purple-600">
							{overviewData?.histories?.reduce(
								(sum, history) => sum + history.completed + history.failed + history.pending,
								0
							) || 0}
						</div>
						<div class="text-lg font-medium text-gray-600">Total Tasks</div>
					</div>

					<Link to="/home/task">
						<Card.Root class="hover:bg-muted/50 cursor-pointer transition-colors">
							<Card.Content class="flex items-center justify-center">
								<span class="text-lg font-medium">View All Tasks</span>
							</Card.Content>
						</Card.Root>
					</Link>
				</div>
			</div>
		</Card.Content>
	</Card.Root>

	{#if overviewData}
		<div class="grid grid-cols-1 gap-6 md:grid-cols-2">
			<Card.Root>
				<Card.Header class="flex flex-row items-center justify-between space-y-0 pb-2">
					<Card.Title class="text-sm font-medium">Total Tokens</Card.Title>
					<BarChart3 class="text-muted-foreground h-4 w-4" />
				</Card.Header>
				<Card.Content>
					<div class="text-2xl font-bold">{overviewData.tokenCount.toLocaleString()}</div>
				</Card.Content>
			</Card.Root>

			<Card.Root>
				<Card.Header class="flex flex-row items-center justify-between space-y-0 pb-2">
					<Card.Title class="text-sm font-medium">Pool Tokens</Card.Title>
					<BarChart3 class="text-muted-foreground h-4 w-4" />
				</Card.Header>
				<Card.Content>
					<div class="grid grid-cols-2 gap-0 lg:grid-cols-4">
						{#each overviewData.poolTokens || [] as category, index}
							<div
								class="flex flex-col items-center justify-center border-r border-b border-gray-200 p-4 last:border-r-0 {index >=
								(overviewData.poolTokens?.length || 0) -
									((overviewData.poolTokens?.length || 0) % 2 === 0 ? 2 : 1)
									? 'border-b-0'
									: ''}"
							>
								<div class="mb-1 text-xs font-medium text-gray-500 uppercase">
									{category.categoryName}
								</div>
								<div class="text-xl font-bold">
									{Math.floor(category.tokenCount / 1000).toLocaleString()}k
								</div>
							</div>
						{/each}
					</div>
				</Card.Content>
			</Card.Root>
		</div>
	{/if}
</Container>
