<script>
export default {
  data() {
    return {
      username: "",
      errormsg: null,
      loading: false
    };
  },
  methods: {
    async login() {
      // Validate input
      if (!this.username.trim()) {
        this.errormsg = "Username is required";
        return;
      }

      this.errormsg = null;
      this.loading = true;

      try {
        // Send username to your backend
        const response = await this.$axios.post("/session", {
          name: this.username
        });

        // 🚨 CRITICAL FIX: Your Go backend sends {"identifier": "123"}
        // We must access .identifier, otherwise we get "[object Object]"
        const userId = response.data.identifier;

        // ✅ Store everything properly in localStorage
        localStorage.setItem("token", String(userId));   
        localStorage.setItem("userId", String(userId));  
        localStorage.setItem("username", this.username); 

        // ✅ Redirect to the home page
        this.$router.push("/");
      } catch (e) {
        this.errormsg = e.response?.data?.message || e.message || "Login failed";
      } finally {
        this.loading = false;
      }
    }
  }
};
</script>

<template>
  <div class="container mt-5" style="max-width: 400px;">
    <div class="card shadow">
      <div class="card-body">
        <h2 class="card-title text-center mb-4">Sign In</h2>
        
        <div v-if="errormsg" class="alert alert-danger">
          {{ errormsg }}
        </div>

        <form @submit.prevent="login">
          <div class="mb-3">
            <label for="username" class="form-label">Username</label>
            <input
              id="username"
              v-model="username"
              type="text"
              class="form-control"
              placeholder="Enter your username"
              required
              autofocus
              autocomplete="username"
            />
          </div>

          <button 
            type="submit" 
            class="btn btn-primary w-100" 
            :disabled="loading"
          >
            <span 
              v-if="loading" 
              class="spinner-border spinner-border-sm me-2" 
              role="status" 
              aria-hidden="true"
            ></span>
            {{ loading ? "Signing in..." : "Sign In" }}
          </button>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
.container {
  min-height: 80vh;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>