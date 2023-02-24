import { createApp } from 'vue';
import App from './App.vue';

import Toast from 'vue-toastification';
import VueClipboard from 'vue-clipboard2';
import 'vue-toastification/dist/index.css';

import Lobby from './components/LobbyMain.vue';
import GameTheMind from './components/games/GameTheMind.vue';
import GameWar from './components/games/GameWar.vue';
import GameUno from './components/games/GameUno.vue';
import GameHalliGalli from './components/games/GameHalliGalli.vue';

import './assets/style.css';

const app = createApp(App);

/* eslint-disable vue/multi-word-component-names */
app.component('lobby', Lobby);
app.component('the_mind', GameTheMind);
app.component('war', GameWar);
app.component('uno', GameUno);
app.component('halli_galli', GameHalliGalli);

app.use(Toast);
VueClipboard.config.autoSetContainer = true
app.use(VueClipboard);

app.mount('#app');
