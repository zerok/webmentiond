<template>
  <div>
  <h1>Mentions</h1>
  <div class="message" v-if="updateMentionStatusStatus == 'pending'">
    Update mention...
  </div>
  <mention-filters v-on:change="onFilterUpdate" />
  <ul v-if="mentions" class="mention-list">
    <li v-for="mention in mentions" class="mention">
      <div class="mention__info">
        <a class="mention__source" :href="mention.source">{{ mention.source }}</a>
        <i class="fas fa-long-arrow-alt-right mention__to"></i>
        <a class="mention__target" :href="mention.target">{{ mention.target }}</a>
        <span class="mention__created_at">({{ mention.created_at }})</span>
      </div>
      <div class="mention__actions">
      <button class="button button--small button--positive" v-on:click="approve(mention)"><i class="far fa-thumbs-up"></i> approve</button>
      <button class="button button--small button--negative" v-on:click="reject(mention)"><i class="far fa-thumbs-down"></i> reject</button>
      </div>
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
