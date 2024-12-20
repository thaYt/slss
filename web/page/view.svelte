<script>
    import bytes from "bytes";

    export let file;
    export let site;

    let name = file.Alias;
    let type = file.Filetype;
    let size = file.Filesize;
    let originalName = file.Path;
    let src = name + "/raw";
</script>

<svelte:head>
    <title>{originalName} &bull; slss</title>
    <meta name="theme-color" content="#8800aa" />
    <meta property="og:title" content={originalName} />
    <meta property="og:description" content="{type} - {bytes.format(size)}" />
    {#if type.startsWith("image")}
        <meta property="og:type" content="image" />
        <meta property="og:image" content={site + "/" + src} />
    {:else if type.startsWith("video")}
        <meta property="og:type" content="video" />
        <meta property="og:image" content={site + "/" + src} />
    {/if}
</svelte:head>

<div class="panel">
    <h1 id="name">{originalName}</h1>
    <p>{bytes.format(size)}</p>
    {#if type.startsWith("image")}
        <img {src} alt={name} class="responsive-media" />
    {:else if type.startsWith("audio")}
        <audio controls class="responsive-media">
            <source {src} {type} />
        </audio>
    {:else if type.startsWith("video")}
        <!-- svelte-ignore a11y-media-has-caption -->
        <!-- i don't expect people to include a separate caption file for their video -->
        <video controls class="responsive-media">
            <source {src} {type} />
        </video>
    {:else}
        <p>No preview available for {type}</p>
    {/if}

    <a href={src} target="_blank" rel="noopener noreferrer"> Download </a>
</div>

<style>
    #name {
        font-size: 2em;
        margin-bottom: 5px;
    }

    .panel {
        border: 1px solid #333;
        border-radius: 5px;
        background-color: #222;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        text-align: center;
        padding: 20px;
        color: white;
        overflow: hidden;
    }

    .responsive-media {
        margin: 20px 0;
        object-fit: contain;
    }

    @media (orientation: landscape) {
        .responsive-media {
            max-width: 60%;
            max-height: 60%;
        }
    }

    @media (orientation: portrait) {
        .responsive-media {
            max-width: 80%;
            max-height: 80%;
        }

        .panel {
            width: fit-content;
        }
    }

    video {
        align-items: center;
        justify-content: center;

        margin: 0;
    }

    audio {
        min-width: 250px;
    }
</style>
