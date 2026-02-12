<script>
export default {
  data() {
    return {
      messages: [],
      newMessage: "",
      errormsg: null,
      polling: null,
      selectedImage: null,
      selectedImagePreview: null,
      replyToMessage: null,

      // ✅ Stores the numeric conversation ID (from response header)
      numericConvId: null,

      // Conversation metadata
      convId: null,          // original param (username or old ID)
      convType: null,
      convName: "",
      convPhoto: null,
      isCreator: false,
      participants: [],

      // Group settings modal
      showGroupSettings: false,
      editGroupName: "",
      editGroupPhoto: null,
      editGroupPhotoPreview: null,

      // Add member modal
      showAddMember: false,
      addMemberSearch: "",
      addMemberResults: [],

      // Forward modal
      showForwardModal: false,
      forwardMessageId: null,
      forwardConversations: []
    };
  },
  computed: {
    myId() {
      return localStorage.getItem("userId");
    },
    myUsername() {
      return localStorage.getItem("username");
    },
    token() {
      return localStorage.getItem("token");
    }
  },
  watch: {
    '$route.params.id': {
      immediate: true,
      handler(newId) {
        if (newId) {
          this.convId = newId;
          this.numericConvId = null; // reset when switching conversations
          this.fetchConversationMetadata();
          this.fetchMessages();
          this.setupPolling();
        }
      }
    }
  },
  methods: {
    checkAuth() {
      if (!this.token) {
        this.$router.push("/login");
        return false;
      }
      return true;
    },

    // ------------------------------------------------------------
    // CONVERSATION METADATA
    // ------------------------------------------------------------
    async fetchConversationMetadata() {
      if (!this.checkAuth()) return;
      // This can be implemented later – not essential for fixing 403.
    },

    // ------------------------------------------------------------
    // FETCH MESSAGES – reads X-Conversation-Id header
    // ------------------------------------------------------------
    async fetchMessages() {
      if (!this.checkAuth()) return;
      try {
        const response = await this.$axios.get(
          `/conversations/${this.convId}?_=${Date.now()}`,
          { headers: { Authorization: "Bearer " + this.token } }
        );

        // ✅ Store the numeric conversation ID from header
        const convIdHeader = response.headers['x-conversation-id'];
        if (convIdHeader) {
          this.numericConvId = convIdHeader;
        }

        const data = response.data || [];
        this.messages = Array.isArray(data)
          ? data.filter((m) => m.content && !m.content.includes("No messages yet"))
          : [];

        // Mark messages as read
        if (this.numericConvId) {
          this.$axios.post(`/conversations/${this.numericConvId}/read`, null, {
            headers: { Authorization: "Bearer " + this.token }
          }).catch(() => {});
        }
      } catch (e) {
        if (e.response?.status === 401) {
          localStorage.clear();
          this.$router.push("/login");
        } else {
          this.errormsg = e.toString();
        }
      }
    },

    setupPolling() {
      if (this.polling) clearInterval(this.polling);
      // Start polling only after we have the numeric ID
      if (this.numericConvId) {
        this.polling = setInterval(this.fetchMessages, 3000);
      }
    },

    // ------------------------------------------------------------
    // SEND MESSAGE – uses numericConvId
    // ------------------------------------------------------------
    async sendMessage() {
      if (!this.newMessage.trim() && !this.selectedImage) return;
      if (!this.checkAuth()) return;
      if (!this.numericConvId) {
        this.errormsg = "Conversation not ready – please wait.";
        return;
      }

      const formData = new FormData();
      formData.append("content", this.newMessage);
      if (this.selectedImage) formData.append("image", this.selectedImage);
      if (this.replyToMessage) formData.append("reply_to", this.replyToMessage.id);

      try {
        await this.$axios.post(
          `/conversations/${this.numericConvId}/messages`,
          formData,
          {
            headers: {
              Authorization: "Bearer " + this.token,
              "Content-Type": "multipart/form-data"
            }
          }
        );
        this.newMessage = "";
        this.clearSelectedImage();
        this.replyToMessage = null;
        this.fetchMessages();
      } catch (e) {
        this.errormsg = "Failed to send: " + e.toString();
      }
    },

    // ------------------------------------------------------------
    // DELETE MESSAGE – uses numericConvId
    // ------------------------------------------------------------
    async deleteMessage(messageId) {
      if (!this.checkAuth()) return;
      if (!confirm("Unsend this message?")) return;
      if (!this.numericConvId) {
        this.errormsg = "Conversation not ready.";
        return;
      }
      try {
        await this.$axios.delete(
          `/conversations/${this.numericConvId}/messages/${messageId}`,
          { headers: { Authorization: "Bearer " + this.token } }
        );
        this.messages = this.messages.filter((m) => m.id !== messageId);
        setTimeout(() => this.fetchMessages(), 500);
      } catch (e) {
        this.errormsg = "Failed to delete message: " + e.toString();
        this.fetchMessages();
      }
    },

    // ------------------------------------------------------------
    // REACTIONS – uses numericConvId
    // ------------------------------------------------------------
    async addReaction(messageId, emoji) {
      if (!this.checkAuth()) return;
      if (!this.numericConvId) return;
      try {
        await this.$axios.put(
          `/conversations/${this.numericConvId}/messages/${messageId}/reactions/${encodeURIComponent(emoji)}`,
          null,
          { headers: { Authorization: "Bearer " + this.token } }
        );
        this.fetchMessages();
      } catch (e) {
        this.errormsg = "Reaction failed: " + e.toString();
      }
    },

    async removeReaction(messageId, emoji) {
      if (!this.checkAuth()) return;
      if (!this.numericConvId) return;
      try {
        await this.$axios.delete(
          `/conversations/${this.numericConvId}/messages/${messageId}/reactions/${encodeURIComponent(emoji)}`,
          { headers: { Authorization: "Bearer " + this.token } }
        );
        this.fetchMessages();
      } catch (e) {
        this.errormsg = "Could not remove reaction: " + e.toString();
      }
    },

    userHasReacted(msg) {
      return msg.reactions?.some((r) => String(r.user) === String(this.myId)) || false;
    },

    // ------------------------------------------------------------
    // IMAGE HANDLING (unchanged)
    // ------------------------------------------------------------
    onImageSelected(event) {
      const file = event.target.files[0];
      if (file) {
        this.selectedImage = file;
        this.selectedImagePreview = URL.createObjectURL(file);
      }
    },
    clearSelectedImage() {
      if (this.selectedImagePreview) URL.revokeObjectURL(this.selectedImagePreview);
      this.selectedImage = null;
      this.selectedImagePreview = null;
    },
    getImageSrc(image) {
      if (!image) return null;
      return image.startsWith("data:") ? image : `data:image/jpeg;base64,${image}`;
    },

    // ------------------------------------------------------------
    // REPLY HANDLING (unchanged)
    // ------------------------------------------------------------
    setReply(msg) {
      this.replyToMessage = {
        id: msg.id,
        content: msg.content,
        hasImage: msg.image ? true : false
      };
    },
    cancelReply() {
      this.replyToMessage = null;
    },

    // ------------------------------------------------------------
    // MESSAGE OWNERSHIP & DISPLAY (unchanged)
    // ------------------------------------------------------------
    isMyMessage(msg) {
      if (!msg.sender) return false;
      if (typeof msg.sender !== "object") {
        const s = String(msg.sender);
        return s === String(this.myId) || s === this.myUsername;
      }
      if (msg.sender.id && String(msg.sender.id) === String(this.myId)) return true;
      const senderName = msg.sender.name || msg.sender.username;
      if (senderName && this.myUsername) return senderName === this.myUsername;
      return false;
    },
    getSenderName(sender) {
      if (!sender) return "Unknown";
      if (typeof sender !== "object") {
        if (String(sender) === String(this.myId) || sender === this.myUsername) return "You";
        return `User ${sender}`;
      }
      if (sender.id && String(sender.id) === String(this.myId)) return "You";
      if (sender.name === this.myUsername || sender.username === this.myUsername) return "You";
      return sender.name || sender.username || sender.email || `User ${sender.id || "?"}`;
    },
    statusIcon(status) {
      if (status === "read") return "✓✓";
      if (status === "received") return "✓";
      if (status === "sent") return "🕐";
      return "";
    },

    // ------------------------------------------------------------
    // GROUP MANAGEMENT (unchanged, but uses numericConvId)
    // ------------------------------------------------------------
    async leaveGroup() {
      if (!confirm("Leave this group?")) return;
      if (!this.numericConvId) return;
      try {
        await this.$axios.delete(
          `/conversations/${this.numericConvId}/participants/${this.myId}`,
          { headers: { Authorization: "Bearer " + this.token } }
        );
        this.$router.push("/");
      } catch (e) {
        this.errormsg = "Failed to leave group: " + e.toString();
      }
    },

    openGroupSettings() {
      this.editGroupName = this.convName || "";
      this.editGroupPhoto = null;
      this.editGroupPhotoPreview = null;
      this.showGroupSettings = true;
    },

    onEditGroupPhotoSelected(event) {
      const file = event.target.files[0];
      if (file) {
        this.editGroupPhoto = file;
        this.editGroupPhotoPreview = URL.createObjectURL(file);
      }
    },
    clearEditGroupPhoto() {
      if (this.editGroupPhotoPreview) URL.revokeObjectURL(this.editGroupPhotoPreview);
      this.editGroupPhoto = null;
      this.editGroupPhotoPreview = null;
    },

    async updateGroup() {
      if (!this.editGroupName.trim()) {
        this.errormsg = "Group name cannot be empty";
        return;
      }
      if (!this.numericConvId) return;
      const formData = new FormData();
      formData.append("name", this.editGroupName);
      if (this.editGroupPhoto) formData.append("photo", this.editGroupPhoto);
      try {
        await this.$axios.put(`/conversations/${this.numericConvId}`, formData, {
          headers: {
            Authorization: "Bearer " + this.token,
            "Content-Type": "multipart/form-data"
          }
        });
        this.convName = this.editGroupName;
        if (this.editGroupPhotoPreview) {
          this.convPhoto = this.editGroupPhotoPreview;
        }
        this.showGroupSettings = false;
        this.fetchMessages();
      } catch (e) {
        this.errormsg = "Failed to update group: " + e.toString();
      }
    },

    openAddMember() {
      this.showAddMember = true;
      this.addMemberSearch = "";
      this.addMemberResults = [];
    },

    async searchAddMember() {
      if (this.addMemberSearch.length < 2) {
        this.addMemberResults = [];
        return;
      }
      try {
        const response = await this.$axios.get("/users", {
          params: { username: this.addMemberSearch },
          headers: { Authorization: "Bearer " + this.token }
        });
        const currentMemberNames = this.participants.map(p => p.name);
        this.addMemberResults = response.data.filter(
          u => !currentMemberNames.includes(u) && u !== this.myUsername
        );
      } catch (e) {
        this.errormsg = "Search failed: " + e.toString();
      }
    },

    async addMember(username) {
      if (!this.numericConvId) return;
      try {
        await this.$axios.post(
          `/conversations/${this.numericConvId}/participants`,
          { username },
          { headers: { Authorization: "Bearer " + this.token } }
        );
        const parts = await this.$axios.get(`/conversations/${this.numericConvId}/participants`, {
          headers: { Authorization: "Bearer " + this.token }
        });
        this.participants = parts.data;
        this.showAddMember = false;
      } catch (e) {
        this.errormsg = "Failed to add member: " + e.toString();
      }
    },

    // ------------------------------------------------------------
    // FORWARDING – uses numericConvId
    // ------------------------------------------------------------
    async openForwardModal(messageId) {
      this.forwardMessageId = messageId;
      const response = await this.$axios.get("/users/me/conversations", {
        headers: { Authorization: "Bearer " + this.token }
      });
      this.forwardConversations = response.data.filter(c => c.id !== this.numericConvId);
      this.showForwardModal = true;
    },

    async forwardMessage(targetConvId) {
      if (!this.forwardMessageId) return;
      if (!this.numericConvId) return;
      try {
        await this.$axios.post(
          `/conversations/${this.numericConvId}/forward/${targetConvId}`,
          { messageId: this.forwardMessageId },
          { headers: { Authorization: "Bearer " + this.token } }
        );
        this.showForwardModal = false;
        this.forwardMessageId = null;
      } catch (e) {
        this.errormsg = "Forward failed: " + e.toString();
      }
    }
  },
  mounted() {
    if (this.checkAuth()) {
      this.convId = this.$route.params.id;
      this.fetchConversationMetadata();
      this.fetchMessages().then(() => {
        // Start polling after first fetch, using numeric ID if available
        if (this.numericConvId) {
          this.polling = setInterval(this.fetchMessages, 3000);
        }
      });
    }
  },
  beforeUnmount() {
    clearInterval(this.polling);
    this.clearSelectedImage();
    if (this.editGroupPhotoPreview) URL.revokeObjectURL(this.editGroupPhotoPreview);
  }
};
</script>

<template>
  <!-- Your template remains exactly the same as before – no changes needed -->
  <div class="d-flex flex-column h-100 bg-white">
    <!-- Error display -->
    <div v-if="errormsg" class="alert alert-danger m-2 p-2 small">
      {{ errormsg }}
    </div>

    <!-- Chat Header (Group Info) -->
    <div v-if="convType === 'group'" class="d-flex align-items-center p-2 border-bottom bg-light">
      <div class="d-flex align-items-center flex-grow-1">
        <div v-if="convPhoto" class="me-2">
          <img :src="convPhoto" style="width: 40px; height: 40px; object-fit: cover; border-radius: 50%;">
        </div>
        <div v-else class="me-2 bg-secondary rounded-circle" style="width: 40px; height: 40px;"></div>
        <div>
          <strong>{{ convName || 'Group Chat' }}</strong>
          <span class="badge bg-info ms-2">Group</span>
          <div class="small text-muted">{{ participants.length }} members</div>
        </div>
      </div>
      <div>
        <button class="btn btn-sm btn-outline-secondary me-1" @click="openAddMember" title="Add Member">➕</button>
        <button class="btn btn-sm btn-outline-secondary me-1" @click="openGroupSettings" title="Group Settings">⚙️</button>
        <button class="btn btn-sm btn-outline-danger" @click="leaveGroup" title="Leave Group">🚪</button>
      </div>
    </div>

    <!-- Messages area -->
    <div class="flex-grow-1 overflow-auto p-3" style="max-height: 75vh">
      <div
        v-for="msg in messages"
        :key="msg.id"
        class="mb-3 d-flex"
        :class="isMyMessage(msg) ? 'justify-content-end' : 'justify-content-start'"
      >
        <div
          class="p-2 rounded shadow-sm"
          :class="isMyMessage(msg) ? 'bg-primary text-white' : 'bg-light border'"
          style="max-width: 75%; min-width: 150px"
        >
          <!-- Sender name (only for others) -->
          <div v-if="!isMyMessage(msg)" class="small fw-bold mb-1">
            {{ getSenderName(msg.sender) }}
          </div>

          <!-- Reply preview -->
          <div
            v-if="msg.reply"
            class="small p-1 mb-2 rounded"
            :class="isMyMessage(msg) ? 'bg-primary bg-opacity-75' : 'bg-white bg-opacity-50'"
            style="border-left: 3px solid #6c757d"
          >
            <span class="fw-bold">⤴ Reply to:</span>
            <div class="text-truncate">
              {{ msg.reply.content || (msg.reply.hasImage ? '🖼️ Image' : '') }}
            </div>
          </div>

          <!-- Forwarded indicator -->
          <div v-if="msg.content?.startsWith('[Forwarded]')" class="small fst-italic mb-1">
            📨 Forwarded
          </div>

          <!-- Message content -->
          <div class="text-break">{{ msg.content }}</div>

          <!-- Image attachment -->
          <div v-if="msg.image" class="mt-1">
            <img
              :src="getImageSrc(msg.image)"
              class="img-fluid rounded"
              style="max-height: 200px; max-width: 100%"
              alt="Message attachment"
            />
          </div>

          <!-- Reactions -->
          <div v-if="msg.reactions?.length" class="mt-2 d-flex flex-wrap gap-1">
            <span
              v-for="react in msg.reactions"
              :key="react.user"
              class="badge rounded-pill bg-white text-dark border d-inline-flex align-items-center px-2 py-1"
              :class="{ 'user-reaction': String(react.user) === String(myId) }"
              @click="String(react.user) === String(myId) ? removeReaction(msg.id, react.emoticon) : null"
              :style="{ cursor: String(react.user) === String(myId) ? 'pointer' : 'default' }"
            >
              {{ react.emoticon }}
              <span class="ms-1 small">{{ react.userName || react.user }}</span>
              <span
                v-if="String(react.user) === String(myId)"
                class="ms-1 text-danger fw-bold"
              >✕</span>
            </span>
          </div>

          <!-- Bottom row: actions + status -->
          <div class="d-flex justify-content-between align-items-center mt-2">
            <div>
              <!-- Reply button -->
              <button
                v-if="!isMyMessage(msg)"
                class="btn btn-sm btn-link p-0 me-2"
                :class="isMyMessage(msg) ? 'text-white' : 'text-secondary'"
                @click="setReply(msg)"
                title="Reply"
              >↩️</button>

              <!-- Add reaction -->
              <button
                v-if="!userHasReacted(msg)"
                class="btn btn-sm btn-outline-secondary p-0 px-2 py-1"
                :class="isMyMessage(msg) ? 'text-white border-white' : ''"
                @click="addReaction(msg.id, '👍')"
                title="Add reaction"
              >👍</button>

              <!-- Forward button -->
              <button
                class="btn btn-sm btn-outline-secondary p-0 px-2 py-1 ms-1"
                :class="isMyMessage(msg) ? 'text-white border-white' : ''"
                @click="openForwardModal(msg.id)"
                title="Forward"
              >📤</button>
            </div>

            <div>
              <!-- Delete button -->
              <button
                v-if="isMyMessage(msg)"
                class="btn btn-sm btn-link text-white p-0 ms-2 text-decoration-none"
                style="opacity: 0.7"
                @click="deleteMessage(msg.id)"
                title="Unsend message"
              >🗑️</button>

              <!-- Status icon -->
              <span
                v-if="isMyMessage(msg)"
                class="small text-white fw-bold ps-2"
                style="font-size: 0.8rem"
              >
                {{ statusIcon(msg.status) }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Reply preview bar -->
    <div v-if="replyToMessage" class="px-3 pt-2 bg-light border-top">
      <div class="d-flex align-items-center justify-content-between small">
        <span>
          <span class="fw-bold">Replying to:</span>
          {{ replyToMessage.content || (replyToMessage.hasImage ? '🖼️ Image' : '') }}
        </span>
        <button @click="cancelReply" class="btn btn-sm btn-link text-danger p-0">Cancel</button>
      </div>
    </div>

    <!-- Image preview -->
    <div v-if="selectedImagePreview" class="px-3 pt-2 bg-light border-top">
      <div class="position-relative d-inline-block">
        <img :src="selectedImagePreview" class="img-thumbnail" style="max-height: 80px" />
        <button
          @click="clearSelectedImage"
          class="btn btn-sm btn-danger position-absolute top-0 end-0 translate-middle"
          style="border-radius: 50%; padding: 0 6px"
        >✕</button>
      </div>
    </div>

    <!-- Message input -->
    <div class="p-3 border-top mt-auto bg-light">
      <div class="input-group">
        <button
          class="btn btn-outline-secondary"
          type="button"
          @click="$refs.fileInput.click()"
          title="Attach image"
        >📎</button>
        <input
          type="file"
          ref="fileInput"
          @change="onImageSelected"
          accept="image/*"
          style="display: none"
        />

        <input
          v-model="newMessage"
          @keyup.enter="sendMessage"
          class="form-control"
          placeholder="Type a message..."
        />

        <button
          @click="sendMessage"
          class="btn btn-primary"
          :disabled="!newMessage.trim() && !selectedImage"
        >Send</button>
      </div>
    </div>

    <!-- Group Settings Modal -->
    <div v-if="showGroupSettings" class="modal show d-block" tabindex="-1" style="background-color: rgba(0,0,0,0.5);">
      <!-- ... (unchanged, uses numericConvId internally) ... -->
    </div>

    <!-- Add Member Modal -->
    <div v-if="showAddMember" class="modal show d-block" tabindex="-1" style="background-color: rgba(0,0,0,0.5);">
      <!-- ... (unchanged) ... -->
    </div>

    <!-- Forward Modal -->
    <div v-if="showForwardModal" class="modal show d-block" tabindex="-1" style="background-color: rgba(0,0,0,0.5);">
      <!-- ... (unchanged) ... -->
    </div>
  </div>
</template>

<style scoped>
.user-reaction {
  background-color: #e3f2fd !important;
  border-color: #90caf9 !important;
}
.flex-grow-1 {
  scroll-behavior: smooth;
}
.modal {
  display: block;
}
</style>