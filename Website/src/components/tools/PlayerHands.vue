<script>
import CardPile from './CardPile.vue';

export default {
	// hands: [][]Card
	// names: []String
	// activePlayers: []Number
	// position: 'primary'|'secondary'
	props: {
		hands: Array,
		names: Array,
		activePlayers: {
			type: Array,
			default: [],
		},
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
		computeNameHeight() {
			return this.names ? 24 : 0;
		},
		computeHandHalfWidth(i) {
			const name = this.computeNameHeight();
			const l = this.len(this.hands[i]);
			if (l === 0) return name / 2;
			return (this.cardWidth + (this.cardWidth - this.computeCardShift(i)) * (l - 1)) / 2 + name;
		},
	},
	components: { CardPile },
};
</script>

<template>
	<div
		v-for="(hand, i) in hands"
		:key="i"
		:class="`hand ${position} pos-${i} ${activePlayers?.includes(i) ? 'active' : ''}`"
		:style="{
			'--hand-half-width': `${computeHandHalfWidth(i)}px`,
			'--hand-half-width-raw': `${cardWidth * len(hand) / 2 + computeNameHeight()}px` }">
		<span v-if="names && hand.length > 0" class="player-name">{{ names[i] }}</span>
		<CardPile :cards="hand" :limit="limit" :cardShift="computeCardShift(i)" :ltr="ltr"></CardPile>
	</div>
</template>

<style>
:root {
	--hand--offset: 10px;
}

.hand {
	display: flex;
	z-index: 1;
	--hand-offset: var(--hand--offset);
}

.hand.secondary {
	z-index: 2;
	--hand-offset: calc(var(--card-height) + var(--hand--offset) * 2);
}

.hand.pos-0 {
	position: fixed;
	bottom: var(--hand-offset);
	left: calc(50vw - var(--hand-half-width));
	z-index: 3;
}

.hand.pos-1 {
	position: fixed;
	top: var(--hand-offset);
	transform: rotate(180deg);
	right: calc(50vw - var(--hand-half-width));
}

.hand.pos-2 {
	position: fixed;
	top: calc(50vh - var(--card-height) - var(--hand-half-width));
	left: var(--hand-offset);
	transform-origin: bottom left;
	transform: rotate(90deg);
}

.hand.pos-3 {
	position: fixed;
	top: calc(50vh - (var(--hand-half-width) + (var(--hand-half-width-raw) - var(--hand-half-width)) * 2));
	right: calc(var(--hand-offset) + var(--card-height));
	transform-origin: top right;
	transform: rotate(-90deg);
}

@media only screen and (max-width: 600px) {
	.hand.pos-2, .hand.pos-3 {
		--hand--offset: -40px !important;
	}

	.hand.pos-2 > .player-name {
		padding-top: 40px;
	}

	.hand.pos-3 > .player-name {
		padding-bottom: 40px;
	}
}

.hand > .player-name {
	writing-mode: vertical-lr;
	transform: rotate(180deg);
	height: var(--card-height);
	text-overflow: ellipsis;
	overflow: hidden;
	white-space: nowrap;
}

.hand.pos-0 > .player-name {
	color: var(--yellow);
}

.hand.pos-3 > .player-name {
	transform: none;
}

/* .pos-1 > .playing-card:not(.no-hover):hover {
	transform: translate(var(--translateX), calc(var(--card-height) * (var(--card-hover-scale) - 1) / 2)) scale(var(--card-hover-scale)) !important;
} */
</style>
