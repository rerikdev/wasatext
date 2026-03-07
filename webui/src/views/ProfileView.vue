<template>
  <div class="d-flex justify-content-center align-items-center vh-100" style="background: #f8f9fa;">
    <div class="p-5 rounded shadow bg-white" style="min-width: 500px; max-width: 600px;">
      <h2 class="mb-5 text-center display-4">Il tuo profilo</h2>
      <form v-if="user" class="text-center" @submit.prevent>
        <div class="d-flex flex-column align-items-center mb-4">
          <img :src="user.profilePicture" alt="Profile" width="150" height="150" class="rounded-circle mb-4 shadow">
          <div class="fs-2 fw-bold mb-2">{{ user.username }}</div>
        </div>
        <div class="mb-4">
          <label class="form-label mb-1 fs-4">Display Name</label>
          <div class="fs-3">{{ user.displayName }}</div>
        </div>

        <!-- Mini sezione cambio username -->
        <div class="mb-4">
          <label class="form-label fs-5">Cambia Username</label>
          <div class="input-group">
            <input v-model="newUsername" class="form-control form-control-lg" placeholder="Nuovo username">
            <button type="button" class="btn btn-primary" @click="setMyUserName">Cambia</button>
          </div>
        </div>

        <!-- Mini sezione cambio foto profilo -->
        <div class="mb-4">
          <label class="form-label fs-5">Cambia Immagine Profilo</label>
          <div class="input-group">
            <input v-model="newProfilePicture" class="form-control form-control-lg" placeholder="Nuovo URL immagine">
            <button type="button" class="btn btn-primary" @click="setMyPhoto">Cambia</button>
          </div>
        </div>

        <div v-if="message" class="alert mt-3" :class="{'alert-success': !error, 'alert-danger': error}">
          {{ message }}
        </div>
      </form>
      <div v-else class="text-center fs-3">
        <p>Caricamento...</p>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      user: null,
      message: "",
      error: false,
      newProfilePicture: "",
      newUsername: ""
    }
  },
  async mounted() {
    await this.getUser();
  },
  methods: {
    async getUser() {
      const userId = localStorage.getItem("userId");
      try {
        const res = await this.$axios.get(`/users/${userId}`, {
          headers: { Authorization: userId }
        });
        this.user = res.data;
      } catch (err) {
        this.user = null;
        this.message = "Utente non trovato, effettua di nuovo il login.";
        this.error = true;
      }
    },
    async setMyPhoto() {
      this.message = "";
      this.error = false;
      if (!this.newProfilePicture) {
        this.message = "Inserisci un URL valido";
        this.error = true;
        return;
      }
      const userId = localStorage.getItem("userId");
      try {
        const res = await this.$axios.patch(`/users/${userId}/photo`, 
          { photoUrl: this.newProfilePicture },
          { headers: { Authorization: userId } }
        );
        this.message = "Immagine profilo aggiornata!";
        this.user.profilePicture = this.newProfilePicture;
        this.newProfilePicture = "";
      } catch (err) {
        this.message = err.response?.data?.message || "Errore";
        this.error = true;
      }
    },
    async setMyUserName() {
      this.message = "";
      this.error = false;
      if (!this.newUsername || this.newUsername.length < 3) {
        this.message = "Username troppo corto";
        this.error = true;
        return;
      }
      if (this.newUsername === this.user.username) {
        this.message = "Hai già questo username";
        this.error = true;
        return;
      }
      const userId = localStorage.getItem("userId");
      try {
        const res = await this.$axios.patch(`/users/${userId}`, 
          { newName: this.newUsername },
          { headers: { Authorization: userId } }
        );
        this.message = "Username aggiornato!";
        localStorage.setItem("username", this.newUsername);
        this.user.username = this.newUsername;
        this.newUsername = "";
        await this.getUser();
      } catch (err) {
        this.message = err.response?.data?.message || "Errore";
        this.error = true;
      }
    }
  }
}
</script>