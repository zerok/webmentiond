<template>
  <div class="rsvp-summary">
    <h3 class="rsvp-summary__title">RSVPs</h3>
    <div class="rsvp-summary__groups">
      <rsvp-summary-group label="Yes" :items="grouped['yes']" icon="calendar-check"/>
      <rsvp-summary-group label="Maybe" :items="grouped['maybe']" icon="calendar"/>
      <rsvp-summary-group label="No" :items="grouped['no']" icon="calendar-times"/>
    </div>
  </div>
</template>
<script>
import RSVPSummaryGroup from './RSVPSummaryGroup.vue';
export default {
  name: 'rsvp-summary',
  components: {
    'rsvp-summary-group': RSVPSummaryGroup
  },
  computed: {
    grouped() {
      const result = {'yes': [], 'no': [], 'maybe': [], 'interested': []};
      const mentions = (this.$store.state.mentions || []);
      for (const mention of mentions) {
        if (mention.type === 'rsvp') {
          if (typeof result[mention.rsvp] !== 'undefined') {
            result[mention.rsvp].push(mention);
          }
        }
      }
      return result;
    }
  }
}
</script>