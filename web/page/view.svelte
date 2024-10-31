<script>
    import Plyr from "plyr";

    const player = new Plyr("#player");
    export let fileId;

    function logEvent(event) {
        console.log(event);
    }
</script>

<div class="youtube-plyr">
    <button on:click={() => player.play()}>PLAY</button>
    <button on:click={() => player.pause()}>PAUSE</button>
</div>

<style>
    .youtube-plyr {
        max-width: 800px;
    }
</style>
