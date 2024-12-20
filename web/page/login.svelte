<script>
    import Button from "./lib/Button.svelte";

    let username = "";
    let password = "";

    let error = "";
    function handleSubmit() {
        fetch("/login", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({ username: username, password: password }),
        })
            .then((res) => {
                if (!res.ok) {
                    throw new Error(`HTTP error! status: ${res.status}`);
                }
                return res.json();
            })
            .then(() => {
                const urlParams = new URLSearchParams(window.location.search);
                const returnUrl = urlParams.get("return");
                window.location.href = returnUrl ? returnUrl : "/dashboard";
            })
            .catch((err) => {
                if (err.status === 401) {
                    error = "invalid username or password";
                } else {
                    error = "an error occurred while logging in";
                }
            });
    }
</script>

<svelte:head>
    <title>slss &bull; login</title>
</svelte:head>

<body>
    {#if error}
        <div
            class="fail"
            role="button"
            tabindex="0"
            on:click={() => (error = "")}
            on:keypress={() => (error = "")}
        >
            <div class="pane">
                <p>{error}</p>
                <Button text="Okay" on:click={() => (error = "")} />
            </div>
        </div>
    {:else}
        <h1>Login</h1>
        <form on:submit|preventDefault={handleSubmit}>
            <label for="username">Username:</label>
            <input type="text" id="username" bind:value={username} />
            <label for="password">Password:</label>
            <input type="password" id="password" bind:value={password} />
            <Button type="submit" text="Login"></Button>
        </form>
    {/if}
</body>

<style>
    body {
        width: 100vw;
        height: 100vh;
        margin: 0px;
        padding: 0px;
    }

    h1,
    p,
    label {
        text-align: center;
        color: #ddd;
    }

    form {
        display: flex;
        flex-direction: column;
        align-items: center;
    }

    .fail {
        width: 100vw;
        height: 100vh;
        top: 0px;
        left: 0px;
        background-color: #0005;
        display: flex;
        justify-content: center;
        align-items: center;
        margin: 0;
        padding: 0;
    }

    .pane {
        background-color: #333;
        padding: 20px;
        border-radius: 5px;
    }

    input {
        background-color: #3333;
        color: #ddd;
        border-color: #333;
        border-style: solid;
        border-radius: 5px;
        padding: 5px;
        margin-bottom: 10px;
    }
</style>
