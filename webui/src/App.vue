<script setup>
import { RouterView, useRouter } from 'vue-router'
const router = useRouter();

function goToProfile() {
    router.push('/profile');
}
function logout() {
    localStorage.clear();
    router.push('/');
}
</script>
<script>
export default {}
</script>

<template>
    <header class="navbar navbar-dark sticky-top bg-dark flex-md-nowrap p-0 shadow" style="height: 70px;">
        <span class="navbar-brand col-md-3 col-lg-2 me-0 px-3 fs-3" style="cursor: default; font-size: 2rem;">
            WASA Text
        </span>
    </header>

    <div class="container-fluid">
        <RouterView />
    </div>
</template>

<style>
html, body, #app {
    height: 100%;
    margin: 0;
    padding: 0;
}
body {
    overflow: hidden;
}
.navbar {
    min-height: 70px;
}
</style>
