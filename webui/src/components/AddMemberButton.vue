<template>
  <div class="d-inline">
    <button
      class="btn btn-outline-primary btn-sm me-2"
      :disabled="busy"
      title="Aggiungi membri"
      @click="open = true"
    >
      Aggiungi Membri
    </button>

    <div v-if="open" class="amb-backdrop">
      <div class="amb-modal">
        <h6 class="mb-3">Aggiungi membri</h6>

        <input
          v-model="searchTerm"
          class="form-control mb-2"
          placeholder="Cerca utenti..."
          autocomplete="off"
          @input="searchUsers"
        >

        <ul
          v-if="searchResults.length"
          class="list-group mb-2"
          style="max-height:160px; overflow-y:auto;"
        >
          <li
            v-for="u in searchResults"
            :key="u.id"
            class="list-group-item list-group-item-action d-flex align-items-center"
            style="cursor:pointer;"
            @click="addCandidate(u)"
          >
            <img :src="u.profilePicture" class="rounded-circle me-2" width="32" height="32" alt="">
            <span>{{ u.username }}</span>
          </li>
        </ul>

        <div v-if="candidates.length" class="mb-2">
          <span
            v-for="c in candidates"
            :key="c.id"
            class="badge bg-primary me-2 mb-2"
            style="font-size:.9rem;"
          >
            <img :src="c.profilePicture" width="18" height="18" class="rounded-circle me-1" alt="">
            {{ c.username }}
            <span class="ms-1" style="cursor:pointer;" @click="removeCandidate(c.id)">×</span>
          </span>
        </div>

        <div class="d-flex justify-content-end gap-2 mt-3">
          <button class="btn btn-secondary btn-sm" :disabled="busy" @click="close">Chiudi</button>
          <button
            class="btn btn-primary btn-sm"
            :disabled="!candidates.length || busy"
            @click="addGroupMembers"
          >
            Aggiungi
          </button>
        </div>

        <div v-if="error" class="alert alert-danger py-1 px-2 mt-3 mb-0 small">{{ error }}</div>
        <div v-if="success" class="alert alert-success py-1 px-2 mt-3 mb-0 small">{{ success }}</div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'AddMemberButton',
  props: {
    groupId: {
      type: [Number, String],
      required: true
    },
    currentMembers: {
      type: Array,
      default: () => []
    }
  },
  data() {
    return {
      open: false,
      searchTerm: '',
      searchResults: [],
      candidates: [],
      error: '',
      success: '',
      busy: false
    }
  },
  computed: {
    existingUsernames() {
      return this.currentMembers.map(m => m.username);
    }
  },
  methods: {
    async searchUsers() {
      this.error = '';
      this.success = '';
      if (this.searchTerm.trim().length < 1) {
        this.searchResults = [];
        return;
      }
      const auth = localStorage.getItem('userId');
      try {
        const res = await this.$axios.get(`/search/users?q=${encodeURIComponent(this.searchTerm.trim())}`, {
          headers: { Authorization: auth }
        });
        const meUsername = localStorage.getItem('username');
        this.searchResults = res.data.filter(u =>
          u.username !== meUsername &&
          !this.existingUsernames.includes(u.username) &&
            !this.candidates.some(c => c.username === u.username)
        );
      } catch {
        this.searchResults = [];
      }
    },
    addCandidate(user) {
      if (!this.candidates.some(c => c.username === user.username)) {
        this.candidates.push(user);
      }
      this.searchTerm = '';
      this.searchResults = [];
    },
    removeCandidate(username) {
      this.candidates = this.candidates.filter(c => c.username !== username);
    },
    resetState() {
      this.searchTerm = '';
      this.searchResults = [];
      this.candidates = [];
      this.error = '';
      this.success = '';
      this.busy = false;
    },
    close() {
      this.open = false;
      this.resetState();
    },
    // operationId: addGroupMembers
    async addGroupMembers() {
      this.error = '';
      this.success = '';
      if (!this.candidates.length) return;
      this.busy = true;
      const auth = localStorage.getItem('userId');
      try {
        const members = this.candidates.map(c => c.username);
        await this.$axios.patch(`/groups/${this.groupId}/members`, { members }, {
          headers: { Authorization: auth }
        });
        this.success = 'Membri aggiunti';
        this.$emit('members-added', members);
        setTimeout(() => this.close(), 600);
      } catch (e) {
        this.error = e.response?.data?.message || 'Errore aggiunta membri';
      } finally {
        this.busy = false;
      }
    }
  }
}
</script>

<style scoped>
.amb-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,.35);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2100;
}
.amb-modal {
  background: #fff;
  width: 320px;
  border-radius: 12px;
  padding: 16px 18px 18px;
  box-shadow: 0 4px 18px rgba(0,0,0,.15);
  position: relative;
}
</style>