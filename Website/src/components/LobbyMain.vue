<script setup>
import Crown from './icons/IconCrown.vue';
import BackArrow from './icons/IconBackArrow.vue';
import Clipboard from './icons/IconClipboard.vue';
</script>

<template>
	<main class="h-100 d-flex justify-content-center align-items-center text-dark">
		<BackArrow
			v-if="moves.includes('lobby.disconnect')"
			class="back-arrow"
			@click="disconnect" />
		<div
			v-if="moves.includes('lobby.join')"
			class="w-100 bg-white p-3 rounded white-shadow"
			:style="{ 'max-width': '20em' }">
			<h1 class="mb-3">CardGame™</h1>
			<form @submit.prevent="$emit('send', 'lobby.join', { lobby, name })">
				<input type="text" v-model="name" autofocus="autofocus" class="form-control mb-2" placeholder="Name" required>
				<input type="text" v-model="lobby" class="form-control mb-2" :placeholder="generatedLobbyID || 'Lobby ID'">
				<button
					type="submit"
					class="btn btn-md btn-primary mx-1"
					:style="{ color: '#fff !important' }">Join</button>
				<button
					type="submit"
					class="btn btn-md btn-primary mx-1"
					@mouseenter="generatedLobbyID = generateLobbyID()"
					@mouseleave="generatedLobbyID = ''"
					@click.prevent="$emit('send', 'lobby.join', {
						lobby: lobby || generatedLobbyID,
						name: name || 'Leader',
					})"
					:style="{ color: '#fff !important' }">Create</button>
			</form>
		</div>
		<div v-else-if="state && state.clients"
			class="mx-2 w-100 bg-white p-3 rounded white-shadow"
			:style="{ 'max-width': '45em' }">
			<h1 class="mb-3 d-flex align-items-center justify-content-center">Lobby {{ state.id }} <span class="h6 mb-0 ms-1 transfer-crown" @click="copyLink" ><Clipboard /></span></h1>
			<div class="list-unstyled border-top border-bottom d-flex justify-content-center">
				<table>
					<tr v-for="client in orderedClients" :key="client.id">
						<td>
							<Crown v-if="client.leader" />
							<Crown
								v-else-if="me.leader"
								class="transfer-crown"
								@click="$emit('send', 'lobby.transfer', { id: client.id })" />
						</td>
						<td class="ps-2 text-start">
							<span
								v-if="me.leader && !client.leader"
								@click="$emit('send', 'lobby.kick', { id: client.id })"
								class="kick pointer">{{ client.name }}</span>
							<span
								v-else
								:style="client.id === me.id ? { color: 'var(--yellow)' } : {}"
								:contenteditable="client.id === me.id"
								spellcheck="false"
								@keydown.enter="updateName">{{ client.name }}</span>
						</td>
					</tr>
				</table>
			</div>
			<div class="d-flex justify-content-around mt-3">
				<button
					v-for="game in games"
					:key="game.id"
					@click="moves.includes('lobby.select') ? $emit('send', 'lobby.select', { game: game.id }) : null"
					:class="`btn btn-sm ${game.id === state.game ? 'bg-success' : 'bg-dark'} border-2 mx-1`"
					:style="{ color: '#fff !important' }">{{ game.name }}</button>
			</div>
			<div>
				<button
					v-if="moves.includes('lobby.start')"
					@click="$emit('send', 'lobby.start')"
					class="btn btn-sm border-success bg-success btn-secondary mt-2 mx-1">Start</button>
				<button
					v-if="moves.includes('lobby.add_bot')"
					@click="$emit('send', 'lobby.add_bot')"
					class="btn btn-sm border-success btn-dark mt-2 mx-1">Add Bot</button>
			</div>
		</div>
		<a href="https://github.com/xDimGG/card-game" target="_blank" class="gh-icon">
			<GitHub />
		</a>
	</main>
</template>

<style>
.kick:hover {
	text-decoration: line-through;
}

.transfer-crown {
	fill: none;
	color: #ccc;
	cursor: pointer;
}

.transfer-crown:hover {
	color: #000;
}

.white-shadow {
	box-shadow: 0 0 2rem rgb(255 255 255 / 38%) !important;
}
</style>

<script>
import GitHub from './icons/IconGitHub.vue';
import { useToast } from 'vue-toastification';
const toast = useToast();

export default {
	data() {
		return {
			games: [
				{
					name: '¡Uno!',
					id: 'uno',
				},
				{
					name: 'Halli Galli',
					id: 'halli_galli',
				},
				{
					name: 'The Mind',
					id: 'the_mind',
				},
				{
					name: 'War!',
					id: 'war',
				},
			],
			lobby: location.hash.slice(1),
			name: '',
			generatedLobbyID: '',
		}
	},
	props: ['state', 'moves', 'data'],
	emits: ['send'],
	computed: {
		playedCards() {
			return this.state.round.play_pile.map(c => c.card);
		},
		orderedClients() {
			return Object.values(this.state.clients).sort((a, b) => {
				return a.joined_at - b.joined_at;
			});
		},
		me() {
			return this.state.clients[this.state.me];
		},
	},
	methods: {
		copyLink() {
			const text = `${location.origin}/#${encodeURIComponent(this.state.id)}`;
			this.$copyText(text).then(
				() => toast.info('Link copied to clipboard'),
				() => alert(`Couldn't copy link to clipboard. URL is ${text}`));
		},
		disconnect() {
			localStorage.clear();
			this.$emit('send', 'lobby.disconnect');
		},
		generateLobbyID() {
			return (Math.random() + 1).toString(36).slice(3, 8).toUpperCase();
		},
		updateName(evt) {
			evt.preventDefault();
			evt.target.innerText = evt.target.innerText.replace(/\n/g, ' '); // no newlines
			const name = evt.target.innerText.trim();
			if (!name) {
				evt.target.innerText = this.me.name;
				return;
			}

			evt.target.blur();
			this.$emit('send', 'lobby.rename', { name });
			this.$forceUpdate();
		},
	},
	components: { GitHub },
};
</script>
