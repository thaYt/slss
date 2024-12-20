<script>
    import Head from "./head.svelte";
    import "./main.css";
</script>

<Head />

<main>
    <slot />
</main>

<style>
    main {
        background-color: #161616;
        margin: 0px;
        padding: 0px;
        position: absolute;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
    }
</style>
