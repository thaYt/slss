<script>
    let username = "";
    let password = "";

    function handleSubmit() {
        console.log("Requesting login with username:", username);
        // post request to login endpoint
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
            .then((data) => {
                if (data.success) {
                    console.log("Login successful");
                    // window.location.href = "/dashboard";
                    window.location.href = "/sharex-config";
                } else {
                    alert(data.error);
                }
            })
            .catch((err) => {
                console.error("Network error:", err);
                alert(
                    "An error occurred while trying to log in. Please try again later.",
                );
            });
    }
</script>

<main>
    <h1>Login</h1>
    <form on:submit|preventDefault={handleSubmit}>
        <label for="username">Username:</label>
        <input type="text" id="username" bind:value={username} />
        <label for="password">Password:</label>
        <input type="password" id="password" bind:value={password} />
        <button type="submit">Login</button>
    </form>
</main>

<style>
    h1 {
        text-align: center;
        color: #ddd;
    }

    form {
        display: flex;
        flex-direction: column;
        align-items: center;
    }
</style>