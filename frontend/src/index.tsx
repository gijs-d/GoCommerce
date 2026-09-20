import { render } from 'solid-js/web';
import { App } from './App';
import './styles/glass.scss';

const root = document.getElementById('root');
if (root) {
  render(() => <App />, root);
}