<template>
  <i :class="'fa fa-fw fa-' + icon"></i>
</template>
<script>
export default {
  props: ['mention'],
  computed: {
    icon() {
      const mention = this.$props.mention;
      switch (mention.type) {
        case 'comment':
          return 'comment';
        case 'like':
          return 'heart';
        case 'rsvp':
          switch(mention.rsvp) {
            case 'yes':
              return 'calendar-check';
            case 'no':
              return 'calendar-times';
            case 'interested':
              return 'calendar-star';
            default:
              return 'calendar';
          }
        default:
          return 'link';
      }
    }
  }
}
</script>