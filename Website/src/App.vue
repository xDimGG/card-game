<template>
	<component
		@send="send"
		:is="statePacket.game"
		:data="statePacket.data"
		:moves="statePacket.moves"
		:state="statePacket.state"></component>
</template>

<script>
import jmp from 'json-merge-patch';
import { useToast } from 'vue-toastification';
const toast = useToast();

export default {
	data() {
		return {
			statePacket: {
				moves: [],
				game: 'lobby',
				state: null,
				data: null,
			},
			ws: null,
		};
	},
	methods: {
		send(moves, data) {
			return this.ws.send(JSON.stringify({
				type: 'message',
				moves: moves && Array.isArray(moves) ? moves : [moves],
				data,
			}));
		},
		connect() {
			this.ws = new WebSocket(process.env.NODE_ENV === 'development'
				? 'ws://localhost:8080/ws'
				: `ws://${location.host}/ws`);
			this.ws.onmessage = msg => {
				const packet = JSON.parse(msg.data);

				if (packet.type === 'error') {
					if ([
						'lobby no longer exists',
						'game has ended',
						'invalid client ID',
						'you are already connected',
						'you have been kicked'].includes(packet.data.message)) {
						localStorage.clear();
					}
					toast.error(packet.data.message);
					return;
				}

				delete packet.type;
				jmp.apply(this.statePacket, packet);
				console.log(JSON.parse(JSON.stringify(this.statePacket.state || '{}')));

				const { game, moves, state } = this.statePacket;

				if (game === 'lobby') {
					const id = localStorage.getItem('last_lobby_id');
					const me = localStorage.getItem('last_lobby_me');
					if (moves.includes('lobby.reconnect') && id && me) {
						this.send('lobby.reconnect', { id, me });
					}
					if (state) {
						localStorage.setItem('last_lobby_id', state.id);
						localStorage.setItem('last_lobby_me', state.me);
					}
				}
			};
			this.ws.onclose = () => this.connect();
			this.ws.onerror = err => {
				console.error;
			};
		},
	},
	mounted() {
		this.connect();
	},
	destroy() {
		if (this.ws) this.ws.close();
	},
};
</script>
