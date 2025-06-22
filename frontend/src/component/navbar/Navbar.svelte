<script lang="ts">
	import { navigate, useLocation } from 'svelte-navigator'
	import { getContext } from 'svelte'
	import type { Writable } from 'svelte/store'
	import type { Setup } from '$/util/type/setup'
	import { onMount } from 'svelte'
	import connectedLogo from '$/assets/connected-logo-sm.png'
	import forumActive from '$/assets/forum-active.png'
	import forumInactive from '$/assets/forum-inactive.png'
	import folderActive from '$/assets/folder-active.png'
	import folderInactive from '$/assets/folder-inactive.png'
	import chipExtraction from '$/assets/chip_extraction-icon.png'
	import { Button } from '$/lib/shadcn/components/ui/button'

	let scrolled = false

	const setup = getContext<Writable<Setup>>('setup')
	const location = useLocation()

	// * reactive statements to update page state when location changes
	$: isChatPage = $location.pathname === '/chat'
	$: isRepoPage = $location.pathname === '/repo'

	onMount(() => {
		const handleScroll = () => {
			scrolled = window.scrollY > 20
		}

		window.addEventListener('scroll', handleScroll)

		return () => {
			window.removeEventListener('scroll', handleScroll)
		}
	})
</script>

<nav
	class="fixed inset-x-0 top-0 z-20 mx-auto flex h-11 max-w-screen-xl items-center justify-between bg-white px-10 py-9"
>
	<div class="flex items-center gap-[18px]">
		<img src={connectedLogo} alt="Connected Logo" class="h-8" />
		<p class="text-[18px] font-medium">ChatBotDocument</p>
	</div>
	<div class="flex gap-3">
		<Button
			class={`h-8 w-8 cursor-pointer rounded-full border-2 p-0 ${isChatPage ? 'bg-black' : 'bg-white'}`}
			onclick={() => {
				navigate('/chat')
			}}
		>
			<div class="flex h-full w-full items-center justify-center">
				<img src={isChatPage ? forumActive : forumInactive} alt="Forum Icon" class="h-5 w-5 object-contain" />
			</div>
		</Button>
		<Button
			class={`h-8 w-8 cursor-pointer rounded-full border-2 p-0 ${isRepoPage ? 'bg-black' : 'bg-white'}`}
			onclick={() => {
				navigate('/repo')
			}}
		>
			<div class="flex h-full w-full items-center justify-center">
				<img
					src={isRepoPage ? folderActive : folderInactive}
					alt="Repository Icon"
					class="h-5 w-5 object-contain"
				/>
			</div>
		</Button>
		<Button
			class="h-8 w-8 cursor-pointer rounded-full border-2 bg-white p-0"
			onclick={() => {
				// * clear mock cookie
				document.cookie = 'login=; path=/; max-age=0'

				// * reset setup store
				setup.set({
					profile: {},
					initialized: true,
					reload: async () => {},
				})

				// * redirect to login
				navigate('/entry/login')
			}}
		>
			<div class="flex h-full w-full items-center justify-center">
				<img src={chipExtraction} alt="Logout Icon" class="h-4 w-4 object-contain" />
			</div>
		</Button>
	</div>
</nav>
