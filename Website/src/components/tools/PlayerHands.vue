<script>
import CardPile from './CardPile.vue';

export default {
	// Hands: [][]Card
	// Names: []String
	// Position: 'primary'|'secondary'
	props: {
		hands: Array,
		names: Array,
		ltr: Boolean,
		position: {
			type: String,
			default: 'primary',
		},
		limit: {
			type: Number,
			default: Infinity,
		},
		cardShift: {
			type: Number,
			default: 20,
		},
		cardWidth: {
			type: Number,
			default: 60,
		},
		screenPad: {
			type: Number,
			default: 60,
		},
		smallScreenPad: {
			type: Number,
			default: 20,
		},
	},
	methods: {
		len(hand) {
			return Math.min(hand.length, this.limit);
		},
		computeCardShift(i) {
			const screen = i <= 1 ? window.innerWidth : window.innerHeight;
			const pad = screen < 480 ? this.smallScreenPad : this.screenPad;
			const condensed = this.cardWidth - (screen - pad * 2 - this.cardWidth) / (this.hands[i].length - 1);
			return Math.max(this.cardShift, condensed);
		},
	},
	components: { CardPile },
};
</script>

<template>
	<div
		v-for="(hand, i) in hands"
		:key="i"
		:class="`hand ${position} pos-${i}`"
		:style="{
			'--hand-half-width': `${(cardWidth + (cardWidth - computeCardShift(i)) * (len(hand) - 1)) / 2}px`,
			'--hand-half-width-raw': `${cardWidth * len(hand) / 2}px`,
		}">
		<span v-if="names">{{ name[i] }}</span>
		<CardPile :cards="hand" :limit="limit" :cardShift="computeCardShift(i)" :ltr="ltr"></CardPile>
	</div>
</template>
