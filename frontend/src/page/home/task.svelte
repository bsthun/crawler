<script lang="ts">
    import {
        Card,
        CardContent,
        CardDescription,
        CardFooter,
        CardHeader,
        CardTitle,
    } from '$/lib/shadcn/components/ui/card'
    import {Button} from '$/lib/shadcn/components/ui/button'
    import {Loader2Icon, Plus} from 'lucide-svelte'
    import {backend, catcher} from '$/util/backend.ts'
    import TaskSubmitDialog from './component/TaskSubmitDialog.svelte'
    import type {
        PayloadTaskListResponse,
        PayloadTaskUploadListResponse,
        PayloadTaskCategoryListResponse,
        PayloadTaskListItem,
    } from '$/util/backend/backend.ts'
    import Container from '$/component/layout/Container.svelte'
    import Pagination from '$/component/share/Pagination.svelte'

    let tasks: PayloadTaskListResponse | null = null
    let uploads: PayloadTaskUploadListResponse | null = null
    let categories: PayloadTaskCategoryListResponse | null = null
    let loading = true
    let currentPage = 1
    let perPage = 12
    let selectedUploadId: number | null = null
    let selectedCategoryId = 0
    let dialogOpen = false

    const formatDate = (dateString: string) => {
        return new Date(dateString).toLocaleDateString()
    }

    const formatTime = (dateString: string) => {
        return new Date(dateString).toLocaleTimeString([], {
            hour: '2-digit',
            minute: '2-digit',
        })
    }

    const loadFilters = () => {
        Promise.all([backend.task.taskUploadList(), backend.task.taskCategoryList()])
            .then(([uploadsRes, categoriesRes]) => {
                if (uploadsRes.success && uploadsRes.data) {
                    uploads = uploadsRes.data
                }
                if (categoriesRes.success && categoriesRes.data) {
                    categories = categoriesRes.data
                }
            })
            .catch((err) => {
                catcher(err)
            })
    }

    const mount = () => {
        loading = true
        const offset = (currentPage - 1) * perPage

        backend.task
            .taskList({
                limit: perPage,
                offset,
                uploadId: selectedUploadId!,
            })
            .then((res) => {
                if (res.success && res.data) {
                    tasks = res.data
                }
                loading = false
            })
            .catch((err) => {
                loading = false
                catcher(err)
            })
    }

    const filteredTasks = (tasks: PayloadTaskListItem[]) => {
        if (selectedCategoryId === 0) return tasks
        return tasks.filter((task) => task.categoryId === selectedCategoryId)
    }

    $: {
        if (currentPage || selectedUploadId !== undefined || selectedCategoryId !== undefined) {
            mount()
        }
    }

    loadFilters()
    mount()
</script>

<Container>
    <div class="mb-6 flex items-center justify-between">
        <h1 class="text-3xl font-bold">Tasks</h1>
        {#if categories}
            <TaskSubmitDialog bind:open={dialogOpen} {categories} on:submitted={mount}>
                <Button class="gap-2">
                    <Plus class="h-4 w-4"/>
                    Add Task
                </Button>
            </TaskSubmitDialog>
        {/if}
    </div>

    <!-- Filters -->
    <div class="mb-6 flex gap-4">
        <div class="flex flex-col gap-2">
            <label class="text-sm font-medium">Upload</label>
            <select bind:value={selectedUploadId} class="min-w-[200px] rounded-md border px-3 py-2">
                <option value={null}>All Uploads</option>
                {#if uploads?.uploads}
                    {#each uploads.uploads as upload}
                        <option value={upload.id}>
                            #{upload.id}
                            {formatDate(upload.createdAt)}
                        </option>
                    {/each}
                {/if}
            </select>
        </div>
        <div class="flex flex-col gap-2">
            <label class="text-sm font-medium">Category</label>
            <select bind:value={selectedCategoryId} class="min-w-[200px] rounded-md border px-3 py-2">
                <option value={0}>All Categories</option>
                {#if categories?.categories}
                    {#each categories.categories as category}
                        <option value={category.id}>{category.name}</option>
                    {/each}
                {/if}
            </select>
        </div>
    </div>

    {#if loading || !tasks}
        <div class="flex min-h-[400px] items-center justify-center">
            <Loader2Icon class="text-primary h-8 w-8 animate-spin"/>
        </div>
    {:else if tasks.count === 0}
        <p class="text-muted-foreground text-lg">No tasks found</p>
    {:else}
        <!-- Task List -->
        <div class="mb-6 grid gap-4">
            {#each filteredTasks(tasks.tasks || []) as task}
                <Card class="shadow-sm py-4 transition-shadow duration-200 hover:shadow-md">
                    <div class="flex items-center justify-between px-4 gap-4">
                        <div class="flex flex-1 items-center gap-4">
                            <div class="flex flex-col min-w-[86px] items-center gap-1">
                                <span class="text-muted-foreground text-sm">{formatDate(task.createdAt)}</span>
                                <span class="text-muted-foreground text-sm">{formatTime(task.createdAt)}</span>
                            </div>

                            <div class="flex flex-col min-w-[128px] gap-1">
                                <span class="text-sm font-medium">#{task.id}</span>
                                {#if task.uploadId}
                                    <span class="text-sm font-medium min-w-6 opacity-50">#{task.uploadId}</span>
                                {/if}
                            </div>

                            <!-- Type and Tokens -->
                            <div class="flex flex-col gap-1">
                                <div class="text-muted-foreground text-sm">
                                    {task.type} • {task.tokenCount} tokens
                                </div>
                                {#if task.source}
                                    <div class="max-w-screen-xs truncate text-sm text-blue-600 hover:text-blue-800">
                                        <a href={task.source} target="_blank" rel="noopener noreferrer"
                                           title={task.source}>
                                            {task.source}
                                        </a>
                                    </div>
                                {/if}
                            </div>
                        </div>

                        {#if task.failedReason}
                            <div class="max-w-[150px] truncate text-sm text-red-600" title={task.failedReason}>
                                {task.failedReason}
                            </div>
                        {/if}
                        <span
                                class="rounded-full px-2 py-1 text-xs {task.status === 'completed'
								? 'bg-green-100 text-green-800'
								: task.status === 'failed'
									? 'bg-red-100 text-red-800'
									: task.status === 'pending'
										? 'bg-yellow-100 text-yellow-800'
										: 'bg-gray-100 text-gray-800'}"
                        >
							{task.status}
						</span>
                        <div class="flex gap-2">
                            <Button variant="outline" size="sm">View Details</Button>
                        </div>
                    </div>
                </Card>
            {/each}
        </div>

        <Pagination class="my-6" count={tasks.count} {perPage} bind:currentPage/>
    {/if}
</Container>
