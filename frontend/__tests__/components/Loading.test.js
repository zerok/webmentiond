import Loading from '../../components/Loading.vue';
import { render } from '@testing-library/vue';

describe('Loading', () => {
  it('should result in a div with the class "loading"', async () => {
    const {container} = render(Loading);
    expect(container.querySelector('.loading')).toBeDefined();
  });
});
