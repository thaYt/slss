<script>
    export let name;
</script>

<!-- svelte-ignore a11y-media-has-caption -->
<video controls id="video">
    <source src={"http://localhost:8000/raw/" + name} type="video/mp4">
</video>

<style>
    #video {
        height: 60%;
    }
</style>
