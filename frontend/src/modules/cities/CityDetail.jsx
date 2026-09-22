import { Link, useLocation, useParams } from 'react-router-dom';
import { ArrowLeft, BarChart2, MapPin, RefreshCw } from 'lucide-react';
import '../shared.css';
import { DataGrid } from '../../app/components/DataGrid';
import { Panel } from '../../app/components/Panel';
import { useApiResource } from '../../hooks/useApiResource';
import { getCityDetail } from '../../services/cities';
import { formatGDP, formatIncome, formatPopulation } from '../../utils/format';

export default function CityDetail() {
  const { ibgeCode } = useParams();
  const location = useLocation();

  const cityRes = useApiResource(
    ({ signal }) => getCityDetail(ibgeCode, { signal }),
    [ibgeCode],
  );

  const city = cityRes.data;
  const cityMatches = city ? String(city.ibge_code) === String(ibgeCode) : false;

  const identificationColumns = [
    {
      label: 'Código IBGE',
      field: 'ibge_code',
      width: '30%',
      render: (value) => String(value),
    },
    { label: 'Estado', field: 'stateName', width: '40%' },
    { label: 'Região', field: 'region', width: '30%' },
  ];

  const identificationRows = cityMatches
    ? [{
        ibge_code: city.ibge_code,
        stateName: `${city.state.name} (${city.state.uf})`,
        region: city.state.region,
      }]
    : [];

  const indicatorRows = cityMatches
    ? [
        {
          label: 'População',
          year: city.indicators?.population?.year ?? '—',
          value: formatPopulation(city.indicators?.population),
        },
        {
          label: 'Renda média',
          year: city.indicators?.income?.year ?? '—',
          value: formatIncome(city.indicators?.income),
        },
        {
          label: 'PIB (Mil R$)',
          year: city.indicators?.gdp?.year ?? '—',
          value: formatGDP(city.indicators?.gdp),
        },
      ]
    : [];

  const indicatorColumns = [
    { label: 'Indicador', field: 'label', width: '35%' },
    { label: 'Ano', field: 'year', width: '15%' },
    { label: 'Valor', field: 'value', width: '50%' },
  ];

  return (
    <div className="module-container">
      <div className="module-main">
        <div className="module-toolbar">
          <div className="module-toolbar-title">
            {cityMatches ? `${city.name} — ${city.state.uf}` : 'Cidade'}
          </div>
          <div className="module-toolbar-actions">
            <Link
              className="btn btn--secondary"
              to="/cidades"
              state={{ ibge: location.state?.ibge }}
            >
              <ArrowLeft size={14} />
              Voltar
            </Link>
            <button
              type="button"
              className="btn btn--secondary"
              onClick={cityRes.reload}
              disabled={cityRes.loading}
              title="Recarregar da API"
            >
              <RefreshCw size={14} className={cityRes.loading ? 'spin' : undefined} />
              Atualizar
            </button>
          </div>
        </div>

        <div className="module-content">
          {cityRes.loading && !cityMatches ? (
            <div className="loading-box">Carregando cidade…</div>
          ) : cityRes.error ? (
            <div className="error-box">
              {cityRes.error.message}
              <button type="button" className="error-retry" onClick={cityRes.reload}>
                Tentar novamente
              </button>
            </div>
          ) : cityMatches ? (
            <>
              <Panel title="Identificação" icon={<MapPin size={16} />}>
                <DataGrid
                  columns={identificationColumns}
                  data={identificationRows}
                  selectable={false}
                />
              </Panel>
              <Panel title="Indicadores" icon={<BarChart2 size={16} />}>
                <DataGrid columns={indicatorColumns} data={indicatorRows} selectable={false} />
              </Panel>
            </>
          ) : null}
        </div>
      </div>
    </div>
  );
}
