<template>
  <form v-on:submit="onSubmit">
    {{ requestTokenStatus }}
    <label>E-mail: <input v-model="email" type="email"></label>
    <button type="submit">Request login</button>
  </form>
</template>
<script>
  import {mapGetters, mapState} from 'vuex';

  export default {
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
        console.log(this.$data.email);
        this.$store.dispatch('requestToken', this.$data.email);
      }
    }
  };
  </script>
