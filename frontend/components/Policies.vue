<template>
  <div>
    <h1 class="title"><img src="../css/webmentiond-logo.svg" alt="" />  Policies</h1>
    <div class="main">
      <Loading v-if="createPolicyLoading || policiesLoading || deletePolicyLoading" />
      <Error v-if="policiesError" err="Failed to load policies" />
      <Error v-if="deletePolicyError" err="Failed to delete policy" />
      <Error v-if="createPolicyError" err="Failed to create policy" />
      <ul v-if="policies && policies.length" class="policy-listing">
        <li v-for="policy in policies" :key="policy.pattern + policy.policy" class="policy">
          <span class="policy__pattern">{{ policy.url_pattern }}</span>
          <span class="policy__policy"><i class="fa fa-exclamation"></i>{{ policy.policy }}</span>
          <span class="policy__weight"><i class="fa fa-weight-hanging"></i>{{ policy.weight }}</span>
          <button class="button button--negative" @click="deletePolicy(policy.id)"><i class="fa fa-trash"></i> Delete</button>
        </li>
      </ul>
      <p class="empty" v-else>No policies defined yet 🙂</p>

      <form class="form" @submit="createPolicy">
        <h2>Create a new policy</h2>
        <div class="form-field">
          <label class="form-field__label" for="create-urlpattern">URL pattern:</label>
          <div class="form-field__control">
            <input id="create-urlpattern" type="text" v-model="newPattern" />
          </div>
        </div>
        <div class="form-field">
          <label class="form-field__label" for="create-policy">Policy:</label>
          <div class="form-field__control">
            <select id="create-policy" v-model="newPolicy">
              <option value="approve">Approve</option>
            </select>
          </div>
        </div>
        <div class="form-field">
          <label class="form-field__label" for="create-weight">Weight:</label>
          <div class="form-field__control">
            <input id="create-weight" type="number" v-model="newWeight" />
          </div>
        </div>
        <div class="form-actions">
          <button class="button button--primary" type="submit"><i class="fa fa-file-plus"></i> Create</button>
        </div>
      </form>
    </div>
  </div>
</template>
<script>
import Loading from './Loading.vue';
import Error from './Error.vue';
import {mapState} from 'vuex';
export default {
  components: {Loading, Error},
  data() {
    return {
      creating: false,
      deleting: false,
      newPattern: '',
      newWeight: 1,
      newPolicy: 'approve'
    };
  },
  methods: {
    deletePolicy(id) {
      this.$data.deleting = true;
      this.$store.dispatch('deletePolicy', id);
    },
    createPolicy(evt) {
      evt.preventDefault();
      this.$data.creating = true;
      this.$store.dispatch('createPolicy', {
        urlPattern: this.$data.newPattern,
        policy: this.$data.newPolicy,
        weight: this.$data.newWeight
      });
    }
  },
  computed: {...mapState([
    'policies', 'policiesLoading', 'policiesError',
    'deletePolicyLoading', 'deletePolicyError',
    'createPolicyLoading', 'createPolicyError',
  ])},
  created() {
    this.$store.dispatch('getPolicies');
  },
  updated() {
    if ((!this.deletePolicyLoading && this.$data.deleting) || (!this.createPolicyLoading && this.$data.creating)) {
      this.$data.deleting = false;
      this.$data.creating = false;
      this.$store.dispatch('getPolicies');
      this.$data.newPolicy = 'approve';
      this.$data.newPattern = '';
      this.$data.newWeight = 1;
    }
  }
}
</script>