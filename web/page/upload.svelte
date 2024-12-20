<script>
    import Button from "./lib/Button.svelte";

    let file;
    export let user;
    export let maxFilesize;

    async function handleSubmit(event) {
        event.preventDefault();
        const formData = new FormData();
        formData.append("file", file[0]);

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
                if (data.error) {
                    alert(data.error);
                }
            } else {
                console.error("Error:", response);
            }
        } catch (error) {
            console.error("Error:", error);
            alert("An error occurred while uploading the file.");
        }
    }
</script>

<svelte:head>
    <title>slss &bull; upload</title>
</svelte:head>

<p>max filesize: {maxFilesize}mb</p>

<form on:submit|preventDefault={handleSubmit}>
    <input type="file" bind:files={file} />
    <Button type="submit" text="Upload"></Button>
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
</style>
