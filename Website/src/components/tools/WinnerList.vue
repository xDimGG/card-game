<script setup>
import Crown from '../icons/IconCrown.vue';
import ordinal from 'ordinal';
</script>

<script>
export default {
	props: {
		winners: {
			// Object mapping position to array of names
			type: [Array, Object],
			required: true,
		},
	},
	computed: {
		normalized() {
			return Object.fromEntries(
				Object.entries(
					Object.values(this.winners).map(w => Array.isArray(w) ? w : [w])));
		},
	},
}
</script>

<template>
	<div class="winner-list">
		<Crown class="winner-crown" />
		<div
			v-for="(names, pos) in normalized"
			:id="`place-${pos}`">
				<span v-for="name in names">{{ ordinal(parseInt(pos) + 1) }} - {{ name }}
				</span>
		</div>
	</div>
</template>

<style>
.winner-list {
	text-align: center;
}

.winner-crown {
	font-size: 10rem;
	fill: none;
	color: var(--yellow);
}

.winner-list span {
	display: block;
	font-size: 1.75rem;
}

#place-0 {
	margin-top: -50px;
}

#place-0 > span {
	color: var(--yellow);
	font-size: 4rem;
}

#place-1 > span {
	color: #C0C0C0;
	font-size: 2.5rem;
}

#place-2 > span {
	color: #CD7F32;
	font-size: 2rem;
}
</style>
