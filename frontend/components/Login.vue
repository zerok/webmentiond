<template>
  <form v-on:submit="onSubmit" class="login">
    <h1 class="title">Log in</h1>
    <div class="main">
    <Loading v-if="requestTokenStatus == 'pending'"/>
    <p class="message message--success" v-if="requestTokenStatus == 'succeeded'">
      If an account exists for this e-mail, you just received a login
      link. You can also authenticate with a token on <router-link to="/authenticate">this
      page</router-link>.
    </p>
    <p class="message message--error" v-if="requestTokenStatus == 'failed'">
      Failed to send an authentication link.
    </p>
    <p class="info">If you already have an authentication code, enter it on <router-link to="/authenticate">this
      page</router-link>.</p>
    <label>E-mail: <input v-model="email" type="email"></label>
    <div class="form__actions">
      <button class="button button--primary" type="submit">Request login</button>
    </div>
    </div>
  </form>
</template>
<script>
  import {mapGetters, mapState} from 'vuex';
  import Loading from './Loading.vue';

  export default {
    components: {Loading},
    data() {
      return {
        email: ""
      };
    },
    computed: {
      ...mapState({requestTokenStatus: 'requestTokenStatus'}),
    },
    methods: {
      onSubmit(evt) {
        evt.preventDefault();
        this.$store.dispatch('requestToken', this.$data.email);
      }
    }
  };
  </script>
