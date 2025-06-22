<script lang="ts">
	let className = ''
	export { className as class }
</script>

<div class="m-auto max-w-screen-xl px-10 py-9 {className}">
	<slot></slot>
</div>
