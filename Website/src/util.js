// Prepares array of ordered id's to be rendered
export const reorder = ps => {
	if (ps.length <= 2) return ps;
	if (ps.length === 3) return [ps[0], ps[2], ps[1]];
	if (ps.length === 4) return [ps[0], ps[2], ps[1], ps[3]];
};