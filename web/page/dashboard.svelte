<script>
    export let user;
    export let files;
</script>

<body>
    <div class="header">
        <h1>Dashboard</h1>
        <a href="/logout">Logout</a>
    </div>

    <main>
        <h2>Welcome, {user.Username}!</h2>
        <p>Here you can upload files to share with others.</p>

        <a href="/upload"><button>Upload File</button></a>

        File list:
        <ul>
            {#each files as file}
                <li>
                    <a href={file.Alias} target="_blank"
                        >{file.Path} ({file.Type}, {file.Size})</a
                    >
                </li>
            {/each}
        </ul>
    </main>
</body>

<style>
    body {
        width: 100%;
        height: 100%;
    }

    .header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 10px 20px;
        background-color: #333;
        color: white;
    }
    h1 {
        margin: 0;
    }
    a,
    p {
        color: white;
        text-decoration: none;
    }

    h2 {
        color: #ddd;
    }


</style>
