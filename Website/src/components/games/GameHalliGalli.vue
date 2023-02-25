<script setup>
import LobbyControls from '../tools/LobbyControls.vue';
import PlayerHands from '../tools/PlayerHands.vue';
</script>

<template>
	<LobbyControls
		@send="$emit('send', $event)"
		:data="data"
		:moves="moves"	
		:state="state" />

	<PlayerHands :hands="playedCards" position="secondary" :limit="1" :ltr="false" />
	<div class="hg-cards">
		<PlayerHands :hands="hands" :names="names" :activePlayers="activePlayers" :cardShift="55" :ltr="false" />
	</div>

	<div class="d-flex h-100 w-100 justify-content-center align-items-center flex-column">
		<span
			v-if="moves.includes('press')"
			class="pointer"
			style="font-size: 5rem; user-select: none;"
			@click="$emit('send', 'press')">🛎️</span>
	</div>
</template>

<style>
.hg-cards .hand.pos-0 > .playing-card {
	--card-stripe-color: #006bff;
}

.hg-cards .hand > .playing-card {
	color: white;
}
</style>

<script>
import Card from '../../Card';
import { reorder } from '../../util';

export default {
	props: ['state', 'moves', 'data'],
	emits: ['send'],
	computed: {
		players() {
			let order = this.state.player_order;
			const me = order.indexOf(this.state.me);
			order = order.slice(me).concat(order.slice(0, me));
			return reorder(order);
		},
		hands() {
			return this.players.map(id => {
				const myTurn = id === this.state.player_order[this.state.current_player] && this.moves.includes('draw');

				return Array(this.state.hands[id]).fill(new Card(
					`×${this.state.hands[id]}`,
					true,
					myTurn ? () => this.$emit('send', 'draw') : null
				));
			});
		},
		playedCards() {
			return this.players.map(id => {
				const played = this.state.played_cards[id];
				return played.map(c => new Card(this.cardToString(c)));
			});
		},
		names() {
			return this.players.map(id => this.state.lobby.clients[id].name);
		},
		activePlayers() {
			return [this.players.indexOf(this.state.player_order[this.state.current_player])];
		},
	},
	methods: {
		cardToString(card) {
			const fruit = ['🍓', '🍐', '🍑', '🍍'][Math.floor(card / 14)];
			const count = Math.floor((card % 14) / 3) + 1;

			return fruit.repeat(count);
		},
	},
};
</script>	