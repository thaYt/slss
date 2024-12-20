<script>
    export let text = "Button";
    export let onClick = () => {};
</script>

<button type="submit" on:click={onClick}>{text}</button>

<style>
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
