<template>
  <div class="mention-filters">
    <div class="mention-filters__status">
      <a href="#" v-on:click="setStatus('new')" v-bind:class="mentionFilterStatus == 'new' ? 'active' : ''">New</a>
      <a href="#" v-on:click="setStatus('verified')" v-bind:class="mentionFilterStatus == 'verified' ? 'active' : ''">Verified</a>
      <a href="#" v-on:click="setStatus('approved')" v-bind:class="mentionFilterStatus == 'approved' ? 'active' : ''">Approved</a>
      <a href="#" v-on:click="setStatus('rejected')" v-bind:class="mentionFilterStatus == 'rejected' ? 'active' : ''">Rejected</a>
    </div>
  </div>
</template>
<script>
  import {mapState} from 'vuex';
  export default {
    methods: {
      setStatus(status) {
        this.$emit('change', status);
        this.$store.dispatch('setMentionFilterStatus', status);
      }
    },
    computed: {
      ...mapState(['mentionFilterStatus']),
    }
  };
</script>
