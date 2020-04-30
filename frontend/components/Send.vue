<template>
  <div>
  <h1 class="title">Send mention</h1>
  <div class="main">
    <form @submit.prevent>
      <input class="input--url input" type="text" v-model="source">
      <button class="button button--action button--primary" type="submit" @click="submit">Send</button>
    </form>
    <Loading v-if="sendStatus == 'pending'"/>
    <ul v-if="sendStatusReport" class="sendreport">
        <li :class="{'sendreport__item': true, 'sendreport__item--success': target.endpoint && !target.error, 'sendreport__item--failure': target.error}" v-for="target in sendStatusReport.targets">
          <a class="sendreport__item__url" :href="target.url">{{ target.url }}</a>
          <span class="sendreport__item__endpoint" v-if="target.endpoint">Endpoint: {{ target.endpoint }}</span>
          <span class="sendreport__item__error" v-if="target.error">Error: {{ target.error }}</span>
        </li>
    </ul>
  </div>
</div>
</template>
<script>
  import {mapState} from 'vuex';
  import Loading from './Loading.vue';
export default {
  components: {
    Loading
  },
  data() {
    return {
      source: ''
    };
  },
  methods: {
    submit(evt) {
      evt.preventDefault();
      this.$store.dispatch('sendMention', this.$data.source);
    },
  },
  computed: {
    ...mapState(['sendStatus', 'sendStatusReport'])
  }
};
</script>
