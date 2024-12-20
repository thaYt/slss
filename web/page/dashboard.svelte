<script>
    export let user;
    export let files;

    import "../layout/main.svelte";
    import "../layout/main.css";
    import bytes from "bytes";
</script>

<svelte:head>
    <title>slss &bull; dashboard ({user.Username})</title>
</svelte:head>

<body>
    <div class="header">
        <h1>slss</h1>

        <h2>{user.Username}</h2>

        <a href="/logout">Logout</a>
    </div>

    <main>
        <a href="/upload"><button>Upload File</button></a>

        File list:
        <ul>
            {#each files as file}
                <li>
                    <a href={file.Alias} target="_blank"
                        >{file.Path} ({file.Filetype}, {bytes.format(
                            file.Filesize,
                        )})</a
                    >
                </li>
            {/each}
        </ul>
    </main>
</body>

<style>
    body {
        width: 100vw;
        height: 100vh;
        margin: 0px;
        padding: 0px;
    }

    .header {
        display: flex;
        width: 100vw;
        justify-content: space-between;
        align-items: center;
        background-color: #333;
        color: white;
    }

    h1,
    h2 {
        margin: 0;
    }

    h2 {
        align-self: right;
    }

    a {
        color: white;
        text-decoration: none;
    }

    h2 {
        color: #ddd;
    }
</style>
