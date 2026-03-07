<template>
  <div class="d-flex justify-content-center align-items-center vh-100" style="background: #f8f9fa;">
    <div
      class="p-5 rounded shadow"
      style="min-width: 400px; max-width: 500px; background: #fff; margin-top: -80px;"
    >
      <h2 class="mb-4 text-center">Login / Registrazione</h2>
      <form @submit.prevent="doLogin">
        <div class="mb-3">
          <label class="form-label">Username</label>
          <input v-model="name" class="form-control form-control-lg" required minlength="3" maxlength="16">
        </div>
        <div v-if="isRegister" class="mb-3">
          <label class="form-label">Display Name</label>
          <input v-model="displayName" class="form-control form-control-lg" required>
        </div>
        <div v-if="isRegister" class="mb-3">
          <label class="form-label">Profile Picture URL</label>
          <input v-model="profilePicture" class="form-control form-control-lg" required>
        </div>
        <div class="mb-3 d-flex justify-content-between align-items-center">
          <button type="submit" class="btn btn-primary btn-lg w-50">{{ isRegister ? 'Registrati' : 'Login' }}</button>
          <button type="button" class="btn btn-link" @click="toggleMode">
            {{ isRegister ? 'Vai al Login' : 'Vai alla Registrazione' }}
          </button>
        </div>
        <div v-if="message" :style="{color: error ? 'red' : 'green'}" class="text-center mt-2">{{ message }}</div>
      </form>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      name: "",
      displayName: "",
      profilePicture: "",
      isRegister: false,
      message: "",
      error: false,
    };
  },
  methods: {
    async doLogin() {
      this.message = "";
      this.error = false;
      const payload = this.isRegister
        ? { name: this.name, displayName: this.displayName, profilePicture: this.profilePicture }
        : { name: this.name };
      try {
        const res = await this.$axios.post('/session', payload);
        const data = res.data;
        this.message = data.message || "Successo!";
        this.error = false;
        if (data.user) {
          localStorage.setItem("userId", data.user.id);
          localStorage.setItem("username", data.user.username);
          localStorage.setItem("displayName", data.user.displayName);
          localStorage.setItem("profilePicture", data.user.profilePicture);
          this.$router.push("/home");
        }
      } catch (err) {
        let msg = "Errore di rete";
        if (err.response && err.response.data && err.response.data.message) {
          msg = err.response.data.message;
        }
        this.message = msg;
        this.error = true;
      }
    },
    toggleMode() {
      this.isRegister = !this.isRegister;
      this.message = "";
      this.error = false;
    },
  },
};
</script>