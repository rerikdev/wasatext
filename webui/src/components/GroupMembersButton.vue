<template>
  <div class="d-inline">
    <button
      class="btn btn-outline-secondary btn-sm me-2"
      title="Membri"
      aria-label="Membri"
      v-bind="$attrs"
      @click="openModal"
    >
      Membri
    </button>

    <div v-if="open" class="modal-backdrop" @click.self="closeModal">
      <div class="modal-dialog">
        <div class="modal-content p-3">
          <h5 class="mb-3">Membri del gruppo</h5>
          <ul class="list-group" style="max-height:300px;overflow:auto;">
            <li
              v-for="m in displayedMembers"
              :key="m.username"
              class="list-group-item d-flex align-items-center"
            >
              <img :src="m.profilePicture || fallbackAvatar" alt="" width="32" height="32" class="rounded-circle me-2">
              <span>{{ m.username }}</span>
            </li>
            <li v-if="!displayedMembers.length" class="list-group-item text-muted">Nessun membro</li>
          </ul>
          <div class="text-end mt-3">
            <button class="btn btn-secondary btn-sm" @click="closeModal">Chiudi</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'GroupMembersButton',
  inheritAttrs: false,
  emits: ['refresh-groups'],
  props: {
    groupId: { type: [String, Number], required: true },
    currentMembers: { type: Array, default: () => [] }
  },
  data() {
    return {
      open: false,
      fallbackAvatar: 'https://cdn-icons-png.flaticon.com/512/847/847969.png'
    }
  },
  computed: {
    displayedMembers() {
      return (this.currentMembers || []).map(m => {
        if (typeof m === 'string') return { username: m, profilePicture: '' };
        if (m && typeof m === 'object') {
          return { username: m.username || String(m.id || ''), profilePicture: m.profilePicture || m.photo || '' };
        }
        return { username: String(m), profilePicture: '' };
      }).filter(x => x.username);
    }
  },
  methods: {
    openModal() {
      this.$emit('refresh-groups');
      this.open = true;
    },
    closeModal() {
      this.open = false;
      this.$emit('refresh-groups');
    }
  }
}
</script>

<style scoped>
/* rimosso .action-btn per usare le classi Bootstrap come gli altri bottoni */
.modal-backdrop {
  position: fixed; inset: 0; background: rgba(0,0,0,0.3);
  z-index: 2000; display: flex; align-items: center; justify-content: center;
}
.modal-dialog { background: #fff; border-radius: 12px; max-width: 420px; width: 100%; }
</style>