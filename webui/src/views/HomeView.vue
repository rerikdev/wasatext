<template>
  <div class="d-flex" style="height: 100vh;">
    <!-- Sidebar -->
    <div class="bg-light p-3" style="width: 350px; min-width: 250px; max-width: 420px; border-right: 0; box-sizing: border-box;">
      <!-- Barra di ricerca utenti con dropdown Bootstrap -->
      <div class="dropdown w-100">
        <input
          v-model="search"
          class="form-control mb-3 dropdown-toggle"
          placeholder="Cerca utenti per username..."
          autocomplete="off"
          data-bs-toggle="dropdown"
          @input="searchUsers"
          @focus="dropdownOpen = true"
          @blur="closeDropdown"
        >
        <ul
          class="dropdown-menu w-100"
          :class="{ show: dropdownOpen && searchResults.length }"
          style="max-height: 300px; overflow-y: auto;"
        >
          <li
            v-for="user in searchResults"
            :key="user.username"
            class="dropdown-item d-flex align-items-center"
            @mousedown.prevent="startConversation(user)"
          >
            <img :src="user.profilePicture" alt="profile" width="32" height="32" class="rounded-circle me-2">
            <span>{{ user.username }}</span>
          </li>
        </ul>
      </div>

      <!-- Qui la lista delle conversazioni -->
      <ul class="list-group mt-3">
        <li
          v-for="t in orderedThreads"
          :key="'thread-' + (t.isGroup ? 'g' : 'dm') + '-' + t.id"
          class="list-group-item list-group-item-action d-flex align-items-center"
          :class="{ 'selected-conv': openConversation && openConversation.id === t.id }"
          style="cursor:pointer;"
          @click="openConv(t)"
        >
          <img :src="t.profilePicture" alt="avatar" width="40" class="rounded-circle me-2">
          <div class="flex-grow-1">
            <div class="fw-bold">{{ t.isGroup ? ('👥 ' + t.username) : t.username }}</div>
            <div class="text-muted small">
              {{ isImageLike(t.lastMessage) ? '📷 Foto' : (t.lastMessage && t.lastMessage.length > 12 ? t.lastMessage.slice(0, 12) + '…' : t.lastMessage) }}
            </div>
          </div>
          <div class="text-end small text-muted ms-2">
            {{ t.lastMessageTime }}
          </div>
        </li>
      </ul>

      <!-- In fondo alla sidebar, subito dopo </ul> -->
      <section class="create-group-section mt-4">
        <hr>
        <button class="btn btn-success w-100" @click="openCreateGroupModal = true">
          + Create Group
        </button>
      </section>

      <!-- Modale per creare gruppo -->
      <GroupModal
        :open-create-group-modal="openCreateGroupModal"
        :edit-group-mode="editGroupMode"
        :edit-group-id="editGroupId"
        @close-create-group-modal="closeCreateGroupModal"
        @refresh-groups="loadAll"
      />
    </div>

    <!-- Main content -->
    <div class="flex-grow-1">
      <!-- Top actions bar (no absolute, evita sovrapposizioni) -->
      <div class="p-3 d-flex justify-content-end gap-2">
        <button class="btn btn-outline-primary" @click="goToProfile">Profilo</button>
        <button class="btn btn-danger" @click="logout">Logout</button>
      </div>

      <div v-if="successMsg" class="alert alert-success text-center" style="max-width: 500px; margin: 0 auto 20px auto;">
        {{ successMsg }}
      </div>

      <ErrorMsg v-if="errormsg" :msg="errormsg" />

      <!-- Main content -->
      <div class="flex-grow-1 d-flex flex-column justify-content-center" style="height: 100vh;">
        <div
          v-if="openConversation"
          class="chat-box d-flex flex-column"
        >
          <div class="border-bottom p-3 rounded-top bg-white d-flex align-items-center justify-content-between">
            <h5 class="mb-0">{{ openConversation.username }}</h5>
            <div v-if="isGroup" class="ms-auto d-flex align-items-center">
              <GroupMembersButton
                :key="'gm-' + openConversation.id + '-' + (groupMembersForOpen?.length || 0)"
                :group-id="openConversation.id"
                :current-members="groupMembersForOpen"
                class="me-3"
                @refresh-groups="listGroups"
              />
              <AddMemberButton
                :group-id="openConversation.id"
                :current-members="groupMembersForOpen"
                class="me-1"
                @members-added="handleMembersAdded"
              />
              <LeaveGroupButton
                :group-id="openConversation.id"
                @left-group="handleGroupLeft"
              />
              <EditGroupButton
                :group-id="openConversation.id"
                :current-members="groupMembersForOpen"
                class="me-1"
                @group-updated="loadAll"
              />
            </div>
          </div>
          <!-- Sezione messaggi scrollabile -->
          <div class="flex-grow-1 overflow-auto p-3 messages-area" ref="messagesArea">
            <div
              v-for="msg in messages" 
              :id="'message-' + msg.id"
              :key="msg.id"
              class="d-flex mb-3 align-items-start message-wrapper"
              :class="isMyMessage(msg) ? '' : 'flex-row-reverse'"
            >
              <img :src="msg.sender.profilePicture" alt="profile" width="44" height="44" class="rounded-circle mx-2">
              <div class="message-container">
                <!-- Mostra il nome SOLO se il messaggio NON è mio -->
                <div
                  v-if="!isMyMessage(msg)"
                  class="fw-bold mb-1"
                  style="font-size: 1rem;"
                >
                  {{ msg.sender.username }}
                </div>

                <!-- Reply Reference (if this message is a reply) -->
                <div 
                  v-if="msg.replyToMessage" 
                  class="reply-reference mb-2"
                  :class="{ 'my-reply': isMyMessage(msg) }"
                  @click="scrollToMessage(msg.replyToMessageId)"
                >
                  <div class="reply-indicator"></div>
                  <div class="reply-content">
                    <span class="reply-author">{{ msg.replyToMessage.sender.displayName || msg.replyToMessage.sender.username }}</span>
                    <span class="reply-text">{{ truncateText(msg.replyToMessage.content, 50) }}</span>
                  </div>
                </div>

                <div
                  :class="isMyMessage(msg) ? 'msg-sent' : 'msg-received'"
                  style="font-size: 1.3rem; max-width: 520px; word-break: break-word;"
                >
                  <span v-if="msg.is_forwarded || msg.isForwarded" class="badge bg-warning text-dark mb-1" style="font-size: 0.9rem;">
                    Inoltrato
                  </span>
                  <template v-if="(msg.mediaType || msg.media_type) === 'image'">
                    <img
                      :src="msg.content"
                      class="chat-image rounded mb-1"
                      style="max-width:260px;max-height:260px;cursor:pointer;display:block;"
                      @click="fullscreenImage = msg.content"
                    >
                  </template>
                  <template v-else>
                    {{ msg.content }}
                  </template>
                  <div class="small text-muted mt-1 d-flex align-items-center">
                    <span>{{ msg.timestamp }}</span>
                    <span v-if="isMyMessage(msg)" :class="getStatusClass(msg.status)" style="margin-left: 8px;">
                      {{ getStatusIcon(msg.status) }}
                    </span>
                    <!-- NEW: Reply Button -->
                    <button 
                      class="btn btn-sm btn-link text-info ms-2" 
                      title="Rispondi" 
                      @click="handleReply(msg)"
                    >
                      ↩️
                    </button>
                    <button v-if="isMyMessage(msg)" class="btn btn-sm btn-link text-danger ms-2" title="Elimina" @click="deleteMessage(msg)">
                      🗑️
                    </button>
                    <button
                      class="btn btn-sm btn-link text-primary ms-2"
                      title="Inoltra"
                      @click="openForwardModal(msg)"
                    >
                      ⏩
                    </button>
                  </div>
                </div>
                <MessageReactions
                  :message="msg"
                  :conversation-id="openConversation.id"
                  @refresh="getConversation(openConversation.id)"
                />
              </div>
            </div>
          </div>

          <!-- NEW: Reply Preview Bar -->
          <div v-if="replyingTo" class="reply-preview-bar">
            <div class="reply-preview-content">
              <div class="reply-preview-label">
                <strong>↩️ Rispondi a {{ replyingTo.sender.displayName || replyingTo.sender.username }}</strong>
              </div>
              <div class="reply-preview-text">{{ truncateText(replyingTo.content, 80) }}</div>
            </div>
            <button class="btn-cancel-reply" @click="cancelReply" title="Annulla risposta">
              ✖
            </button>
          </div>

          <!-- Barra invio messaggi SEMPRE visibile in basso -->
          <form class="d-flex flex-column border-top p-3 bg-white rounded-bottom" @submit.prevent="sendMessage">
            <div v-if="imagePreview" class="mb-2 d-flex align-items-start gap-2">
              <div class="position-relative">
                <img :src="imagePreview" style="max-width:120px;max-height:120px;" class="rounded border">
                <button
                  type="button"
                  class="btn btn-sm btn-danger position-absolute top-0 start-100 translate-middle"
                  style="line-height:1;padding:2px 6px;"
                  @click="removeImage"
                >
                  ×
                </button>
              </div>
              <div class="small text-muted">
                Anteprima immagine
              </div>
            </div>
            <div class="d-flex w-100">
              <button
                type="button"
                class="btn btn-outline-secondary me-2"
                title="Immagine"
                @click="onPickImage"
              >
                📷
              </button>
              <input
                ref="imageInput"
                type="file"
                accept="image/*"
                class="d-none"
                @change="onImageSelected"
              >
              <input
                v-model="newMessage"
                ref="messageInput"
                class="form-control me-2"
                placeholder="Scrivi un messaggio..."
                :required="!imagePreview"
              >
              <button class="btn btn-primary" type="submit">Invia</button>
            </div>
          </form>
        </div>
        <div
          v-if="fullscreenImage"
          class="modal-backdrop"
          style="background:rgba(0,0,0,0.85);"
          @click="fullscreenImage = null"
        >
          <img
            :src="fullscreenImage"
            style="max-width:90%;max-height:90%;object-fit:contain;border:4px solid #fff;border-radius:12px;"
          >
        </div>
        <div v-else class="flex-grow-1 d-flex align-items-center justify-content-center text-muted">
          Nessuna conversazione trovata. Inizia a cercare utenti per iniziare una chat!
        </div>
      </div>
    </div>

    <!-- Modal per inoltrare messaggio -->
    <div v-if="forwardModalOpen" class="modal-backdrop">
      <div class="modal-dialog">
        <div class="modal-content p-3">
          <h5>Scegli la chat dove inoltrare</h5>

          <!-- NUOVO: cerca utenti (no gruppi) e inoltra -->
          <div class="mb-2">
            <input
              v-model="forwardSearch"
              class="form-control"
              placeholder="Cerca utente per inoltrare..."
              autocomplete="off"
              @input="forwardSearchUsers"
            >
          </div>
          <ul v-if="forwardSearch.length" class="list-group mb-3" style="max-height:220px; overflow:auto;">
            <li
              v-for="u in forwardResults"
              :key="'fuser-' + u.id"
              class="list-group-item list-group-item-action d-flex align-items-center"
              style="cursor:pointer;"
              @click="forwardToUser(u)"
            >
              <img :src="u.profilePicture || 'https://cdn-icons-png.flaticon.com/512/847/847969.png'" alt="profile" width="32" class="rounded-circle me-2">
              {{ u.username }}
            </li>
            <li v-if="forwardResults.length === 0" class="list-group-item text-muted">Nessun utente trovato</li>
          </ul>

          <ul class="list-group">
            <!-- Conversazioni 1:1 -->
            <li
              v-for="conv in conversations"
              :key="'chat-' + conv.id"
              class="list-group-item list-group-item-action"
              style="cursor:pointer;"
              @click="forwardMessage(conv.id)"
            >
              <img :src="conv.profilePicture" alt="profile" width="32" class="rounded-circle me-2">
              {{ conv.username }}
            </li>

            <!-- Gruppi -->
            <li
              v-for="group in groups"
              :key="'group-' + group.id"
              class="list-group-item list-group-item-action"
              style="cursor:pointer;"
              @click="forwardMessage(group.id)"
            >
              <img :src="group.photo || 'https://cdn-icons-png.flaticon.com/512/74/74472.png'" alt="group" width="32" class="rounded-circle me-2">
              👥 {{ group.name }}
            </li>
          </ul>

          <button class="btn btn-secondary mt-3" @click="closeForwardModal">Annulla</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import MessageReactions from '@/components/MessageReactions.vue';
import GroupModal from '@/components/GroupModal.vue';
import LeaveGroupButton from '@/components/LeaveGroupButton.vue';
import AddMemberButton from '@/components/AddMemberButton.vue';
import GroupMembersButton from '@/components/GroupMembersButton.vue';
import EditGroupButton from '@/components/EditGroupButton.vue';

export default {
  components: {
    MessageReactions,
    GroupModal,
    LeaveGroupButton,
    AddMemberButton,
    GroupMembersButton,
    EditGroupButton
  },
  data() {
    return {
      errormsg: null,
      loading: false,
      some_data: null,
      successMsg: null,
      search: "",
      searchResults: [],
      dropdownOpen: false,
      openConversationId: null,
      conversations: [],
      polling: null,
      openConversation: null,
      messages: [],
      messagesPolling: null,
      newMessage: "",
      forwardModalOpen: false,
      forwardMsg: null,
      groups: [],
      openCreateGroupModal: false,
      editGroupMode: false,
      editGroupId: null,
      imageFile: null,
      imagePreview: null,
      fullscreenImage: null,
      forwardSearch: "",
      forwardResults: [],
      replyingTo: null, // NEW: Message being replied to
    }
  },
  computed: {
    isGroup() {
      return this.openConversation && (
        this.openConversation.name ||
        (this.openConversation.username && this.openConversation.username.startsWith('👥'))
      );
    },
    orderedThreads() {
      const GROUP_ICON = 'https://cdn-icons-png.flaticon.com/512/74/74472.png';
      const convs = (this.conversations || []).map(c => ({ ...c, isGroup: false }));
      const grps = (this.groups || []).map(g => ({
        id: g.id,
        username: g.name,
        name: g.name,
        profilePicture: g.photo || GROUP_ICON,
        lastMessage: g.lastMessage || '',
        lastMessageTime: g.lastMessageTime || '',
        members: g.members || [],
        isGroup: true
      }));
      return [...convs, ...grps].sort((a, b) => {
        const ta = Date.parse(a.lastMessageTime) || 0;
        const tb = Date.parse(b.lastMessageTime) || 0;
        return tb - ta;
      });
    },
    groupMembersForOpen() {
      if (!this.openConversation) return [];
      const g = (this.groups || []).find(x => String(x.id) === String(this.openConversation.id));
      return Array.isArray(g?.members) ? g.members : (this.openConversation.members || []);
    }
  },
  mounted() {
    this.loadAll();
    this.polling = setInterval(this.loadAll, 6000);
    if (this.$route.query.msg) {
      this.successMsg = this.$route.query.msg;
      this.$router.replace({ path: this.$route.path, query: {} });
    }
  },
  beforeUnmount() {
    clearInterval(this.polling);
  },
  methods: {
    goToProfile() {
      this.$router.push('/profile');
    },
    logout() {
      localStorage.clear();
      this.$router.push('/');
    },
    async refresh() {
      this.loading = true;
      this.errormsg = null;
      try {
        let response = await this.$axios.get("/");
        this.some_data = response.data;
      } catch (e) {
        this.errormsg = e.toString();
      }
      this.loading = false;
    },
    async searchUsers() {
      if (this.search.length < 1) {
        this.searchResults = [];
        this.dropdownOpen = false;
        return;
      }
      const userId = localStorage.getItem("userId");
      const myUsername = localStorage.getItem("username");
      try {
        const res = await this.$axios.get(`/search/users?q=${encodeURIComponent(this.search)}`, {
          headers: { Authorization: userId }
        });
        const results = res.data;
        this.searchResults = results.filter(u => u.username !== myUsername);
        this.dropdownOpen = !!this.searchResults.length;
      } catch {
        this.searchResults = [];
        this.dropdownOpen = false;
      }
    },
    closeDropdown() {
      setTimeout(() => { this.dropdownOpen = false; }, 150);
    },
    async startConversation(user) {
      const userId = localStorage.getItem("userId");
      try {
        const res = await this.$axios.post("/conversations", { userId: user.id }, {
          headers: { "Content-Type": "application/json", Authorization: userId }
        });
        const data = res.data;
        if (data.conversationId) {
          this.search = "";
          this.searchResults = [];
          this.dropdownOpen = false;
          await this.loadAll();
          const conv = this.conversations.find(c => c.id === data.conversationId);
          if (conv) {
            this.openConv(conv);
          }
        }
      } catch {}
    },
    async getMyConversations() {
      const userId = localStorage.getItem("userId");
      try {
        const res = await this.$axios.get("/conversations", {
          headers: { Authorization: userId }
        });
        this.conversations = res.data;
        if (!this.openConversation && this.conversations.length > 0) {
          this.openConv(this.conversations[0]);
        }
      } catch {
        this.conversations = [];
      }
    },
    async openConv(conv) {
      this.openConversation = conv;
      this.messages = [];
      this.replyingTo = null; // NEW: Clear reply when switching conversations
      await this.getConversation(conv.id);
      await this.markMessagesRead();
      this.startMessagesPolling();
    },
    startMessagesPolling() {
      if (this.messagesPolling) clearInterval(this.messagesPolling);
      if (!this.openConversation) return;
      this.messagesPolling = setInterval(async () => {
        if (this.openConversation) {
          await this.getConversation(this.openConversation.id);
          await this.markMessagesRead();
        }
      }, 1000);
    },
    async getConversation(conversationId) {
      const userId = localStorage.getItem("userId");
      try {
        const res = await this.$axios.get(`/conversations/${conversationId}/messages`, {
          headers: { Authorization: userId }
        });
        const msgs = res.data;
        for (const msg of msgs) {
          try {
            const reactionRes = await this.$axios.get(
              `/conversations/${conversationId}/messages/${msg.id}/reactions`,
              { headers: { Authorization: userId } }
            );
            msg.reactions = reactionRes.data;
          } catch {
            msg.reactions = [];
          }
        }
        this.messages = msgs;
      } catch {
        this.messages = [];
      }
    },
    onPickImage() { this.$refs.imageInput.click(); },
    async onImageSelected(e) {
      const file = e.target.files && e.target.files[0];
      if (!file) return;
      if (file.size > 2 * 1024 * 1024) {
        alert("Immagine troppo grande (max 2MB)");
        e.target.value = "";
        return;
      }
      this.imageFile = file;
      const r = new FileReader();
      r.onload = ev => { this.imagePreview = ev.target.result; };
      r.readAsDataURL(file);
    },
    removeImage() {
      this.imageFile = null;
      this.imagePreview = null;
      if (this.$refs.imageInput) this.$refs.imageInput.value = "";
    },
    // NEW: Handle Reply
    handleReply(message) {
      this.replyingTo = message;
      this.$nextTick(() => {
        if (this.$refs.messageInput) {
          this.$refs.messageInput.focus();
        }
      });
    },
    // NEW: Cancel Reply
    cancelReply() {
      this.replyingTo = null;
    },
    // NEW: Scroll to Message
    scrollToMessage(messageId) {
      const element = document.getElementById(`message-${messageId}`);
      if (element) {
        element.scrollIntoView({ behavior: 'smooth', block: 'center' });
        element.classList.add('highlight-message');
        setTimeout(() => element.classList.remove('highlight-message'), 2000);
      }
    },
    // NEW: Truncate Text
    truncateText(text, maxLength) {
      if (!text) return '';
      if (text.length <= maxLength) return text;
      return text.substring(0, maxLength) + '...';
    },
    async sendMessage() {
      if (!this.openConversation) return;
      const hasText = this.newMessage.trim().length > 0;
      const hasImage = !!this.imagePreview;
      if (!hasText && !hasImage) return;
      const userId = localStorage.getItem("userId");
      try {
        const payload = {
          content: hasImage ? this.imagePreview : this.newMessage.trim(),
          mediaType: hasImage ? "image" : "text",
          isForwarded: false
        };

        // NEW: Add reply reference if replying
        if (this.replyingTo) {
          payload.replyToMessageId = this.replyingTo.id;
        }

        await this.$axios.post(`/conversations/${this.openConversation.id}/messages`, payload, {
          headers: { "Content-Type": "application/json", Authorization: userId }
        });
        this.newMessage = "";
        this.removeImage();
        this.replyingTo = null; // NEW: Clear reply after sending
        await this.getConversation(this.openConversation.id);

        // Scroll to bottom
        this.$nextTick(() => {
          const messagesArea = this.$refs.messagesArea;
          if (messagesArea) {
            messagesArea.scrollTop = messagesArea.scrollHeight;
          }
        });
      } catch {}
    },
    async markMessagesRead() {
      if (!this.openConversation) return;
      const userId = localStorage.getItem("userId");
      try {
        await this.$axios.patch(`/conversations/${this.openConversation.id}/messages/read`, {}, {
          headers: { Authorization: userId }
        });
      } catch {}
    },
    async deleteMessage(msg) {
      if (!confirm("Sei sicuro di voler eliminare questo messaggio?")) return;
      const userId = localStorage.getItem("userId");
      try {
        await this.$axios.delete(`/conversations/${msg.conversation_id}/messages/${msg.id}`, {
          headers: { Authorization: userId }
        });
        this.messages = this.messages.filter(m => m.id !== msg.id);
      } catch {
        alert("Errore durante l'eliminazione del messaggio.");
      }
    },
    isMyMessage(msg) {
      const myId = localStorage.getItem("userId");
      return String(msg.sender.id) === String(myId);
    },
    getStatusIcon(status) {
      if (status === "read") return "✔✔";
      return "✔";
    },
    getStatusClass(status) {
      if (status === "read") return "text-primary";
      return "text-secondary";
    },
    openForwardModal(msg) {
      this.forwardMsg = msg;
      this.forwardModalOpen = true;
    },
    async forwardSearchUsers() {
      const q = this.forwardSearch.trim();
      if (!q) { this.forwardResults = []; return; }
      const userId = localStorage.getItem("userId");
      const myUsername = localStorage.getItem("username");
      try {
        const res = await this.$axios.get(`/search/users?q=${encodeURIComponent(q)}`, {
          headers: { Authorization: userId }
        });
        const results = Array.isArray(res.data) ? res.data : [];
        this.forwardResults = results.filter(u => u.username !== myUsername);
      } catch {
        this.forwardResults = [];
      }
    },
    async forwardToUser(user) {
      if (!this.forwardMsg || !user?.id) return;
      const convId = await this.ensureConversationWithUser(user.id);
      if (convId) await this.forwardMessage(convId);
    },
    async ensureConversationWithUser(targetUserId) {
      const userId = localStorage.getItem("userId");
      try {
        const res = await this.$axios.post(
          "/conversations",
          { userId: targetUserId },
          { headers: { "Content-Type": "application/json", Authorization: userId } }
        );
        return res.data?.conversationId || null;
      } catch {
        return null;
      }
    },
    closeForwardModal() {
      this.forwardMsg = null;
      this.forwardModalOpen = false;
      this.forwardSearch = "";
      this.forwardResults = [];
    },
    async forwardMessage(targetConversationId) {
      if (!this.forwardMsg) return;
      const userId = localStorage.getItem("userId");
      try {
        await this.$axios.post(
          `/conversations/${this.forwardMsg.conversation_id}/messages/${this.forwardMsg.id}/forward`,
          { targetConversationId },
          { headers: { "Content-Type": "application/json", Authorization: userId } }
        );
        this.messages = this.messages.filter(m => m.id !== this.forwardMsg.id);
        this.successMsg = "Messaggio inoltrato!";
        setTimeout(() => { this.successMsg = null; }, 1000);
        this.closeForwardModal();
      } catch {
        alert("Errore durante l'inoltro del messaggio.");
      }
    },
    async loadAll() {
      await this.getMyConversations();
      await this.listGroups();
      if (this.openConversation) {
        await this.getConversation(this.openConversation.id);
        await this.markMessagesRead();
      }
    },
    async listGroups() {
      const userId = localStorage.getItem("userId");
      try {
        const res = await this.$axios.get("/groups", {
          headers: { Authorization: userId }
        });
        this.groups = res.data;
      } catch {
        this.groups = [];
      }
    },
    closeCreateGroupModal() {
      this.openCreateGroupModal = false;
      this.editGroupMode = false;
      this.editGroupId = null;
    },
    handleGroupLeft(groupId) {
      this.groups = this.groups.filter(g => g.id !== groupId);
      if (this.openConversation && this.openConversation.id === groupId) {
        if (this.messagesPolling) {
          clearInterval(this.messagesPolling);
          this.messagesPolling = null;
        }
        this.forwardModalOpen = false;
        this.forwardMsg = null;
        this.fullscreenImage = null;
        this.newMessage = '';
        this.removeImage?.();
        this.openConversation = null;
        this.messages = [];
        this.replyingTo = null; // NEW: Clear reply
      }
      this.listGroups();
    },
    handleMembersAdded() { this.loadAll(); },
    isImageLike(v) {
if (!v) return false;
const s = String(v);
return s.startsWith('data:image') || /.(png|jpe?g|gif|webp|bmp|svg)$/i.test(s);
}
}
}
</script>
<style>
.selected-conv {
    background: rgba(0, 123, 255, 0.15) !important;
}

.chat-box {
    width: 100%;
    height: calc(100vh - 120px); /* Fit within viewport */
    max-height: calc(100vh - 120px);
    background: #f8f9fa;
    border-radius: 12px;
    box-shadow: 0 2px 16px rgba(0,0,0,0.08);
    overflow: hidden;
    display: flex;
    flex-direction: column;
}

.messages-area {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    background: #f8f9fa;
    padding: 1rem;
}

.message-wrapper {
  transition: background-color 0.3s;
}

.message-container {
  max-width: 70%;
}

.msg-sent {
    background: #9accff;
    color: #222;
    display: inline-block;
    padding: 0.75rem;
    border-radius: 1.2rem;
}

.msg-received {
    background: #67b3ff;
    color: #222;
    display: inline-block;
    padding: 0.75rem;
    border-radius: 1.2rem;
}

/* Reply Reference Styles */
.reply-reference {
  background: rgba(0, 0, 0, 0.08);
  border-left: 3px solid #007bff;
  padding: 6px 10px;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.2s;
  max-width: 100%;
  font-size: 0.9rem;
}

.reply-reference:hover {
  background: rgba(0, 0, 0, 0.12);
}

.reply-reference.my-reply {
  background: rgba(255, 255, 255, 0.25);
  border-left-color: white;
}

.reply-reference.my-reply:hover {
  background: rgba(255, 255, 255, 0.35);
}

.reply-content {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.reply-author {
  font-weight: 600;
  font-size: 0.8rem;
  color: #007bff;
}

.my-reply .reply-author {
  color: white;
}

.reply-text {
  font-size: 0.8rem;
  opacity: 0.85;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* Reply Preview Bar Styles */
.reply-preview-bar {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  background: linear-gradient(135deg, #e3f2fd 0%, #f3f4f6 100%);
  border-top: 2px solid #007bff;
  border-left: 4px solid #007bff;
  animation: slideDown 0.3s ease-out;
}

@keyframes slideDown {
  from {
    transform: translateY(-20px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

.reply-preview-content {
  flex: 1;
  min-width: 0;
}

.reply-preview-label {
  font-size: 0.85rem;
  color: #007bff;
  margin-bottom: 2px;
}

.reply-preview-text {
  font-size: 0.9rem;
  color: #6c757d;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.btn-cancel-reply {
  background: none;
  border: none;
  font-size: 18px;
  cursor: pointer;
  color: #6c757d;
  padding: 4px 8px;
  transition: all 0.2s;
  border-radius: 50%;
  line-height: 1;
}

.btn-cancel-reply:hover {
  background: rgba(220, 53, 69, 0.1);
  color: #dc3545;
}

/* Highlight animation for scrolled messages */
@keyframes highlightPulse {
  0%, 100% { background-color: transparent; }
  50% { background-color: rgba(255, 235, 59, 0.4); }
}

.highlight-message {
  animation: highlightPulse 1.5s ease-in-out;
  border-radius: 12px;
}

.modal-backdrop {
  position: fixed;
  top: 0; 
  left: 0; 
  right: 0; 
  bottom: 0;
  background: rgba(0,0,0,0.5);
  z-index: 2000;
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-dialog {
  background: #fff;
  border-radius: 12px;
  max-width: 400px;
  width: 90%;
  max-height: 80vh;
  overflow-y: auto;
}

.create-group-section {
  padding: 1rem 0 0 0;
}

.chat-image:hover { 
  opacity: 0.9;
  transition: opacity .15s; 
}

/* Make sure parent containers don't overflow */
.d-flex.flex-column.justify-content-center {
  height: auto !important;
  max-height: 100vh;
  overflow: hidden;
}
</style>