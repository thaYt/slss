<script>
    let file;
    export let user;

    async function handleSubmit(event) {
        event.preventDefault();
        const formData = new FormData();
        formData.append("file", file[0]);
        console.log("Uploading file:", file);

        try {
            const response = await fetch("/upload", {
                method: "POST",
                headers: {
                    Authorization: user.Token,
                },
                body: formData,
            });

            if (response.ok) {
                const data = await response.json();
                window.location.href = data.url;
            } else {
                alert("Unauthorized.");
            }
        } catch (error) {
            console.error("Error:", error);
            alert("An error occurred while uploading the file.");
        }
    }
</script>

<form on:submit={handleSubmit}>
    <input type="file" bind:files={file} />
    <button type="submit">Upload</button>
</form>

<style>
    form {
        display: flex;
        flex-direction: column;
        align-items: center;
    }
    input[type="file"] {
        margin-bottom: 10px;
    }
    button {
        padding: 10px 20px;
        background-color: #333;
        color: white;
        border: none;
        border-radius: 5px;
        cursor: pointer;
    }
    button:hover {
        background-color: #555;
    }
    button:active {
        background-color: #777;
    }
    button:focus {
        outline: none;
    }
</style>