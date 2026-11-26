// Bets: how each strategy has done over every settled bet, for all streamers.
// The section stays hidden until there is a bet to show.
let betsSig = '';
function renderBets(d) {
  const rows = d.strategies || [];
  const sig = JSON.stringify(rows);
  if (sig === betsSig) return;
  betsSig = sig;
  $('bets').hidden = !rows.length;
  const total = rows.reduce((n, s) => n + s.bets, 0);
  text($('bets-n'), fmt(total) + (total === 1 ? ' bet' : ' bets'));
  $('bets-rows').replaceChildren(...rows.map(s => {
    const decided = s.won + s.lost;
    const cell = (value, className) => el('td', { className: 'num ' + (className || ''), textContent: value });
    return el('tr', null,
      el('th', { scope: 'row', textContent: s.strategy || 'Unknown' }),
      cell(fmt(s.bets)),
      cell(fmt(s.won)),
      cell(fmt(s.lost)),
      cell(fmt(s.refunded)),
      cell(decided ? Math.round(100 * s.won / decided) + '%' : '–'),
      cell(fmt(s.staked)),
      cell(signed(s.net), cls(s.net)));
  }));
}
async function loadBets() {
  try { renderBets(await getJSON('api/bets')); } catch { /* the next refresh tries again */ }
}
