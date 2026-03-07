<template>

  <div class="d-inline">
    <button
      class="btn btn-outline-secondary btn-sm me-2"
      title="Modifica gruppo"
      @click="openModal"
    >
      Modifica
    </button>

    <div v-if="open" class="egb-backdrop" @click.self="closeModal">
      <div class="egb-modal">
        <h6 class="mb-3">Modifica gruppo</h6>

        <div class="mb-2">
          <label class="form-label mb-1">Nome</label>
          <input v-model="nameLocal" class="form-control form-control-sm" maxlength="32">
        </div>

        <div class="mb-2">
          <label class="form-label mb-1">URL foto</label>
          <input v-model="photoLocal" class="form-control form-control-sm" placeholder="https://...">
          <div v-if="photoLocal" class="d-flex align-items-center mt-2">
            <img :src="photoLocal" alt="" width="36" height="36" class="rounded me-2" @error="onImgError">
            <small class="text-muted">Anteprima</small>
          </div>
        </div>

        <div v-if="error" class="alert alert-danger py-1 px-2 my-2">{{ error }}</div>

        <div class="d-flex justify-content-end gap-2 mt-3">
          <button class="btn btn-secondary btn-sm" :disabled="loading" @click="closeModal">Annulla</button>
          <button class="btn btn-primary btn-sm" :disabled="loading || !changed" @click="save">
            <span v-if="loading" class="spinner-border spinner-border-sm me-1" />
            Salva
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'EditGroupButton',
  props: {
    groupId: { type: [String, Number], required: true },
    groupName: { type: String, default: '' },
    groupPhoto: { type: String, default: '' }
  },
  emits: ['refresh-groups'],
  data() {
    return {
      open: false,         
      nameLocal: '',       
      photoLocal: '',      
      loading: false,      
      error: ''            
    }
  },
  computed: {
    changed() {
      return this.nameLocal !== (this.groupName || '') || this.photoLocal !== (this.groupPhoto || '');
    }
  },
  methods: {
    openModal() {
      this.nameLocal = this.groupName || '';
      this.photoLocal = this.groupPhoto || '';
      this.error = '';
      this.open = true;
    },
    closeModal() {
      if (this.loading) return;
      this.open = false;
    },
    onImgError(e) {
      e.target.onerror = null;
      e.target.src = 'https://cdn-icons-png.flaticon.com/512/74/74472.png';
    },

    async setGroupName() {
      const userId = localStorage.getItem('userId');
      await this.$axios.patch(
        `/groups/${this.groupId}/name`,
        { name: this.nameLocal },
        { headers: { Authorization: userId } }
      );
    },
    async setGroupPhoto() {
      const userId = localStorage.getItem('userId');
      await this.$axios.patch(
        `/groups/${this.groupId}/photo`,
        { photo: this.photoLocal },
        { headers: { Authorization: userId } }
      );
    },

    async save() {
      this.loading = true; this.error = '';
      try {
        if (this.nameLocal !== (this.groupName || '')) {
          await this.setGroupName();
        }
        if (this.photoLocal !== (this.groupPhoto || '')) {
          await this.setGroupPhoto();
        }
        this.$emit('refresh-groups');
        this.closeModal();
      } catch (e) {
        this.error = e?.response?.data?.message || 'Errore salvataggio';
      } finally {
        this.loading = false;
      }
    }
  }
}
</script>

<style scoped>
.egb-backdrop {
  position: fixed; inset: 0; background: rgba(0,0,0,.35);
  display: flex; align-items: center; justify-content: center;
  z-index: 2100;
}
.egb-modal {
  background: #fff; width: 320px; border-radius: 12px;
  padding: 14px 16px 16px; box-shadow: 0 4px 18px rgba(0,0,0,.15);
}
</style>