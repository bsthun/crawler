<script lang="ts">
    import {onMount, getContext} from 'svelte'
    import {backend} from '$/util/backend.ts'
    import type {Writable} from 'svelte/store'
    import type {Setup} from '$/util/type/setup'
    import type {PayloadOverview} from '$/util/backend/backend'
    import * as Card from "$/lib/shadcn/components/ui/card"
    import * as Chart from "$/lib/shadcn/components/ui/chart"
    import {scaleBand} from "d3-scale"
    import {BarChart, type ChartContextValue} from "layerchart"
    import {cubicInOut} from "svelte/easing"
    import {BarChart3, Wallet, Clock} from "lucide-svelte"
    import {catcher} from "$/util/backend.js";
    import {Link} from "svelte-navigator";
    import Container from "$/component/layout/Container.svelte";

    const setup = getContext<Writable<Setup>>('setup')
    let overviewData = $state<PayloadOverview | null>(null)
    let context = $state<ChartContextValue>()

    const mount = () => {
        backend.state.stateOverview()
            .then((res) => {
                overviewData = res.data
            })
            .catch((err) => {
                catcher(err)
            })
    }

    onMount(mount)

    const chartConfig = {
        completed: {label: "Completed", color: "#A78BFA"},
        failed: {label: "Failed", color: "#C4B5FD"},
        pending: {label: "Pending", color: "#DDD6FE"}
    } satisfies Chart.ChartConfig
</script>

<Container class="flex flex-col gap-6">
    {#if $setup?.profile?.id}
        <Card.Root>
            <Card.Content>
                <div class="flex items-center space-x-4">
                    <div class="w-12 h-12 rounded-full bg-gradient-to-r from-blue-500 to-purple-600 flex items-center justify-center text-white font-semibold text-lg">
                        {$setup.profile.name?.charAt(0)?.toUpperCase() || $setup.profile.email?.charAt(0)?.toUpperCase() || 'U'}
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
            <Card.Title>7-Day Activity History</Card.Title>
            <Card.Description>Task status breakdown over the last 7 days</Card.Description>
        </Card.Header>
        <Card.Content>
            <div class="flex items-center gap-6">
                <div class="flex-[1]">
                    <Chart.Container class="h-96 py-12 chart" config={chartConfig}>
                        <BarChart
                                axis="x"
                                bind:context
                                data={overviewData?.histories?.map((history, index) => ({
                                    date: `Day ${index - 6}`,
                                    completed: history.completed,
                                    failed: history.failed,
                                    pending: history.pending
                                })) || []}
                                grid={false}
                                highlight={false}
                                props={{
                                    bars: {
                                        stroke: "none",
                                        initialY: context?.height,
                                        initialHeight: 0,
                                        motion: {
                                            y: { type: "tween", duration: 500, easing: cubicInOut },
                                            height: { type: "tween", duration: 500, easing: cubicInOut }
                                        }
                                    },
                                }}
                                rule={false}
                                series={[
                                    {
                                        key: "completed",
                                        label: "Completed",
                                        color: chartConfig.completed.color,
                                        props: { rounded: "bottom" }
                                    },
                                    {
                                        key: "failed",
                                        label: "Failed",
                                        color: chartConfig.failed.color
                                    },
                                    {
                                        key: "pending",
                                        label: "Pending",
                                        color: chartConfig.pending.color
                                    }
                                ]}
                                seriesLayout="stack"
                                x="date"
                                xScale={scaleBand().padding(0.4)}
                        >
                            {#snippet tooltip()}
                                <Chart.Tooltip/>
                            {/snippet}
                        </BarChart>
                    </Chart.Container>
                </div>

                <div class="flex-[2] space-y-4">
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

                    <div class="text-center p-4 bg-purple-50 rounded-lg">
                        <div class="text-3xl font-bold text-purple-600">
                            {overviewData?.histories?.reduce((sum, history) =>
                                sum + history.completed + history.failed + history.pending, 0
                            ) || 0}
                        </div>
                        <div class="text-lg text-gray-600 font-medium">Total Tasks</div>
                    </div>

                    <Link to="/home/task">
                        <Card.Root class="hover:bg-muted/50 transition-colors cursor-pointer">
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
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <Card.Root>
                <Card.Header class="flex flex-row items-center justify-between space-y-0 pb-2">
                    <Card.Title class="text-sm font-medium">Total Tokens</Card.Title>
                    <BarChart3 class="w-4 h-4 text-muted-foreground"/>
                </Card.Header>
                <Card.Content>
                    <div class="text-2xl font-bold">{overviewData.tokenCount.toLocaleString()}</div>
                </Card.Content>
            </Card.Root>

            <Card.Root>
                <Card.Header class="flex flex-row items-center justify-between space-y-0 pb-2">
                    <Card.Title class="text-sm font-medium">Pool Tokens</Card.Title>
                    <Wallet class="w-4 h-4 text-muted-foreground"/>
                </Card.Header>
                <Card.Content>
                    <div class="text-2xl font-bold">{overviewData.poolTokenCount.toLocaleString()}</div>
                </Card.Content>
            </Card.Root>
        </div>
    {/if}
</Container>

<style lang="postcss">
    :global(.lc-tooltip-rect) {
        transform: translateY(-100%);
    }
</style>