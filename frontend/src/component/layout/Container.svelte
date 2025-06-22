<script lang="ts">
	let className = ''
	export { className as class }
</script>

<div class="max-w-screen-[1280px] m-auto px-10 py-9 {className}">
	<slot></slot>
</div>
