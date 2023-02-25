<script setup>
import Card from '../../Card';
import LobbyControls from '../tools/LobbyControls.vue';
import PlayerHands from '../tools/PlayerHands.vue';
</script>

<template>
	<LobbyControls
		@send="$emit('send', $event)"
		:data="data"
		:moves="moves"	
		:state="state" />

	<div class="uno-cards">
		<PlayerHands :hands="hands()" :cardShift="10" :cardWidth="70"></PlayerHands>
	</div>

	<div class="d-flex h-100 w-100 justify-content-center align-items-center flex-column">
		<div :class="`circle color-${state.chosen_color}`" v-if="state.chosen_color !== -1"></div>

		<div class="hand uno-cards">
			<div
				v-for="(card, i) in lastTwoCards"
				:key="card"
				:class="['playing-card', i === 0 && lastTwoCards.length > 1 ? 'inactive' : '']"
				:style="{
					'--translateX': lastTwoCards.length === 1 ? null : `${-40 * i + 3}px`,
					...card.style,
				}">
				{{ card.name }}
			</div>
		</div>

		<button
			v-if="moves.includes('draw')"
			class="btn btn-sm btn-light mt-2"
			@click="$emit('send', 'draw')">Draw {{ Math.max(1, state.draw_num) }}</button>

		<button
			v-if="moves.includes('uno')"
			class="btn btn-sm btn-light mt-2"
			@click="$emit('send', 'uno')">Uno!</button>

		<!-- Hidden circle to center everything -->
		<div class="circle" v-if="state.chosen_color !== -1" style="visibility: hidden;"></div>
	</div>
	<div
		v-if="moves.includes(`${color_prompt}_0`)"
		class="hand pos-0 secondary d-flex justify-content-center w-100">
		<div
			v-for="n in 4"
			:key="n"
			:class="`circle color-${n - 1} choice pointer`"
			@click="$emit('send', `${color_prompt}_${n - 1}`)"></div>
	</div>
</template>

<style>
.circle {
	width: 70px;
	height: 70px;
	border-radius: 100%;
	border: white 1px solid;
	margin: 20px;
}

.circle.choice {
	transition: box-shadow 0.2s;
	box-shadow: 0 0 0 white;
}

.circle.choice:hover {
	box-shadow: 0 0 16px white;
}

.color-0 { background-color: #ff5555; }
.color-1 { background-color: #ffaa00; }
.color-2 { background-color: #55aa55; }
.color-3 { background-color: #5555ff; }

@media screen and (max-width: 600px) {
	.circle {
		width: 60px;
		height: 60px;	
		margin: 15px;
	}
}

.uno-cards .playing-card {
	--card-width: 70px;
	--card-height: 102px;
	border-radius: 10px;
}

.uno-cards .playing-card:not(.back) {
	background-image: url('@/assets/uno_cards.svg');
  background-repeat: no-repeat;
	background-size: 1400%;
}
</style>

<script>
import { reorder } from '../../util';

export default {
	props: ['state', 'moves', 'data'],
	emits: ['send'],
	data() {
		return {
			color_prompt: -1,
		};
	},
	computed: {
		players() {
			let order = this.state.player_order;
			const me = order.indexOf(this.state.me);
			order = order.slice(me).concat(order.slice(0, me));
			return reorder(order);
		},
		lastTwoCards() {
			return this.state.play_pile.slice(-2).map(r => this.convertCard(r));
		},
	},
	methods: {
		hands() {
			return this.players.map(id => {
				if (this.state.me === id)
					return this.state.hand.map((raw, i) => {
						const card = this.convertCard(raw);

						if (this.moves.includes(raw.toString())) {
							card.onclick = () => this.$emit('send', raw.toString());
						} else if (this.moves.includes(`${raw}_0`)) {
							// Color choosing cards
							card.onclick = () => {
								if (this.color_prompt === raw) {
									this.color_prompt = -1;
								} else {
									this.color_prompt = raw;
								}
							};
						} else {
							card.style.filter = 'brightness(0.5)';
						}

						return card;
					});

				return Array(this.state.other_hands[id]).fill(new Card('', true, null, {
					'--card-stripe-color': id === this.state.player_order[this.state.current_player] ? 'white' : null,
				}));
			});
		},
		convertCard(raw) {
			const card = new Card();
			if (raw >= 15 * 4) {
				raw -= 14 * 4
			}

			let x = Math.floor(raw / 4);
			let y = raw % 4;
			if (x === 14) {
				y += 4;
				x = 13;
			}

			card.style['background-position-x'] = `${-64 * x}px`;
			card.style['background-position-y'] = `${-96 * y}px`;

			return card;
		},
	}
}
</script>