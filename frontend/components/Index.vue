<template>
  <div>
  <h1>Index</h1>
  <div class="message" v-if="updateMentionStatusStatus == 'pending'">
    Update mention...
  </div>
  <mention-filters v-on:change="onFilterUpdate" />
  <ul v-if="mentions">
    <li v-for="mention in mentions">
      {{ mention.status }}: {{ mention.source }} to {{ mention.target }}
      <button v-on:click="approve(mention)">approve</button>
      <button v-on:click="reject(mention)">reject</button>
    </li>
  </ul>
  </div>
</template>
<script>
  import {mapState} from 'vuex';
  import MentionFilters from './MentionFilters.vue';
export default {
  components: {
    MentionFilters,
  },
  data() {
    return {
      updateTriggered: false
    };
  },
  created() {
    this.$store.dispatch('getMentions');
  },
  updated() {
    if (this.updateTriggered) {
      if (this.updateMentionStatusStatus === 'succeeded') {
        this.$store.dispatch('getMentions', {
          status: this.mentionFilterStatus
        });
      }
      if (this.updateMentionStatusStatus !== 'pending') {
        this.updateTriggered = false;
      }
    }
  },
  methods: {
    onFilterUpdate(status) {
      this.$store.dispatch('getMentions', {
        'status': status
      });
    },
    approve(mention) {
      this.$data.updateTriggered = true;
      this.$store.dispatch('approveMention', mention);
    },
    reject(mention) {
      this.$data.updateTriggered = true;
      this.$store.dispatch('rejectMention', mention);
    },
  },
  computed: {
    ...mapState(['mentions', 'updateMentionStatusStatus', 'mentionFilterStatus'])
  }
};
</script>
