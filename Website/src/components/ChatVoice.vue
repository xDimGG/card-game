<script setup>
import IconChevronDown from './icons/IconChevronDown.vue';
import IconChevronUp from './icons/IconChevronUp.vue';
</script>

<style scoped>
.boxes {
	position: absolute;
	bottom: 0;
	left: 0;
	width: 100%;
	max-width: 400px;
	overflow: hidden;
	z-index: 999;
	border-top-right-radius: 5px;
}

.content {
	background-color: #101010;
	width: 100%;
	padding-bottom: 44px;
	opacity: 0.6;
	transition: opacity 0.2s;
}

.content:hover,
.content:has(textarea:focus) {
	opacity: 1;
}

.messages {
	text-align: left;
	word-wrap: break-word;
	white-space: pre-wrap;
	height: 100%;
	overflow-y: auto;
	padding: 12px;
	height: 180px;
}

.top-bar {
	width: 100%;
	height: 24px;
	background-color: #030303;
	cursor: pointer;
	user-select: none;
	display: flex;
	justify-content: space-between;
	align-items: center;
}

.top-bar > span {
	padding-left: 8px;
	font-size: 14px;
}

.top-bar > svg {
	padding-right: 8px;
	width: 20px;
	height: 20px;
}

.messages > * {
	display: block;
}

.whisper {
	opacity: 0.7;
}

.content > .text {
	position: absolute;
	bottom: 0;
	left: 5px;
	right: 5px;
}

.text > textarea {
	overflow-y: auto;
	margin: 0;
	padding: 4px;
	border: none;
	height: 32px;
	flex: 1;
	width: 100%;
	resize: none;
	outline: none;
}

.users {
	text-align: left;
	padding: 5px 20px;
}

.users > div {
	display: flex;
	justify-content: space-between;
}

.hover-black:hover {
	background-color: black;
}

@media (max-width: 400px) {
	.boxes {
		border-top-right-radius: 0;
	}

	.top-bar {
		height: 18px;
	}

	.top-bar > span {
		font-size: 12px;
	}

	.top-bar > svg {
		padding-right: 8px;
		width: 18px;
		height: 18px;
	}
}
</style>

<template>
	<div class="boxes" v-if="connected">
		<div class="top-bar" @click="voiceOpen = !voiceOpen">
			<span>Voice</span>
			<IconChevronDown v-if="voiceOpen" />
			<IconChevronUp v-else />
		</div>
		<div class="content pb-0" :style="{ display: voiceOpen ? 'block' : 'none' }">
			<div v-if="voiceWs" class="users">
				<div>
					<span style="color: var(--yellow)">{{ data.clients[data.me].name }}</span>
					<span class="text-muted pointer user-select-none" @click="toggleMute(data.me)">
						{{ hasMic[data.me] ? muted[data.me] ? '(unmute)' : '(mute)' : '(no mic)' }}
					</span>
				</div>
				<div v-if="peers" v-for="id in peers.keys()">
					<span :class="speaking[id] ? 'text-white' : 'text-muted'">{{ data.clients[id].name }}</span>
					<span :class="`text-muted user-select-none ${hasMic[id] ? 'pointer' : ''}`" @click="hasMic[id] && toggleMute(id)">
						{{ hasMic[id] ? muted[id] ? '(unmute)' : '(mute)' : '(no mic)' }}
					</span>
				</div>
			</div>
			<button
				@click="voiceWs ? leaveVoice() : joinVoice(data.key)"
				class="hover-black btn btn-md rounded-0 border-0 text-white w-100 h-100">
			{{ voiceWs ? 'LEAVE' : 'JOIN'  }}
			</button>
		</div>
		<div class="top-bar" @click="toggleChat">
			<span>Text{{ unread ? ` (${unread})` : '' }}</span>
			<IconChevronDown v-if="open" />
			<IconChevronUp v-else />
		</div>
		<div class="content" :style="{ display: open ? 'block' : 'none' }">
			<div class="messages" ref="messages" @scroll="checkLock">
				<span class="whisper">You have joined the chat</span>
				<span v-for="{ sender, content, id } in messages"
					:key="id"
					:style="sender === data.me ? 'color: var(--yellow)' : ''">
						{{ `<${data.clients[sender]?.name || 'unknown'}>` }} {{ content }}
				</span>
			</div>
			<div class="text">
				<textarea
					placeholder="Send a message"
					rows="1"
					ref="textarea"
					@keydown="submit"
					@input="resize"
				></textarea>
			</div>
		</div>
	</div>
</template>

<script>
import { nextTick } from 'vue';
import { useToast } from 'vue-toastification';
const toast = useToast();

export default {
	data() {
		return {
			ws: null,
			voiceWs: null,
			connected: false,
			voiceOpen: false,
			open: true,
			messages: [],
			counter: 0,
			unread: 0,
			locked: true,
			peers: new Map(),
			track: null,
			peerAudio: {},
			speaking: {},
			hasMic: {},
			muted: {},
		};
	},
	props: ['data', 'wsBase', 'me'],
	watch: {
		data(newVal = {}) {
			if (newVal.key && !this.connected) this.connect(newVal.key);
			if (!newVal.clients) {
				this.disconnect();
				this.leaveVoice();
			}
		},
	},
	methods: {
		connect(key) {
			if (this.ws) this.disconnect();
			this.ws = new WebSocket(`${this.wsBase}/chat`);
			this.connected = false;

			this.ws.onopen = () => this.ws.send(key);
			this.ws.onclose = () => this.connected = false;
			this.ws.onerror = console.log;

			this.ws.onmessage = async ({ data }) => {
				if (data === 'OK') return this.connected = true;
				const i = data.indexOf(':');

				this.messages.push({
					sender: data.slice(0, i),
					content: data.slice(i + 1, data.length),
					id: this.counter++,
				});

				if (this.open && this.locked) {
					await nextTick();
					this.scrollToBottom();
				} else {
					this.unread++;
				}
			};
		},
		disconnect() {
			this.ws?.close();
			this.ws = null;
			this.connected = false;
			this.messages = [];
		},
		submit(evt) {
			// 13: return/enter
			if (!evt.shiftKey && evt.keyCode === 13) {
				evt.preventDefault();
				const content = evt.currentTarget.value.trim();
				if (content) {
					this.ws.send(content);
					evt.currentTarget.value = '';
					evt.currentTarget.blur();
				}
			}
			this.resize();
		},
		resize() {
			const el = this.$refs.textarea;
			el.style.height = '1px';
			el.style.height = `${Math.min(100, el.scrollHeight)}px`;
		},
		async toggleChat() {
			if (this.open) {
				this.checkLock();
				this.open = false;
			} else {
				this.open = true;
				if (this.locked) {
					await nextTick();
					this.scrollToBottom();
				}
			}
		},
		scrollToBottom() {
			const el = this.$refs.messages;
			el.scrollTop = el.scrollHeight;
			this.unread = 0;
		},
		checkLock() {
			const el = this.$refs.messages;
			if (this.open) {
				this.locked = (el.scrollHeight - el.clientHeight === el.scrollTop);
				if (this.locked) this.unread = 0;
			}
		},
		async onKeyDown(evt) {
			if (this.$refs.textarea !== document.activeElement
				&& (evt.code === 'Tab' || (evt.code === 'Enter' && evt.shiftKey))) {
				evt.preventDefault();
				if (!this.open) this.toggleChat();
				await nextTick();
				this.$refs.textarea.focus();
			}
			if (evt.code === 'Escape') {
				evt.preventDefault();
				this.$refs.textarea.blur()
			}
		},
		async joinVoice(key) {
			if (!key) return toast.error('no key available');
			if (this.voiceWs) return;

			const rtcOptions = {
				iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
			};

			let microphone;
			try {
				microphone = await navigator.mediaDevices.getUserMedia({ audio: true });
				this.myTrack = microphone.getAudioTracks()[0];
			} catch (_) {}

			const initPeer = (id, peer) => {
				this.peers.set(id, peer);

				if (this.myTrack) {
					peer.addTrack(this.myTrack, microphone);
					this.hasMic[this.data.me] = true;
					this.muted[this.data.me] = false;
				}

				peer.addEventListener('track', ({ streams: [stream] }) => {
					const audio = new Audio();
					audio.srcObject = stream.clone();
					audio.play();
					this.peerAudio[id] = audio;
					this.hasMic[id] = true;
					this.muted[id] = false;

					const context = new AudioContext();
					const source = context.createMediaStreamSource(stream);
					const analyzer = context.createAnalyser();
					analyzer.fftSize = 2 ** 13; // Larger sample window
					source.connect(analyzer);

					// The array we will put sound wave data in
					const array = new Uint8Array(analyzer.fftSize);

					function getPeakLevel() {
						analyzer.getByteTimeDomainData(array);
						return array.reduce((max, current) => Math.max(max, Math.abs(current - 127)), 0) / 128;
					}

					setInterval(() => {
						if (getPeakLevel() > 0.01) {
							this.speaking[id] = true;
						} else {
							this.speaking[id] = false;
						}
					}, 200);
				});
				peer.addEventListener('icecandidate', event => {
					if (event.candidate) ws.send(`${id}${JSON.stringify(event.candidate)}`);
				});
				peer.addEventListener('connectionstatechange', () => {
					if (['closed', 'failed'].includes(peer.connectionState)) {
						this.peers.delete(id);
					}
				});
			};

			const ws = new WebSocket(`${this.wsBase}/voice`);
			navigator.mediaDevices.getUserMedia({ audio: true });
			this.voiceWs = ws;
			ws.onopen = () => ws.send(key);
			ws.onmessage = async ({ data }) => {
				const UUID_LEN = 36;
				// If we receive some JSON
				if (data[UUID_LEN] === '{') {
					const sender = data.slice(0, UUID_LEN)
					const payload = JSON.parse(data.slice(UUID_LEN));
					console.log(sender, payload);

					if (payload.type === 'offer') {
						if (this.peers.has(sender)) return;

						const peer = new RTCPeerConnection(rtcOptions);
						peer.setRemoteDescription(new RTCSessionDescription(payload));
						initPeer(sender, peer);

						const answer = await peer.createAnswer({ offerToReceiveAudio: true });
						await peer.setLocalDescription(answer);

						return ws.send(`${sender}${JSON.stringify(answer)}`);
					}

					if (!this.peers.has(sender)) return;
					const peer = this.peers.get(sender);

					if (payload.type === 'answer') {
						await peer.setRemoteDescription(new RTCSessionDescription(payload));
					} else if (payload.candidate) {
						try {
							await peer.addIceCandidate(payload);
						} catch (e) {
							console.error('Error adding received ice candidate', e);
						}
					} else if (payload.disconnected) {
						this.peers.delete(sender);
						peer.close();
					}
				} else { // If we receive just a UUID
					const peer = new RTCPeerConnection(rtcOptions);
					const offer = await peer.createOffer({ offerToReceiveAudio: true });
					await peer.setLocalDescription(offer);
					ws.send(`${data}${JSON.stringify(offer)}`);
					initPeer(data, peer);
				}
			};
			ws.onerror = console.log;
			ws.onclose = () => this.leaveVoice();
		},
		leaveVoice() {
			for (const [id, peer] of this.peers.entries()) {
				if (this.voiceWs.readyState === WebSocket.OPEN) {
					this.voiceWs.send(`${id}{"disconnected":true}`);
				}
				peer.close();
			}
			this.myTrack = null;
			this.peers = new Map();
			this.peerAudio = {};
			this.speaking = {};
			this.muted = {};
			this.hasMic = {};

			if (this.voiceWs) {
				this.voiceWs.close();
				this.voiceWs = null;
			}
		},
		async toggleMute(id) {
			if (id === this.data.me) {
				this.myTrack.enabled = this.muted[id];
			} else {
				this.peerAudio[id].muted = !this.muted[id];
			}
			this.muted[id] = !this.muted[id];
		},
	},
	mounted() {
		window.addEventListener('keydown', this.onKeyDown);
		if (this.data.key && !this.connected) this.connect(this.data.key);
	},
	unmounted() {
		window.removeEventListener('keydown', this.onKeyDown);
	},
}
</script>