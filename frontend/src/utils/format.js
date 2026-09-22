const integerFormat = new Intl.NumberFormat('pt-BR', { maximumFractionDigits: 0 });

const currencyFormat = new Intl.NumberFormat('pt-BR', {
  style: 'currency',
  currency: 'BRL',
});

const numberFormat = new Intl.NumberFormat('pt-BR', { maximumFractionDigits: 2 });

export function formatPopulation(indicator) {
  if (!indicator) return '—';
  return integerFormat.format(indicator.value);
}

export function formatInteger(value) {
  if (value == null) return '—';
  return integerFormat.format(value);
}

export function formatIncome(indicator) {
  if (!indicator) return '—';
  return currencyFormat.format(indicator.value);
}

export function formatGDP(indicator) {
  if (!indicator) return '—';
  return numberFormat.format(indicator.value);
}
