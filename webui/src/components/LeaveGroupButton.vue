<template>
  <button
    class="btn btn-outline-danger btn-sm ms-3"
    title="Lascia gruppo"
    @click="leaveGroup"
  >
    Lascia Gruppo
  </button>
</template>

<script>
export default {
  name: 'LeaveGroupButton',
  props: {
    groupId: {
      type: [String, Number],
      required: true
    }
  },
  methods: {
    async leaveGroup() {
      if (!confirm("Sei sicuro di voler lasciare questo gruppo?")) return;
      const userId = localStorage.getItem("userId");
      try {
        await this.$axios.delete(`/groups/${this.groupId}/members`, {
          headers: { Authorization: userId }
        });
        this.$emit('left-group', this.groupId); // Passa l'id del gruppo!
      } catch (e) {
        alert("Errore durante l'uscita dal gruppo.");
      }
    }
  }
}
</script>