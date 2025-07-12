<script lang="ts">
	import {
		Dialog,
		DialogContent,
		DialogHeader,
		DialogTitle,
		DialogTrigger,
	} from '$/lib/shadcn/components/ui/dialog'
	import { Loader2Icon } from 'lucide-svelte'
	import { backend, catcher } from '$/util/backend.ts'
	import type { PayloadTaskDetailResponse } from '$/util/backend/backend.ts'
	import { createEventDispatcher } from 'svelte'

	export let open = false
	export let taskId: number

	let taskDetail: PayloadTaskDetailResponse | null = null
	let loading = false

	const dispatch = createEventDispatcher<{
		openTask: { taskId: string }
	}>()

	const formatDate = (dateString: string) => {
		return new Date(dateString).toLocaleDateString()
	}

	const formatTime = (dateString: string) => {
		return new Date(dateString).toLocaleTimeString([], {
			hour: '2-digit',
			minute: '2-digit',
		})
	}

	const loadTaskDetail = async () => {
		if (!taskId) return

		loading = true
		try {
			const res = await backend.task.taskDetail({ taskId })
			if (res.success && res.data) {
				taskDetail = res.data
			}
		} catch (err) {
			catcher(err)
		} finally {
			loading = false
		}
	}

	const parseTaskIdReferences = (text: string) => {
		// Regex to match whitespace + # + 11 alphanumeric characters + whitespace
		const regex = /(\s)(#[a-zA-Z0-9]{11})(\s)/g
		const parts = []
		let lastIndex = 0
		let match

		while ((match = regex.exec(text)) !== null) {
			// Add text before the match
			if (match.index > lastIndex) {
				parts.push({ type: 'text', content: text.slice(lastIndex, match.index) })
			}

			// Add the whitespace before #
			parts.push({ type: 'text', content: match[1] })

			// Add the task ID as clickable
			parts.push({ 
				type: 'taskId', 
				content: match[2],
				id: match[2].slice(1) // Remove the # to get just the ID
			})

			// Add the whitespace after
			parts.push({ type: 'text', content: match[3] })

			lastIndex = regex.lastIndex
		}

		// Add remaining text
		if (lastIndex < text.length) {
			parts.push({ type: 'text', content: text.slice(lastIndex) })
		}

		return parts
	}

	const openNestedTaskDialog = (taskIdStr: string) => {
		dispatch('openTask', { taskId: taskIdStr })
	}

	$: if (open && taskId) {
		loadTaskDetail()
	}
</script>

<Dialog bind:open>
	<DialogTrigger>
		<slot />
	</DialogTrigger>
	<DialogContent class="w-4xl">
		<DialogHeader>
			<DialogTitle>#{taskId}</DialogTitle>
		</DialogHeader>

		{#if loading}
			<div class="flex min-h-[400px] items-center justify-center">
				<Loader2Icon class="text-primary h-8 w-8 animate-spin" />
			</div>
		{:else if taskDetail}
			<!-- Task Info Table -->
			<div class="space-y-4">
				<table class="w-full text-sm">
					<tbody>
					<tr class="border-b">
						<td class="py-2 pr-4 font-medium">Status</td>
						<td class="py-2 truncate">
                                <span
									class="rounded-full px-2 py-1 text-xs {taskDetail.status === 'completed'
                                        ? 'bg-green-100 text-green-800'
                                        : taskDetail.status === 'failed'
                                            ? 'bg-red-100 text-red-800'
                                            : taskDetail.status === 'pending'
                                                ? 'bg-yellow-100 text-yellow-800'
                                                : 'bg-gray-100 text-gray-800'}"
								>
                                    {taskDetail.status}
                                </span>
						</td>
					</tr>
					<tr class="border-b">
						<td class="py-2 pr-4 font-medium">Type</td>
						<td class="py-2 truncate">{taskDetail.type}</td>
					</tr>
					<tr class="border-b">
						<td class="py-2 pr-4 font-medium">Tokens</td>
						<td class="py-2 truncate">{taskDetail.tokenCount}</td>
					</tr>
					{#if taskDetail.uploadId}
						<tr class="border-b">
							<td class="py-2 pr-4 font-medium">Upload</td>
							<td class="py-2 truncate">#{taskDetail.uploadId}</td>
						</tr>
					{/if}
					{#if taskDetail.title}
						<tr class="border-b  ">
							<td class="py-2 pr-4 font-medium">Title</td>
							<td class="py-2 max-w-sm truncate" title={taskDetail.title}>{taskDetail.title}</td>
						</tr>
					{/if}
					{#if taskDetail.source}
						<tr class="border-b">
							<td class="py-2 pr-4 font-medium">Source</td>
							<td class="py-2 max-w-sm truncate">
								<a
									href={taskDetail.source}
									target="_blank"
									rel="noopener noreferrer"
									class="text-blue-600 hover:text-blue-800"
									title={taskDetail.source}
								>
									{taskDetail.source}
								</a>
							</td>
						</tr>
					{/if}
					<tr class="border-b">
						<td class="py-2 pr-4 font-medium">Created</td>
						<td class="py-2 truncate">
							{formatDate(taskDetail.createdAt)} {formatTime(taskDetail.createdAt)}
						</td>
					</tr>
					<tr class="border-b">
						<td class="py-2 pr-4 font-medium">Updated</td>
						<td class="py-2 truncate">
							{formatDate(taskDetail.updatedAt)} {formatTime(taskDetail.updatedAt)}
						</td>
					</tr>
					{#if taskDetail.failedReason}
						<tr class="border-b">
							<td class="py-2 pr-4 font-medium text-red-600">Failed:</td>
							<td class="py-2 max-w-sm text-red-600" title={taskDetail.failedReason}>
								<div class="break-words">
									{#each parseTaskIdReferences(taskDetail.failedReason) as part}
										{#if part.type === 'taskId'}
											<button
												type="button"
												class="hover:text-red-800 hover:underline cursor-pointer font-medium bg-transparent border-none p-0 m-0 font-inherit text-inherit inline"
												title="Click to view task {part.id}"
												onclick={() => openNestedTaskDialog(part.id || '')}
											>
												{part.content}
											</button>
										{:else}
											{part.content}
										{/if}
									{/each}
								</div>
							</td>
						</tr>
					{/if}
					</tbody>
				</table>

				<!-- Content -->
				<div>
					<div class="mb-2 font-medium">Content</div>
					<div class="h-[400px]  overflow-auto rounded-md border bg-gray-50 ">
						<pre class="p-4 text-sm whitespace-pre max-w-md">{taskDetail.content}</pre>
					</div>
				</div>
			</div>
		{:else}
			<div class="text-muted-foreground text-center">No task details available</div>
		{/if}
	</DialogContent>
</Dialog>