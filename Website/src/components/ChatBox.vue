<style scoped>
.chatbox-container {
	position: absolute;
	bottom: 0;
	left: 0;
	width: 100%;
	max-width: 350px;
	overflow: hidden;
	z-index: 999;
	border-top-right-radius: 5px;
	opacity: 0.6;
	transition: opacity 0.2s;
}

.chatbox-container:hover,
.chatbox-container:has(textarea:focus) {
	opacity: 1;
}

.chatbox {
	background-color: #101010;
	width: 100%;
	padding-bottom: 44px;
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
	display: flex;
	justify-content: space-between;
	align-items: center;
}

.top-bar > span {
	padding-left: 8px;
	font-size: 14px;
}

.top-bar svg {
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

.chatbox > .text {
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
</style>

<template>
	<div class="chatbox-container" v-if="connected">
		<div class="top-bar" @click="toggleChat">
			<span>Chat{{ unread ? ` (${unread})` : '' }}</span>
			<IconChevronDown v-if="open" />
			<IconChevronUp v-else />
		</div>
		<div class="chatbox" :style="{ display: open ? 'block' : 'none' }">
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
import IconChevronDown from './icons/IconChevronDown.vue';
import IconChevronUp from './icons/IconChevronUp.vue';

export default {
	data() {
		return {
			ws: null,
			connected: false,
			open: true,
			messages: [],
			counter: 0,
			unread: 0,
			locked: true,
		};
	},
	props: ['data', 'wsBase', 'me'],
	watch: {
		data(newVal = {}) {
			if (newVal.key && !this.connected) this.connect(newVal.key);
			if (!newVal.clients) this.disconnect();
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
			this.connected = false;
			this.messages = [];
		},
		submit(evt) {
			if (!evt.shiftKey && evt.code === 'Enter') {
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
		onKeyDown(evt) {
			if (this.$refs.textarea !== document.activeElement
				&& (evt.code === 'Tab' || (evt.code === 'Enter' && evt.shiftKey))) {
				evt.preventDefault();
				if (!this.open) this.toggleChat();
				this.$refs.textarea.focus();
			}
			if (evt.code === 'Escape') {
				evt.preventDefault();
				this.$refs.textarea.blur();
			}
		}
	},
	mounted() {
		window.addEventListener('keydown', this.onKeyDown);
	},
	unmounted() {
		window.removeEventListener('keydown', this.onKeyDown);
	},
	components: { IconChevronDown, IconChevronUp }
}
</script>