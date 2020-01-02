<template>
  <form v-on:submit="onSubmit" class="auth">
    <h1>Authenticate account</h1>
    <label>Token: <input v-model="token"></label>
    <div class="form__actions">
      <button type="submit" class="button button--primary">Authenticate</button>
    </div>
  </form>
</template>
<script>
  import {mapState} from 'vuex';
  export default {
    methods: {
      onSubmit(evt) {
        this.$store.dispatch('authenticate', this.$data.token);
      }
    },
    watch: {
      authStatus(newState, oldState) {
        if (newState === 'succeeded') {
          this.$router.push('/');
        }
      }
    },
    computed: {
      ...mapState(['authStatus'])
    },
    data: function() {
      return {
        token: this.$route.params.token
      };
    }
  };
</script>
