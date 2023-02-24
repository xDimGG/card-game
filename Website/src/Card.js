export default class Card {
	constructor(name, hidden = false, onclick = null, style = {}) {
		this.name = name;
		this.hidden = hidden;
		this.onclick = onclick;
		this.style = style;
	}

	get clickable() {
		return this.onclick instanceof Function;
	}
}