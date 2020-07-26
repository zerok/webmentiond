<template>
<div class="webmention-list">
  <p v-if="!mentions">Loading...</p>
  <p v-if="mentions && !mentions.length">This page hasn't been mentioned anywhere yet.</p>
  <ul class="webmention-list__list" v-if="mentions && mentions.length">
    <li v-for="mention in mentions" :key="mention.id">
      <div class="webmention webmention--comment" v-if="mention.type == 'comment'">
        <i class="fa fa-comment"></i>
        <a class="webmention__author" :href="mention.source">{{ mention.author_name }}</a>
        <span class="webmention__date">@ {{ mention.created_at }}</span>
        <blockquote class="webmention__content">{{ mention.content }}</blockquote>
      </div>
      <div class="webmention" v-else-if="mention.type == 'rsvp'">
        <i class="fa fa-calendar-check" v-if="mention.rsvp == 'yes'"></i>
        <i class="fa fa-calendar-times" v-else-if="mention.rsvp == 'no'"></i>
        <i class="fa fa-calendar-star" v-else-if="mention.rsvp == 'interested'"></i>
        <i class="fa fa-calendar" v-else></i>
        <a class="webmention__author" :href="mention.source">{{ mention.author_name }}</a>
        <span class="webmention__rsvp" v-if="mention.rsvp == 'yes'">will attend</span>
        <span class="webmention__rsvp" v-if="mention.rsvp == 'no'">will not attend</span>
        <span class="webmention__rsvp" v-if="mention.rsvp == 'maybe'">will maybe attend</span>
        <span class="webmention__rsvp" v-if="mention.rsvp == 'interested'">is interested in the event</span>
        <span class="webmention__date">@ {{ mention.created_at }}</span>
      </div>
      <div class="webmention" v-else>
        <i class="fa fa-link"></i>
        <a class="webmention__source" :href="mention.source">{{ mention.title }}</a>
        <span v-if="mention.author_name">by {{ mention.author_name }}</span>
        <span class="webmention__date">@ {{ mention.created_at }}</span>
      </div>
    </li>
  </ul>
</div>
</template>
<script>
export default {
  computed: {
    mentions() {
      return this.$store.state.mentions;
    }
  }
};
</script>
