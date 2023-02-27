<script setup>
import LobbyControls from '../tools/LobbyControls.vue';
import PlayerHands from '../tools/PlayerHands.vue';
import WinnerList from '../tools/WinnerList.vue';
</script>

<template>
	<LobbyControls
		@send="$emit('send', $event)"
		:data="data"
		:moves="moves"
		:state="state" />

	<PlayerHands :hands="placedCards" position="secondary"/>
	<PlayerHands :hands="hands" :names="names" position="primary" />
	<div class="game-center">
		<WinnerList v-if="state.game_over" :winners="winners" />
		<span
			v-else
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
		names() {
			const ids = [this.state.me, ...Object.keys(this.state.lobby.clients).filter(k => k !== this.state.me)];
			return ids.map(id => this.state.lobby.clients[id].name);
		},
		winners() {
			const best = Object.values(this.state.wins).sort((a, b) => b - a);
			const winners = {};
			for (const [id, wins] of Object.entries(this.state.wins)) {
				const pos = best.indexOf(wins);
				if (!winners[pos]) winners[pos] = [];
				winners[pos].push(this.state.lobby.clients[id].name);
			}

			return winners;
		},
	},
};
</script>