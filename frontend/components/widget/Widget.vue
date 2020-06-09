<template>
  <div class="webmention-widget">
    <h2>{{ title }}</h2>
    <webmention-list />
    <rsvp-summary v-if="showRSVPSummary" />
  </div>
</template>
<script>
import WebmentionList from './WebmentionList.vue';
import RSVPSummary from './RSVPSummary.vue';

export default {
  components: {
    WebmentionList,
    'rsvp-summary': RSVPSummary
  },
  mounted: function() {
    if (this.$props.mentions === null) {
      this.$store.dispatch('fetchMentions', {
        endpoint: this.$props.endpoint,
        target: this.$props.target
      });
    } else {
      this.$store.commit('setMentions', this.$props.mentions);
    }
  },
  props: {
    mentions: {
      type: Array,
      default: () => {
        return null
      }
    },
    title: {
      type: String,
      default: 'Mentions'
    },
    endpoint: {
      type: String
    },
    showRSVPSummary: {
      type: Boolean,
      default: false
    },
    target: {
      type: String
    }
  }
}
</script>