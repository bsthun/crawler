<script lang="ts">
	import { Link, navigate, useLocation } from 'svelte-navigator'
	import { getContext } from 'svelte'
	import type { Writable } from 'svelte/store'
	import type { Setup } from '$/util/type/setup'
	import { onMount } from 'svelte'
	import { Button } from '$/lib/shadcn/components/ui/button'
	import { LogOut } from 'lucide-svelte'
	import Cookies from 'js-cookie'

	let scrolled = false

	onMount(() => {
		const handleScroll = () => {
			scrolled = window.scrollY > 20
		}

		window.addEventListener('scroll', handleScroll)

		return () => {
			window.removeEventListener('scroll', handleScroll)
		}
	})

	// * handle logout functionality
	const handleLogout = () => {
		Cookies.remove('login')
		navigate('/entry/login')
	}
</script>

<nav
	class="fixed inset-x-0 top-0 z-20 mx-auto flex h-11 max-w-screen-xl items-center justify-between bg-white px-10 py-9"
>
	<div class="flex items-center gap-[18px]">
		<Link to="/">
			<p class="text-[18px] font-medium">Crawler</p>
		</Link>
	</div>

	<Button
		variant="ghost"
		onclick={handleLogout}
	>
		<LogOut size={18} />
	</Button>
</nav>