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

<script>
export default {
	props: ['state', 'moves', 'data'],
	emits: ['send'],
};
</script>