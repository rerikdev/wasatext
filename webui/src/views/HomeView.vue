<script>
export default {
  data() {
    return {
      errormsg: null,
      loading: false,
      conversations: [],
      polling: null,
      searchQuery: "",
      searchResults: [],
      myUsername: localStorage.getItem("username"),
      
      // Group creation modal
      showCreateGroup: false,
      groupName: "",
      groupPhoto: null,
      groupPhotoPreview: null,
      selectedMembers: [],
      userSearchQuery: "",
      userSearchResults: [],

      // Profile edit modal
      showProfileModal: false,
      profileName: "",
      profilePhoto: null,
      profilePhotoPreview: null
    };
  },
  methods: {
    checkAuth() {
      const token = localStorage.getItem("token");
      if (!token) {
        this.$router.push("/login");
        return false;
      }
      return true;
    },

    async refresh(isSilent = false) {
      if (!this.checkAuth()) return;
      if (!isSilent) this.loading = true;
      this.errormsg = null;
      try {
        const response = await this.$axios.get("/users/me/conversations", {
          headers: {
            Authorization: "Bearer " + localStorage.getItem("token")
          }
        });
        this.conversations = response.data;
      } catch (e) {
        if (e.response?.status === 401) {
          this.logout();
        } else {
          if (!isSilent) this.errormsg = "Backend error: " + e.toString();
        }
      }
      if (!isSilent) this.loading = false;
    },

    async searchUsers() {
      if (!this.checkAuth()) return;
      if (this.searchQuery.length < 2) {
        this.searchResults = [];
        return;
      }
      try {
        const response = await this.$axios.get("/users", {
          params: { username: this.searchQuery },
          headers: { Authorization: "Bearer " + localStorage.getItem("token") }
        });
        this.searchResults = response.data;
      } catch (e) {
        this.errormsg = "Search failed: " + e.toString();
      }
    },

    startConversation(username) {
      this.$router.push({ path: "/conversations/" + username });
      this.searchQuery = "";
      this.searchResults = [];
    },

    selectConversation(id) {
      this.$router.push({ path: "/conversations/" + id });
    },

    async deleteConversation(userId) {
      if (!confirm("Delete this entire conversation? This cannot be undone.")) return;
      try {
        await this.$axios.delete(`/conversations/${userId}`, {
          headers: { Authorization: "Bearer " + localStorage.getItem("token") }
        });
        this.refresh();
        if (this.$route.path === `/conversations/${userId}`) {
          this.$router.push("/");
        }
      } catch (e) {
        this.errormsg = "Failed to delete chat: " + e.toString();
      }
    },

    // ------------------------------------------------------------
    // GROUP CREATION (unchanged)
    // ------------------------------------------------------------
    openCreateGroupModal() {
      this.showCreateGroup = true;
      this.groupName = "";
      this.groupPhoto = null;
      this.groupPhotoPreview = null;
      this.selectedMembers = [];
      this.userSearchQuery = "";
      this.userSearchResults = [];
    },

    async searchUsersForGroup() {
      if (this.userSearchQuery.length < 2) {
        this.userSearchResults = [];
        return;
      }
      try {
        const response = await this.$axios.get("/users", {
          params: { username: this.userSearchQuery },
          headers: { Authorization: "Bearer " + localStorage.getItem("token") }
        });
        this.userSearchResults = response.data.filter(
          u => !this.selectedMembers.includes(u) && u !== this.myUsername
        );
      } catch (e) {
        this.errormsg = "Search failed: " + e.toString();
      }
    },

    addMember(username) {
      if (!this.selectedMembers.includes(username)) {
        this.selectedMembers.push(username);
      }
      this.userSearchQuery = "";
      this.userSearchResults = [];
    },

    removeMember(username) {
      this.selectedMembers = this.selectedMembers.filter(u => u !== username);
    },

    onGroupPhotoSelected(event) {
      const file = event.target.files[0];
      if (file) {
        this.groupPhoto = file;
        this.groupPhotoPreview = URL.createObjectURL(file);
      }
    },

    clearGroupPhoto() {
      if (this.groupPhotoPreview) {
        URL.revokeObjectURL(this.groupPhotoPreview);
      }
      this.groupPhoto = null;
      this.groupPhotoPreview = null;
    },

    async createGroup() {
      if (!this.groupName.trim()) {
        this.errormsg = "Group name is required";
        return;
      }
      if (this.selectedMembers.length === 0) {
        this.errormsg = "Add at least one member";
        return;
      }

      const formData = new FormData();
      formData.append("type", "group");
      formData.append("name", this.groupName);
      if (this.groupPhoto) {
        formData.append("photo", this.groupPhoto);
      }
      formData.append("members", JSON.stringify(this.selectedMembers));

      try {
        const response = await this.$axios.post("/conversations", formData, {
          headers: {
            Authorization: "Bearer " + localStorage.getItem("token"),
            "Content-Type": "multipart/form-data"
          }
        });
        const convId = response.data.id;
        this.showCreateGroup = false;
        this.refresh();
        this.$router.push(`/conversations/${convId}`);
      } catch (e) {
        this.errormsg = "Failed to create group: " + e.toString();
      }
    },

    // ------------------------------------------------------------
    // PROFILE EDITING
    // ------------------------------------------------------------
    openProfileModal() {
      this.profileName = this.myUsername;
      this.profilePhoto = null;
      this.profilePhotoPreview = null;
      this.showProfileModal = true;
    },

    onProfilePhotoSelected(event) {
      const file = event.target.files[0];
      if (file) {
        this.profilePhoto = file;
        this.profilePhotoPreview = URL.createObjectURL(file);
      }
    },

    clearProfilePhoto() {
      if (this.profilePhotoPreview) {
        URL.revokeObjectURL(this.profilePhotoPreview);
      }
      this.profilePhoto = null;
      this.profilePhotoPreview = null;
    },

    async updateProfile() {
      if (!this.profileName.trim()) {
        this.errormsg = "Name cannot be empty";
        return;
      }
      try {
        // Update name
        await this.$axios.put("/users/me/name", { name: this.profileName }, {
          headers: { Authorization: "Bearer " + localStorage.getItem("token") }
        });

        // Update photo if changed
        if (this.profilePhoto) {
          const formData = new FormData();
          formData.append("photo", this.profilePhoto);
          await this.$axios.put("/users/me/photo", formData, {
            headers: {
              Authorization: "Bearer " + localStorage.getItem("token"),
              "Content-Type": "multipart/form-data"
            }
          });
        }

        // Update localStorage
        localStorage.setItem("username", this.profileName);
        this.myUsername = this.profileName;
        this.showProfileModal = false;
        this.errormsg = null;

        // Refresh to update displayed name everywhere
        this.refresh();
      } catch (e) {
        if (e.response?.status === 409) {
          this.errormsg = "Username already taken";
        } else {
          this.errormsg = "Failed to update profile: " + e.toString();
        }
      }
    },

    logout() {
      localStorage.clear();
      this.$router.push("/login");
    }
  },
  mounted() {
    if (this.checkAuth()) {
      this.refresh();
      this.polling = setInterval(() => {
        if (this.searchQuery.length === 0) {
          this.refresh(true);
        }
      }, 5000);
    }
  },
  beforeUnmount() {
    if (this.polling) clearInterval(this.polling);
    if (this.groupPhotoPreview) URL.revokeObjectURL(this.groupPhotoPreview);
    if (this.profilePhotoPreview) URL.revokeObjectURL(this.profilePhotoPreview);
  }
};
</script>

<template>
  <div class="container-fluid">
    <!-- Header -->
    <div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
      <h1 class="h2">
        <span v-if="myUsername">Welcome, {{ myUsername }}!</span>
        <span v-else>My Conversations</span>
      </h1>
      <div class="btn-toolbar mb-2 mb-md-0">
        <div class="btn-group me-2">
          <button type="button" class="btn btn-sm btn-outline-secondary" @click="refresh(false)">
            Refresh
          </button>
          <button type="button" class="btn btn-sm btn-outline-success" @click="openCreateGroupModal">
            + New Group
          </button>
          <button type="button" class="btn btn-sm btn-outline-primary" @click="openProfileModal">
            Edit Profile
          </button>
          <button type="button" class="btn btn-sm btn-outline-danger" @click="logout">
            Logout
          </button>
        </div>
      </div>
    </div>

    <ErrorMsg v-if="errormsg" :msg="errormsg" />

    <div class="row">
      <!-- Left column -->
      <div class="col-md-4">
        <!-- Search users -->
        <div class="mb-3 position-relative">
          <input
            v-model="searchQuery"
            @input="searchUsers"
            class="form-control"
            placeholder="Search for users..."
          />
          <div
            v-if="searchResults.length"
            class="list-group mt-1 shadow-sm position-absolute w-100"
            style="z-index: 1000"
          >
            <button
              v-for="user in searchResults"
              :key="user"
              @click="startConversation(user)"
              class="list-group-item list-group-item-action py-2"
            >
              Connect with: <strong>{{ user }}</strong>
            </button>
          </div>
        </div>

        <!-- Loading spinner -->
        <div v-if="loading" class="text-center my-3">
          <div class="spinner-border spinner-border-sm text-primary"></div>
        </div>

        <!-- Conversation list -->
        <div class="list-group shadow-sm">
          <button
            v-for="chat in conversations"
            :key="chat.id"
            @click="selectConversation(chat.id)"
            class="list-group-item list-group-item-action d-flex justify-content-between align-items-center"
          >
            <div>
              <strong>
                {{ chat.name }}
                <span v-if="chat.type === 'group'" class="badge bg-info ms-2">Group</span>
              </strong>
              <div class="small text-muted">
                {{ chat.lastMessage ? chat.lastMessage.preview : "No messages yet" }}
              </div>
            </div>
            
            <div class="d-flex align-items-center">
              <span v-if="chat.lastMessage" class="badge bg-secondary rounded-pill small me-2">
                {{
                  new Date(chat.lastMessage.timestamp).toLocaleTimeString([], {
                    hour: "2-digit",
                    minute: "2-digit"
                  })
                }}
              </span>
              <button 
                @click.stop="deleteConversation(chat.id)" 
                class="btn btn-sm btn-outline-danger py-0 px-1 ms-1"
                title="Delete conversation">
                &times;
              </button>
            </div>
          </button>
          <div
            v-if="conversations.length === 0 && !loading"
            class="list-group-item text-muted text-center py-4"
          >
            No active chats. Start searching to say hello!
          </div>
        </div>
      </div>

      <!-- Right column: ChatView -->
      <div class="col-md-8 border-start" style="min-height: 400px">
        <RouterView />
      </div>
    </div>

    <!-- ------------------------------------------------------------------
         Create Group Modal (unchanged)
    ------------------------------------------------------------------- -->
    <div v-if="showCreateGroup" class="modal show d-block" tabindex="-1" style="background-color: rgba(0,0,0,0.5);">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">Create New Group</h5>
            <button type="button" class="btn-close" @click="showCreateGroup = false"></button>
          </div>
          <div class="modal-body">
            <!-- Group name -->
            <div class="mb-3">
              <label class="form-label">Group Name</label>
              <input v-model="groupName" type="text" class="form-control" placeholder="e.g., Study Group">
            </div>

            <!-- Group photo -->
            <div class="mb-3">
              <label class="form-label">Group Photo (optional)</label>
              <div class="d-flex align-items-center">
                <button class="btn btn-outline-secondary me-2" @click="$refs.groupPhotoInput.click()">
                  📷 Upload
                </button>
                <input type="file" ref="groupPhotoInput" @change="onGroupPhotoSelected" accept="image/*" style="display: none;">
                <span v-if="groupPhotoPreview" class="position-relative d-inline-block">
                  <img :src="groupPhotoPreview" style="max-height: 50px; max-width: 50px;" class="rounded">
                  <button @click="clearGroupPhoto" class="btn btn-sm btn-danger position-absolute top-0 end-0 translate-middle" style="border-radius: 50%;">✕</button>
                </span>
              </div>
            </div>

            <!-- Add members -->
            <div class="mb-3">
              <label class="form-label">Add Members</label>
              <div class="position-relative">
                <input
                  v-model="userSearchQuery"
                  @input="searchUsersForGroup"
                  class="form-control"
                  placeholder="Search users..."
                />
                <div v-if="userSearchResults.length" class="list-group mt-1 shadow-sm position-absolute w-100" style="z-index: 1050;">
                  <button
                    v-for="user in userSearchResults"
                    :key="user"
                    @click="addMember(user)"
                    class="list-group-item list-group-item-action py-1"
                  >
                    + {{ user }}
                  </button>
                </div>
              </div>
            </div>

            <!-- Selected members -->
            <div v-if="selectedMembers.length" class="mb-3">
              <label class="form-label">Members ({{ selectedMembers.length }})</label>
              <div class="d-flex flex-wrap gap-1">
                <span v-for="member in selectedMembers" :key="member" class="badge bg-primary p-2">
                  {{ member }}
                  <button @click="removeMember(member)" class="btn btn-sm btn-link text-white p-0 ms-1" style="text-decoration: none;">✕</button>
                </span>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="showCreateGroup = false">Cancel</button>
            <button type="button" class="btn btn-success" @click="createGroup">Create Group</button>
          </div>
        </div>
      </div>
    </div>

    <!-- ------------------------------------------------------------------
         Profile Edit Modal
    ------------------------------------------------------------------- -->
    <div v-if="showProfileModal" class="modal show d-block" tabindex="-1" style="background-color: rgba(0,0,0,0.5);">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">Edit Profile</h5>
            <button type="button" class="btn-close" @click="showProfileModal = false"></button>
          </div>
          <div class="modal-body">
            <!-- Username -->
            <div class="mb-3">
              <label class="form-label">Username</label>
              <input v-model="profileName" type="text" class="form-control" placeholder="Your username">
            </div>

            <!-- Profile photo -->
            <div class="mb-3">
              <label class="form-label">Profile Photo (optional)</label>
              <div class="d-flex align-items-center">
                <button class="btn btn-outline-secondary me-2" @click="$refs.profilePhotoInput.click()">
                  📷 Upload
                </button>
                <input type="file" ref="profilePhotoInput" @change="onProfilePhotoSelected" accept="image/*" style="display: none;">
                <span v-if="profilePhotoPreview" class="position-relative d-inline-block">
                  <img :src="profilePhotoPreview" style="max-height: 50px; max-width: 50px;" class="rounded">
                  <button @click="clearProfilePhoto" class="btn btn-sm btn-danger position-absolute top-0 end-0 translate-middle" style="border-radius: 50%;">✕</button>
                </span>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="showProfileModal = false">Cancel</button>
            <button type="button" class="btn btn-primary" @click="updateProfile">Save Changes</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.list-group-item.active {
  background-color: #0d6efd;
  border-color: #0d6efd;
}
.modal {
  display: block;
}
</style>