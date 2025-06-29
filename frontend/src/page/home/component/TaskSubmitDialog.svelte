<script lang="ts">
	import { createEventDispatcher } from 'svelte'
	import * as Dialog from '$/lib/shadcn/components/ui/dialog'
	import { Button } from '$/lib/shadcn/components/ui/button'
	import { Input } from '$/lib/shadcn/components/ui/input'
	import { Label } from '$/lib/shadcn/components/ui/label'
	import { backend, catcher } from '$/util/backend.ts'
	import axios from 'axios'
	import type { PayloadTaskCategoryListResponse } from '$/util/backend/backend.ts'
	import { PayloadTaskSubmitRequestTypeEnum } from '$/util/backend/backend.ts'
	import { FileText, Upload, AlertCircle, CheckCircle } from 'lucide-svelte'
	import { toast } from 'svelte-sonner'

	export let categories: PayloadTaskCategoryListResponse
	export let open = false

	const dispatch = createEventDispatcher()

	let selectedCategory = categories[0]
	let selectedType: PayloadTaskSubmitRequestTypeEnum = PayloadTaskSubmitRequestTypeEnum.Web
	let source = ''
	let csvFile: File | null = null
	let fileInput: HTMLInputElement
	let dragOver = false
	let uploading = false
	let uploadProgress = 0
	let uploadStatus: 'idle' | 'uploading' | 'success' | 'error' = 'idle'

	const handleSingleSubmit = () => {
		if (!selectedCategory || !source.trim()) {
			toast.error('Please fill in all required fields')
			return
		}

		backend.task
			.taskSubmit({
				category: selectedCategory,
				type: selectedType,
				source: source.trim(),
			})
			.then((response) => {
				if (response.success) {
					toast.success('Task submitted successfully!')
					resetForm()
					open = false
					dispatch('submitted')
				} else {
					toast.error('Failed to submit task')
				}
			})
			.catch((error) => {
				catcher(error)
			})
	}

	const handleDragOver = (event: DragEvent) => {
		event.preventDefault()
		dragOver = true
	}

	const handleDragLeave = () => {
		dragOver = false
	}

	const handleDrop = (event: DragEvent) => {
		event.preventDefault()
		dragOver = false

		if (event.dataTransfer?.files) {
			const file = event.dataTransfer.files[0]
			if (file && file.type === 'text/csv') {
				csvFile = file
			} else {
				toast.error('Please upload a CSV file')
			}
		}
	}

	const handleFileInputChange = (event: Event) => {
		const input = event.target as HTMLInputElement
		if (input.files && input.files[0]) {
			const file = input.files[0]
			if (file.type === 'text/csv' || file.name.endsWith('.csv')) {
				csvFile = file
			} else {
				toast.error('Please upload a CSV file')
				input.value = ''
			}
		}
	}

	const handleBatchSubmit = () => {
		if (!csvFile) {
			toast.error('Please select a CSV file')
			return
		}

		uploading = true
		uploadStatus = 'uploading'
		uploadProgress = 0

		const formData = new FormData()
		formData.append('file', csvFile)

		axios
			.post('/api/task/submit/batch', formData, {
				headers: {
					'Content-Type': 'multipart/form-data',
				},
				onUploadProgress: (progressEvent) => {
					if (progressEvent.total) {
						uploadProgress = Math.round((progressEvent.loaded * 100) / progressEvent.total)
					}
				},
			})
			.then((response) => {
				uploadStatus = 'success'
				toast.success(`Successfully submitted ${response.data.data.tasksCreated} tasks!`)

				setTimeout(() => {
					resetForm()
					open = false
					dispatch('submitted')
				}, 1500)
			})
			.catch((error) => {
				uploadStatus = 'error'
				toast.error('Failed to upload CSV file')
				catcher(error)
			})
			.finally(() => {
				setTimeout(() => {
					uploading = false
					uploadStatus = 'idle'
					uploadProgress = 0
				}, 2000)
			})
	}

	const resetForm = () => {
		selectedCategory = ''
		selectedType = PayloadTaskSubmitRequestTypeEnum.Web
		source = ''
		csvFile = null
		if (fileInput) {
			fileInput.value = ''
		}
		uploadStatus = 'idle'
		uploadProgress = 0
	}

	$: statusIcon = uploadStatus === 'success' ? CheckCircle : uploadStatus === 'error' ? AlertCircle : FileText
	$: borderClasses = dragOver ? 'border-blue-500 bg-blue-50' : 'border-gray-300 border-dashed'
</script>

<Dialog.Root bind:open>
	<Dialog.Trigger>
		<slot />
	</Dialog.Trigger>
	<Dialog.Content class="max-w-4xl">
		<Dialog.Header>
			<Dialog.Title>Submit Tasks</Dialog.Title>
			<Dialog.Description>Submit individual tasks or upload a CSV file for batch processing</Dialog.Description>
		</Dialog.Header>

		<div class="mt-4 flex gap-6">
			<div class="flex-1 space-y-4">
				<h3 class="text-lg font-semibold">Individual Task</h3>

				<div class="space-y-2">
					<Label for="category">Category</Label>
					<select
						bind:value={selectedCategory}
						class="w-full rounded-md border px-3 py-2"
						id="category"
						required
					>
						{#if categories?.categories}
							{#each categories.categories as category}
								<option value={category.name}>{category.name}</option>
							{/each}
						{/if}
					</select>
				</div>

				<div class="space-y-2">
					<Label for="type">Type</Label>
					<select bind:value={selectedType} class="w-full rounded-md border px-3 py-2" id="type" required>
						<option value="web">Web</option>
						<option value="doc">PDF</option>
						<option value="youtube">YouTube</option>
					</select>
				</div>

				<div class="space-y-2">
					<Label for="source">Source URL</Label>
					<Input bind:value={source} id="source" placeholder="https://example.com" required type="url" />
				</div>

				<Button class="w-full" disabled={!selectedCategory || !source.trim()} onclick={handleSingleSubmit}>
					Submit Task
				</Button>
			</div>

			<div class="flex-1 space-y-4">
				<h3 class="text-lg font-semibold">Batch Task</h3>
				<div
					class="flex flex-col items-center justify-center rounded-lg border-2 p-6 transition-colors {borderClasses}"
					on:dragleave={handleDragLeave}
					on:dragover={handleDragOver}
					on:drop={handleDrop}
					role="application"
				>
					<div class="flex flex-col items-center space-y-4">
						<svelte:component class="h-12 w-12 text-gray-400" this={statusIcon} />

						{#if csvFile}
							<div class="pb-1.5 text-center">
								<p class="text-sm font-medium text-green-600">
									{csvFile.name}
								</p>
								<p class="text-xs text-gray-500">
									{(csvFile.size / 1024).toFixed(1)} KB
								</p>
							</div>
						{:else}
							<div class="text-center">
								<p class="pb-2.5 text-xs font-medium text-gray-700">
									Drop CSV file here or click to browse
								</p>
							</div>
						{/if}

						{#if uploadStatus === 'uploading'}
							<div class="w-full max-w-xs">
								<div class="w-full rounded-full bg-gray-200">
									<div
										class="h-2 rounded-full bg-blue-500 transition-all duration-300"
										style="width: {uploadProgress}%"
									></div>
								</div>
							</div>
						{/if}

						<Button disabled={uploading} onclick={() => fileInput.click()} variant="outline">
							<Upload class="mr-2 h-4 w-4" />
							Browse Files
						</Button>
					</div>
				</div>

				<Button class="w-full" disabled={!csvFile || uploading} onclick={handleBatchSubmit}>
					{uploading ? 'Uploading...' : 'Upload CSV'}
				</Button>

				<input
					accept=".csv"
					bind:this={fileInput}
					class="hidden"
					on:change={handleFileInputChange}
					type="file"
				/>
			</div>
		</div>
	</Dialog.Content>
</Dialog.Root>
