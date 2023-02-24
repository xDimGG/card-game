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
	<PlayerHands :hands="hands" :cardShift="55" :ltr="false" />

	<div class="d-flex h-100 w-100 justify-content-center align-items-center flex-column">
		<span
			v-if="moves.includes('press')"
			class="pointer"
			style="font-size: 5rem; user-select: none;"
			@click="$emit('send', 'press')">🛎️</span>
	</div>
</template>

<style>
@media only screen and (max-width: 600px) {
	.pos-2, .pos-3 {
		--hand--offset: -40px !important;
	}
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
				const active = id === this.state.player_order[this.state.current_player];
				const myTurn = this.moves.includes('draw') && active;

				const style = { color: 'white' };
				if (active) style['box-shadow'] = '0 0 10px 0px #c1c1c1';
				if (this.state.me === id) style['--card-stripe-color'] = '#006bff';

				return Array(this.state.hands[id]).fill(new Card(
					`×${this.state.hands[id]}`,
					true,
					myTurn ? () => this.$emit('send', 'draw') : null,
					style,
				));
			});
		},
		playedCards() {
			return this.players.map(id => {
				const played = this.state.played_cards[id];
				return played.map(c => new Card(this.cardToString(c)));
			});
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