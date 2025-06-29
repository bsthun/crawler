<script lang="ts">
	import Navbar from '$/component/navbar/Navbar.svelte'
	import { getContext } from 'svelte'
	import type { Writable } from 'svelte/store'
	import type { Setup } from '$/util/type/setup'

	const setup = getContext<Writable<Setup>>('setup')

	$: if (!$setup.profile.userId) {
		window.location.href = '/entry/login'
	}
</script>

<div>
	<Navbar />
	<div class="pt-16">
		<slot />
	</div>
</div>
