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

	<PlayerHands :hands="placedCards" position="secondary"/>
	<PlayerHands :hands="hands" position="primary" />
	<div class="d-flex h-100 w-100 justify-content-center align-items-center flex-column">
		<span
			v-for="(score, id) in state.wins"
			:key="id">{{ state.lobby.clients[id].name }}: {{ score }}</span>
		<button
			v-if="moves.includes('next_round')"
			class="btn btn-sm btn-light mt-2"
			@click="$emit('send', 'next_round')">Next Round</button>

		<button
			v-if="moves.includes('war')"
			class="btn btn-sm btn-light mt-2"
			@click="$emit('send', 'war')">War!</button>
	</div>
	Player
</template>

<script>
import Card from '../../Card';

export default {
	props: ['state', 'moves', 'data'],
	emits: ['send'],
	computed: {
		placedCards() {
			const hands = [
				this.state.placed ? [new Card(this.state.placed)] : [],
				...(this.state.placed_revealed
					? Object.values(this.state.placed_revealed).map(c => [new Card(c)])
					: Object.values(this.state.other_placed).map(c => c ? [new Card('', true)] : []))
			];
			const winningHand = hands.find(h => h[0] && h[0].name == this.state.round_highest);
			if (winningHand) {
				winningHand[0].style = { 'box-shadow': '0 0 20px 8px rgba(235, 255, 70)' };
			}

			return hands;
		},
		hands() {
			return [
				this.state.hand
					.map(c => new Card(c, false, this.moves.includes(c.toString())
						? () => this.$emit('send', c.toString())
						: null)),
				...Object.values(this.state.other_hands)
					.map(c => Array(c).fill(new Card(null, true))),
			];
		},
	},
};
</script>