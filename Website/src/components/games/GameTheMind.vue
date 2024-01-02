<script setup>
import Card from '../../Card';
import PlayerHands from '../tools/PlayerHands.vue';
import LobbyControls from '../tools/LobbyControls.vue';
import Heart from '../icons/IconHeart.vue';
</script>

<template>
	<LobbyControls
		@send="$emit('send', $event)"
		:data="data"
		:moves="moves"
		:state="state" />

	<div class="game-center">
		<div class="hand">
			<div
				v-for="(card, i) in lastTwoCards"
				:key="card"
				:class="`playing-card ${i === 0 && lastTwoCards.length > 1 ? 'inactive' : null}`"
				:style="{ '--translateX': lastTwoCards.length === 1 ? null : `${-20 * i - 10}px` }">
				{{ card }}
			</div>
		</div>
		<span>
			<Heart
				v-for="n in Math.max(5, state.lives)"
				:key="n"
				:fill="n <= state.lives ? 'white' : null" />
		</span>
		<span>
			Round: {{ state.round_num }}/{{ 16 - (2 * Object.keys(state.lobby.clients).length) }}
		</span>
		<span>
			Shurikens: {{ state.shurikens }}
		</span>
		<span v-if="state.won">You won the game. Awesome!</span>
		<span v-else-if="state.lost">You lost the game. :(</span>
		<span v-else-if="state.round.won">You won the round. Great job.</span>
		<span v-else-if="state.round.lost">You lost the round... but you can try again!</span>
		<button
			v-if="moves.includes('next_round')"
			class="btn btn-sm btn-light mt-2"
			@click="$emit('send', 'next_round')">Next Round {{ reward }}</button>
		<button
			v-if="moves.includes('retry_round')"
			class="btn btn-sm btn-light mt-2"
			@click="$emit('send', 'retry_round')">Retry round</button>
		<button
			v-if="moves.includes('restart_game')"
			class="btn btn-sm btn-light mt-2"
			@click="$emit('send', 'restart_game')">Restart game</button>
		<button
			v-if="moves.includes('use_shuriken')"
			class="btn btn-sm btn-light mt-2"
			@click="$emit('send', 'use_shuriken')">Use shuriken</button>
	</div>

	<PlayerHands :hands="hands" :names="names" />
</template>

<script>
export default {
	props: ['state', 'moves', 'data'],
	emits: ['send'],
	computed: {
		reward() {
			if (!this.data) return null;

			switch (this.data.reward) {
			case 'shuriken':
				return '(+1 Shuriken)';
			case 'life':
				return '(+1 Life)';
			}

			return null;
		},
		hands() {
			return [this.state.hand, ...this.otherHands]
				.map(c => c.sort(
					(a, b) => (a || Infinity) - (b || Infinity)))
				.map(h => h.map(c => {
					if (c && this.moves.includes(c.toString())) {
						return new Card(c, false, () => this.$emit('send', c.toString()));
					}

					return new Card(c, c === null, null, { '--card-stripe-color': '#006bff' });
				}));
		},
		otherHands() {
			return Object.keys(this.state.other_hands_exposed || {}).length === 0
				? Object
					.entries(this.state.other_hands)
					.map(([k, v]) => {
						if (this.state.round.lowest_cards?.[k]) {
							return [this.state.round.lowest_cards[k], ...Array(v - 1).fill(null)];
						}

						return Array(v).fill(null);
					})
				: Object.values(this.state.other_hands_exposed);
		},
		lastTwoCards() {
			return this.state.round.play_pile.slice(-2);
		},
		names() {
			const ids = [this.state.lobby.me, ...Object.keys(this.state.lobby.clients).filter(k => k !== this.state.lobby.me)];
			return ids.map(id => this.state.lobby.clients[id].name);
		},
	},
};
</script>