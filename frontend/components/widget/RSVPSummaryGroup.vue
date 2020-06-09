<template>
  <div class="rsvp-summary__group">
    <h4 class="rsvp-summary__group__title"><i :class="'fa fa-' + icon"></i> {{ label }} ({{ items.length }})</h4>
    <ul class="rsvp-summary__group__items" v-if="items.length">
      <li v-for="item in items"><a :href="item.source">{{ item.author_name || item.title }}</a></li>
    </ul>
  </div>
</template>
<script>
export default {
  props: ['items', 'label', 'icon']
}
</script>