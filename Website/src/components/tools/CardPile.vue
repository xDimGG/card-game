<script>
export default {
	props: {
		cards: Array,
		limit: {
			type: Number,
			default: Infinity,
		},
		cardShift: {
			type: Number,
			default: 20,
		},
		ltr: {
			type: Boolean,
			default: true,
		},
	},
	computed: {
		renderedCards() {
			if (isFinite(this.limit)) {
				return this.cards.slice(-this.limit);
			}

			return this.cards;
		},
	}
};
</script>

<template>
	<div
		v-for="(card, j) in renderedCards"
		:key="j"
		:class="`playing-card ${card.hidden ? 'back' : ''} ${card.clickable ? 'pointer' : ''}`"
		:style="{
			'--translateX': `-${cardShift * j}px`,
			'z-index': ltr ? j : -j,
			...card.style,
		}"
		@click="card.onclick">
		{{ card.name }}
	</div>
</template>