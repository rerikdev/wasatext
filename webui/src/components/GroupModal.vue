<template>
  <div v-if="openCreateGroupModal" class="modal-backdrop">
    <div class="modal-dialog">
      <div class="modal-content p-4">
        <h5>{{ editGroupMode ? 'Modifica gruppo' : 'Crea nuovo gruppo' }}</h5>
        <form @submit.prevent="editGroupMode ? submitEditGroup() : addToGroup()">
          <div class="mb-3">
            <label class="form-label">Nome gruppo</label>
            <input v-model="newGroupName" class="form-control" required minlength="3" maxlength="16">
          </div>
          <div class="mb-3">
            <label class="form-label">URL foto gruppo (opzionale)</label>
            <input v-model="newGroupPhoto" class="form-control" placeholder="https://...">
          </div>
          <div class="mb-3">
            <label class="form-label">Aggiungi membri</label>
            <div class="dropdown w-100">
              <input
                v-model="searchUser"
                class="form-control mb-2 dropdown-toggle"
                placeholder="Cerca utenti per username..."
                autocomplete="off"
                data-bs-toggle="dropdown"
                @input="searchUsersGroup"
                @focus="dropdownOpenGroup = true"
                @blur="closeDropdownGroup"
              >
              <ul
                class="dropdown-menu w-100"
                :class="{ show: dropdownOpenGroup && searchResultsGroup.length }"
                style="max-height: 200px; overflow-y: auto;"
              >
                <li
                  v-for="user in searchResultsGroup"
                  :key="user.username"
                  class="dropdown-item d-flex align-items-center"
                  @mousedown.prevent="addFirstMembers(user)"
                >
                  <img :src="user.profilePicture" alt="profile" width="32" height="32" class="rounded-circle me-2">
                  <span>{{ user.username }}</span>
                </li>
              </ul>
            </div>
            <div class="mt-2">
              <span
                v-for="user in groupMembers"
                :key="user.username"
                class="badge bg-primary me-2"
                style="font-size:1rem;"
              >
                <img :src="user.profilePicture" alt="profile" width="20" height="20" class="rounded-circle me-1">
                {{ user.username }}
                <span class="ms-1" style="cursor:pointer;" @click="removeMemberFromGroup(user.username)">×</span>
              </span>
            </div>
          </div>
          <div class="d-flex justify-content-end gap-2">
            <button type="button" class="btn btn-secondary" @click="closeCreateGroupModal">Annulla</button>
            <button type="submit" class="btn btn-primary">
              {{ editGroupMode ? 'Salva modifiche' : 'Crea' }}
            </button>
          </div>
          <div v-if="groupError" class="alert alert-danger mt-2">{{ groupError }}</div>
        </form>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'GroupModal',
  props: [
    'openCreateGroupModal',
    'editGroupMode',
    'editGroupId'
  ],
  data() {
    return {
      newGroupName: "",
      newGroupPhoto: "",
      groupMembers: [],
      groupError: "",
      searchUser: "",
      searchResultsGroup: [],
      dropdownOpenGroup: false
    }
  },
  methods: {
    async searchUsersGroup() {
      if (this.searchUser.length < 1) {
        this.searchResultsGroup = [];
        this.dropdownOpenGroup = false;
        return;
      }
      const userId = localStorage.getItem("userId");
      const myUsername = localStorage.getItem("username");
      try {
        const res = await this.$axios.get(`/search/users?q=${encodeURIComponent(this.searchUser)}`, {
          headers: { Authorization: userId }
        });
        const results = res.data;
        this.searchResultsGroup = results.filter(
          u => u.username !== myUsername && !this.groupMembers.some(m => m.username === u.username)
        );
        this.dropdownOpenGroup = !!this.searchResultsGroup.length;
      } catch (e) {
        this.searchResultsGroup = [];
        this.dropdownOpenGroup = false;
      }
    },
    closeDropdownGroup() {
      setTimeout(() => { this.dropdownOpenGroup = false; }, 150);
    },
    addFirstMembers(user) {
      if (!this.groupMembers.some(u => u.username === user.username)) {
        this.groupMembers.push(user);
      }
      this.searchUser = "";
      this.searchResultsGroup = [];
      this.dropdownOpenGroup = false;
    },
    removeMemberFromGroup(username) {
      this.groupMembers = this.groupMembers.filter(u => u.username !== username);
    },
    closeCreateGroupModal() {
      this.$emit('close-create-group-modal');
      this.newGroupName = "";
      this.newGroupPhoto = "";
      this.groupMembers = [];
      this.groupError = "";
      this.searchUser = "";
      this.searchResultsGroup = [];
      this.dropdownOpenGroup = false;
    },
    // operationId: addToGroup (creazione)
    async addToGroup() {
      this.groupError = "";
      if (!this.newGroupName || this.newGroupName.length < 3 || this.newGroupName.length > 16) {
        this.groupError = "Il nome deve essere tra 3 e 16 caratteri";
        return;
      }
      if (this.groupMembers.length === 0) {
        this.groupError = "Aggiungi almeno un membro";
        return;
      }
      try {
        const userId = localStorage.getItem("userId");
        const myUsername = localStorage.getItem("username");
        let membersArr = this.groupMembers.map(u => u.username);
        if (!membersArr.includes(myUsername)) {
          membersArr.push(myUsername);
        }
        const body = {
          name: this.newGroupName,
          members: membersArr
        };
        if (this.newGroupPhoto) body.photo = this.newGroupPhoto;
        await this.$axios.post('/groups', body, {
          headers: { Authorization: userId }
        });
        this.closeCreateGroupModal();
        this.$emit('refresh-groups');
      } catch (e) {
        this.groupError = e.response?.data?.message || "Errore creazione gruppo";
      }
    }
  }
}
</script>