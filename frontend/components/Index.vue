<template>
  <div>
  <h1 class="title">Mentions</h1>
  <div class="main">
  <mention-filters v-on:change="onFilterUpdate" />
  <Loading v-if="updateMentionStatusStatus == 'pending'" />
  <ul v-if="mentions && mentions.length" class="mention-list">
    <li v-for="mention in mentions" class="mention">
      <div class="mention__info">
        <span class="mention__title" v-if="mention.title">{{ mention.title }}</span>
        <a class="mention__source" :href="mention.source">{{ mention.source }}</a>
        <i class="fas fa-long-arrow-alt-right mention__to"></i>
        <a class="mention__target" :href="mention.target">{{ mention.target }}</a>
        <span class="mention__created_at">({{ mention.created_at }})</span>
        <div class="mention__content" v-if="mention.content">
          {{ mention.content }}
        </div>
      </div>
      <div class="mention__actions">
      <button class="button button--small button--positive" v-on:click="approve(mention)"><i class="far fa-thumbs-up"></i> approve</button>
      <button class="button button--small button--negative" v-on:click="reject(mention)"><i class="far fa-thumbs-down"></i> reject</button>
      </div>
    </li>
  </ul>
  <p class="empty" v-else>No mentions found.</p>
  </div>
  </div>
</template>
<script>
  import {mapState} from 'vuex';
  import MentionFilters from './MentionFilters.vue';
  import Loading from './Loading.vue';
export default {
  components: {
    MentionFilters,
    Loading,
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
