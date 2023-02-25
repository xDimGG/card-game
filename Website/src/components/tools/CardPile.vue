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

<style>
:root {
	--card-height: 100px;
	--card-width: 60px;
	--card-hover-scale: 1.3;
	--card-font-size: 1.3rem;
	--card-stripe-color: #dc3545;
}

.playing-card {
	user-select: none;
	width: var(--card-width);
	height: var(--card-height);
	font-size: var(--card-font-size);
	background-color: white;
	color: black;
	border-radius: 6px;
	border: 3px solid gray;
	/* box-shadow: 0 0 6px black; */
	transform: translateX(var(--translateX));
	display: flex;
	align-items: center;
	justify-content: center;
}

.playing-card {
	transition: box-shadow 0.3s, border-color 0.3s;
}

.hand.active > .back.playing-card {
	box-shadow: inset 0 0 10px 3px white, 0 0 10px white;
	border-color: white;
}

.playing-card.inactive {
	filter: brightness(0.5);
}

.playing-card.back {
	background-color: #000000;
	background-size: 10px 10px;
	background-image: repeating-linear-gradient(45deg, #000000 0, var(--card-stripe-color) 1px, #000000 0, #000000 50%);
	border: 3px solid black;
}

.playing-card:not(.no-hover):hover {
	z-index: 999 !important;
	transform: translate(var(--translateX), calc(var(--card-height) * (1 - var(--card-hover-scale)) / 2)) scale(var(--card-hover-scale)) !important;
	filter: none;
}
</style>