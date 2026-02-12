<script>
export default {
	data() {
		return {
			// This helps us show/hide the 'Logout' button
			token: localStorage.getItem('token'),
		}
	},
	methods: {
		logout() {
			localStorage.removeItem('token');
			this.token = null;
			this.$router.push('/login');
		}
	},
	// This "watches" the route to update the token state
	watch: {
		$route() {
			this.token = localStorage.getItem('token');
		}
	}
}
</script>

<template>
	<div>
		<nav class="navbar navbar-dark bg-dark mb-4">
			<div class="container-fluid">
				<span class="navbar-brand mb-0 h1">WASAText</span>
				<div class="d-flex" v-if="token">
					<button class="btn btn-outline-light btn-sm" @click="logout">Logout</button>
				</div>
			</div>
		</nav>

		<main class="container">
			<RouterView />
		</main>
		
		<footer class="text-center mt-5 py-3 border-top text-muted">
			<small>WASA Project - February 2026</small>
		</footer>
	</div>
</template>

<style>
/* No fancy CSS, just making sure the page doesn't look broken */
body {
	background-color: #f8f9fa;
}
</style>