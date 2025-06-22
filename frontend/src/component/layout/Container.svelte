<script lang="ts">
    let className = ''
    export {className as class}
</script>

<div class="max-w-screen-xl m-auto px-10 py-9 {className}">
    <slot></slot>
</div>
