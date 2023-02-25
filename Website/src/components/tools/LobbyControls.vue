<script setup>
import BackArrow from '../icons/IconBackArrow.vue';
</script>

<template>
	<div v-if="state.lobby.frozen" class="freeze d-flex justify-content-center align-items-center flex-column">
		<span>Game is currently frozen. Please be patient.</span>
		<button
			v-if="moves.includes('lobby.return')"
			class="btn btn-sm btn-primary"
			@click="$emit('send', 'lobby.return')">Return to lobby</button>
	</div>
	<BackArrow
		v-else-if="moves.includes('lobby.return')"
		class="back-arrow"
		@click="$emit('send', 'lobby.return')" />
</template>

<style>
.freeze {
	position: absolute;
	height: 100vh;
	width: 100vw;
	color: #fff;
	background: rgba(0, 0, 0, 0.5);
	z-index: 9999;
}
</style>

<script>
export default {
	props: ['state', 'moves', 'data'],
	emits: ['send'],
};
</script>