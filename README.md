# TRADS — Painel de Inteligência de Mercado

## 1. O que é

Aplicação fullstack desenvolvida para o desafio técnico de **Programador Júnior I da Trads Corretora** (vaga: https://github.com/Trads-Corretora/teste-junior). Ela consome a **API pública do IBGE**, persiste os dados em um **PostgreSQL próprio** e serve um **dashboard** que responde à pergunta de negócio do desafio: *em quais regiões do Brasil estão os melhores mercados, e para qual público?*

**Stack:** Go 1.26 (backend, sem framework HTTP) · PostgreSQL 16 (banco) · React 19 + Vite (frontend) · Docker Compose (infraestrutura local) · dados oficiais IBGE/SIDRA.

---

## 2. Como rodar localmente (do zero)

Pré-requisitos: **Docker**. O caminho rápido (2.1) dispensa Go e Node; o modo desenvolvimento (2.2 em diante) pede ainda **Go 1.26+** e **Node 20.19+**.

### 2.1 Caminho rápido — um comando (recomendado na avaliação)

```bash
git clone https://github.com/GabrielAthenaOffice/teeeesteeeeee.git
cd teeeesteeeeee

docker compose up --build -d
docker compose logs -f import
```

O compose orquestra a cadeia **Postgres (saudável) → migrations → importador do IBGE (1ª execução, até 5 min) → API → frontend**. Quando o import terminar:

| O quê | Onde |
| --- | --- |
| Frontend (React + nginx) | http://localhost:8085 |
| API Go | http://localhost:8082 |
| Adminer (opcional) | http://localhost:8084 |

Nesse caminho **nenhum `.env` é necessário** — as variáveis vêm do próprio `docker-compose.yml`. Checkpoint: `curl http://localhost:8085/api/v1/states` deve devolver as **27 UFs** (o nginx do frontend faz proxy de `/api` e `/health` para o serviço `api`). Para reimportar os dados: `docker compose run --rm import`. Para parar tudo: `docker compose down`.

### 2.2 Modo desenvolvimento — banco e migrations

Para desenvolver na máquina (Vite com hot reload em 5173 e `go run`):

```bash
docker compose up -d postgres
docker compose run --rm migrate
```

O Adminer (opcional — interface web do banco, porta **8084**) sobe com `docker compose up -d adminer`, que por dependência também sobe o Postgres.

> Atenção: a API da máquina (`go run`) e o container `api` disputam a **8082** — use um por vez (`docker compose stop api` antes de `go run ./cmd/api`).

### 2.3 Crie as variáveis de ambiente do backend

No modo desenvolvimento os `.env` são necessários (no caminho rápido, não). Eles são **gitignored** (segredos não vão para o repositório). Crie `backend-golang/.env`:

```ini
PORT=8082
APP_ENV=development
DATABASE_URL=postgres://trads:trads@localhost:5444/trads?sslmode=disable
BASE_URL_IBGE=https://servicodados.ibge.gov.br/api/v3
BASE_URL_LOCALIDADES=https://servicodados.ibge.gov.br/api/v1
```

### 2.4 Importe os dados do IBGE (uma vez)

```bash
cd backend-golang
go run ./cmd/import
```

O importador busca estados, municípios e indicadores nas APIs oficiais e grava no banco (timeout interno de 5 minutos). Ao terminar, o banco tem **27 estados, 5.571 municípios** e os indicadores de população (Censo 2022), renda (2022) e PIB (2023). A partir daqui, **a API não bate no IBGE por requisição de usuário** — só no seu banco (requisito 3 do desafio).

### 2.5 Suba a API

```bash
go run ./cmd/api
```

Checkpoint: `curl http://localhost:8082/health` deve responder **200**.

### 2.6 Suba o frontend (outra janela de terminal)

```bash
cd frontend
npm install
npm run dev
```

Abra **http://localhost:5173**. Em desenvolvimento o Vite faz proxy de `/api` e `/health` para `http://localhost:8082` (sem CORS). Para apontar outra base, crie `frontend/.env.local` com `VITE_API_PROXY_TARGET` (opcional — o padrão já é 8082).

### Portas

| Serviço | Porta |
| --- | --- |
| PostgreSQL | 5444 |
| API Go | 8082 |
| Frontend — caminho rápido (nginx/Docker) | 8085 |
| Frontend — desenvolvimento (Vite) | 5173 |
| Adminer (opcional) | 8084 |

---

## 3. Arquitetura

```
backend-golang/
  Dockerfile           # imagem multi-stage com os binários `api` e `import`
  cmd/api/             # wiring: config → repositórios → use cases → HTTP
  cmd/import/          # importador IBGE → Postgres (executar 1x)
  internal/core/       # domain (entidades), ports (interfaces), usecases (regras)
  internal/adapter/    # http (handlers/DTOs), postgres (SQL), ibge (client oficial)
  db/migrations/       # schema versionado (6 migrations)
frontend/
  Dockerfile           # build do Vite + nginx
  nginx.conf           # estático + proxy de /api e /health para o serviço `api`
  src/app/             # design-system, layout, componentes globais (DataGrid, Panel…)
  src/modules/         # páginas: dashboard, states, cities (grid + detalhe)
  src/services/        # cliente HTTP — um arquivo por recurso, contratos do backend
  src/hooks/           # useApiResource: loading/erro/cancelamento/reload
  src/utils/           # formatação pt-BR de indicadores (null → “—”)
docker-compose.yml     # postgres, migrate, import, api, frontend, adminer
```

**Fluxo:** `cmd/import` consome o IBGE (localidades **v1**) e o SIDRA (agregados **v3**) → persiste com procedência (ano + fonte por indicador) → a API serve **somente do banco** → o React consome via proxy.

**Superfície da API** (todas `GET`):

| Caminho | Descrição |
| --- | --- |
| `/health`, `/health/db` | vida da aplicação e do banco |
| `/api/v1/states` | 27 UFs |
| `/api/v1/states/{ibgeCode}/cities?page&pageSize` | municípios da UF **com indicadores** (paginação, `pageSize` máx. 100) |
| `/api/v1/cities/{ibgeCode}` | detalhe do município (`state` + indicadores; 404 `city_not_found`) |
| `/api/v1/dashboard/national` | agregado nacional (municípios + população/renda/PIB com ano) |
| `/api/v1/dashboard/states` | os 27 estados com agregados |
| `/api/v1/dashboard/top-cities?limit` | top PIB, renda e população numa resposta (limit 1–100) |

Erros seguem o contrato `{"error":{"code","message"}}` com códigos estáveis (`invalid_request`, `state_not_found`, `city_not_found`, `not_found`, `internal_error`); a única exceção é 405, que sai em texto puro do `net/http`.

---

## 4. Decisões técnicas e trade-offs

- **Go + `net/http` padrão, sem framework.** O `ServeMux` do Go já resolve rota por método e path params. Sem mágica: handlers, DTOs e erros escritos à mão e legíveis de ponta a ponta — importa porque o avaliador consegue seguir o código inteiro sem aprender uma framework.
- **PostgreSQL + SQL explícito (pgx), sem ORM.** As queries ficam visíveis no repositório e as garantias vivem no banco: `FOREIGN KEY`, `UNIQUE (city_id, year, source)`, `CHECK (value >= 0)` (população `-1` é impossível de gravar), `CHECK` de ano. Trade-off: mapeamento manual de linhas → DTOs.
- **Migrations versionadas** (`golang-migrate` via imagem oficial no compose). Ninguém roda SQL solto; o schema sobe igual em qualquer máquina.
- **Arquitetura em camadas** `core` (domain/ports/usecases) → `adapter` (http/postgres/ibge): as regras de negócio não dependem de framework, o que destrava testes de unidade depois. A inversão das dependências (adapter injetando o core, e não o contrário) é a próxima etapa de refatoração — registrada nas limitações.
- **Frontend React + Vite sem UI kit e sem lib de gráficos.** As visualizações (barras de região, rankings, tabelas) são HTML/CSS próprios sobre os componentes do projeto: bundle enxuto, zero dependência pesada, tudo auditável. Trade-off: menos gráficos "prontos" do que Chart.js/Recharts.
- **Escolha dos indicadores** (faz parte da avaliação, segundo o enunciado):

  | Indicador | Fonte oficial verificada | Ano | Unidade | Por quê |
  | --- | --- | --- | --- | --- |
  | População | SIDRA 4709, variável 93 — "População residente" | 2022 (Censo) | Pessoas | tamanho do mercado por município/UF |
  | Renda média | SIDRA 10295, variável 13431 — "rendimento nominal médio mensal domiciliar per capita" | 2022 | Reais (máx. 2 decimais) | capacidade de compra para plano de saúde/vida/odonto |
  | PIB | SIDRA 5938, variável 37 — "PIB a preços correntes" | 2023 | **Mil Reais** (unidade oficial da variável) | dinamismo econômico regional |

  Os nomes/unidades acima foram lidos na resposta da própria API do SIDRA, não inferidos. **Faixa etária** é o indicador natural seguinte para "para qual público" (planos são sensíveis à idade) e está pendente — ver limitações.
- **Validação por oráculo:** cada agregação foi conferida célula a célula contra consultas SQL independentes (Σ população = **203.080.756** = Censo 2022 nacional; renda ponderada recomposta = 1.638,60).

---

## 5. A herança: análise do legado

A pasta [`legado/`](https://github.com/Trads-Corretora/teste-junior/tree/main/legado) do desafio contém a tentativa anterior (PHP + HTML/jQuery + CSV + bilhete do ex-funcionário). Seguindo a instrução do enunciado, **cada afirmação foi tratada como hipótese e verificada** — contra o código, contra a API oficial do IBGE ao vivo e contra o nosso banco validado.

### 5.1 O que a tentativa anterior fazia (ou dizia que fazia)

- `coleta_ibge.php`: buscava o PIB por UF na API do IBGE, aplicava um "filtro da diretoria" de renda mínima e supostamente gravava em MySQL (tabela `estados`).
- `config.php`: credenciais do MySQL.
- `dados_exportados.csv`: "backup" de municípios (código, município, renda, população).
- `painel_antigo.html`: painel estático com tabela de municípios ordenada, o mesmo "filtro de 2 salários mínimos" e um botão "Atualizar dados" que chamaria a API.
- `LEIA-ME-primeiro.txt`: bilhete alegando "90% pronto, é só rodar", banco funcionando e CSV "com números certinhos".

### 5.2 O que está errado ou não é confiável

| Alegação | Veredito | Evidência |
| --- | --- | --- |
| "Já salva no banco, é só rodar" | falsa | `salvar_no_banco()` só escreve `coleta_debug.txt`; o `INSERT` prometido nunca foi escrito ("depois eu ligo o insert") |
| "Roda hoje" | falsa | `mysql_connect` foi removido desde o PHP 7 (2015) → fatal error em PHP moderno |
| "Endpoint v9, o mais atual" | falsa | Teste ao vivo: `v1 → 200`, `v9 → 000` (sem resposta). O oficial é o **v1** de localidades |
| "Números certinhos e atualizados de 2020" | falsa | Contra o nosso banco validado: **4 de 10 populações erradas** (Fortaleza +274.683, Brasília +276.944, Curitiba +190.008, SP −754); Manaus como string `"1.234.567"`; Porto Alegre = `-1`; município fantasma `9999999`; João Pessoa **duplicado**; **renda: 0 de 10 batem** (SP 3200 vs. 2.713,36 oficial); sem coluna de ano → procedência misturada |
| "Coluna 3 = população, coluna 4 = renda" | falsa | O próprio header diz o contrário (`renda;populacao_toal`) — bilhete contradiz CSV |
| "Filtro da diretoria já pronto" | inútil | A variável 37 da tabela 5938 é **PIB total em Mil Reais** (verificado na resposta oficial: SP/2020 = 2.377.638.980). Comparar PIB total contra `2824` ("2 SM") **não filtra nenhuma UF** (menor PIB estadual é ordens de grandeza maior); filtra UF enquanto a regra fala município; `2824` = 2× salário mínimo de **2024** com dado "de 2020"; PIB ≠ renda |
| "Abaixo de 2 SM o IBGE não tem dado confiável" | falsa | Regra de negócio disfarçada de limitação técnica — o SIDRA publica renda para todos os municípios |
| "Abre e já mostra a tabela" | falsa | O `ready` chama `montatabela()` mas a função é `montaTabela` → `ReferenceError`, nada renderiza |
| "O botão Atualizar puxa da API" | quebra a tela | Substitui os dados pelo shape de `/estados` (sem população/renda) → colunas `undefined` |
| "Ordena por população" | falsa | Ordena pela coluna da **renda** e filtra pela **população** (trocadas); o comparador retorna boolean, então a ordenação é inválida |
| Banca/CSV/painel são os mesmos dados | falsas | 3 fontes desconectadas (`coleta_debug.txt`, CSV, cache hardcoded no JS) — o painel **nunca** leu banco nem CSV |
| Credenciais seguras | violadas | Senha hardcoded + TODO "tirar a senha antes do git" — presente em repositório público |
| CSV confiável | corrompido | Typo no header (`populacao_toal`), mojibake (`JoÃ£o Pessoa`), linha duplicada, coluna extra (`;extra`), separador inconsistente com o bilhete |
| "Só falta 10%" | falsa | Sem persistência, sem tratamento de erro (falha no `file_get_contents` e o script imprime "Coleta finalizada com sucesso!") — o plano de erro documentado é "rode de novo" |

Além disso, o produto nunca teve nível territorial definido: o coletor fala em **UF**, o CSV e o painel em **município**, a regra em **município**, a tabela prometida em `estados` — e o "painel por região" não agrupa por região nenhuma.

### 5.3 Por que descartamos e como esta solução resolve

Descarte integral, como o desafio manda (sem continuar o PHP nem usar as ferramentas dele):

- **Persistência real:** `cmd/import` grava em PostgreSQL com migrations, constraints (`CHECK/UNIQUE/FK`) e **ano + fonte por indicador** — o `INSERT` que não existia, existe; o `-1` e o duplicado são impossíveis de gravar.
- **Fontes verificadas:** usamos os endpoints oficiais confirmados ao vivo (v1 de localidades; SIDRA 4709/93, 10295/13431, 5938/37) com ano e unidade corretos — nada de "v9" nem PIB confundido com renda.
- **Qualidade conferida:** validação célula a célula contra oráculo SQL independente; Σ população = Censo 2022 (203.080.756). O município sem dado oficial (5101837, Boa Esperança do Norte/MT) vira `null` explícito na API e "—" na tela — não fantasma, não `-1`.
- **Erros honestos:** contrato JSON com códigos estáveis, em vez de "sucesso" impresso sobre falha.
- **Segredos fora do código:** `.env` gitignored (o oposto do `config.php` legado).
- **Dashboard de verdade:** cards nacionais, rankings (top PIB/renda/população), tabela das 27 UFs e filtros que atualizam as views — o que o `painel_antigo.html` prometia mas não entregava.

**Suposição registrada (regra dos 2 salários mínimos):** a "regra da diretoria" citada no bilhete **não tem fonte confirmada** além dele mesmo, e sua implementação legada era um filtro morto sobre PIB total (ver 5.2). Tratamos a regra como **hipótese de produto não verificada**: ela não foi implementada e não afetamos os dados por ela. Se a diretoria confirmar a regra, o caminho natural é um filtro de renda mínima no grid de cidades — ver limitações.

---

## 6. Núcleo vs. extras

**Núcleo (obrigatórios do desafio):**

| # | Obrigatório | Status | Onde |
| --- | --- | --- | --- |
| 1 | Análise do legado no README | ✅ | seção 5 |
| 2 | Integração com a API do IBGE | ✅ | `internal/adapter/ibge` (`cmd/import`) |
| 3 | Banco próprio, sem bater no IBGE por request | ✅ | importador + API lendo só o Postgres |
| 4 | Consultas e filtros | ✅ | paginação, filtro por UF, filtro/busca por região, rankings ordenados |
| 5 | Dashboard com ≥2 visuais + filtros | ✅ | cards, barras por região, 3 rankings, tabela de UFs, 2 grids com filtros |
| 6 | Docker | ✅ | `docker compose up --build` sobe banco → migrations → import → API → frontend (nginx) sem instalar Go/Node — seção 2.1 |
| 7 | README com decisões | ✅ | esta seção e as vizinhas |
| 8 | Repo público com histórico | ✅ | 77+ commits incrementais |

**Extras (além do obrigatório):** health checks de app e banco; contrato de erro rico com códigos; paginação com clamp (`pageSize` ≤ 100); read-model agregado 3-em-1 (`/dashboard/top-cities`); listagem de municípios já enriquecida com indicadores; rota de detalhe por município; formatação pt-BR com `Intl` e tratamento explícito de ausência (`null` → `—`); Adminer no compose; validação automatizada célula a célula contra SQL.

---

## 7. Limitações conhecidas e o que faria com mais tempo

- **Faixa etária (`age_indicators`):** tabela existe mas está vazia (nenhuma importação). Próximo passo: investigar no SIDRA a tabela, nível territorial (N6), período e unidade corretos → importar → widget de distribuição etária. É o indicador que mais responde "para qual público".
- **Filtro "renda ≥ 2 salários mínimos":** não implementado (suposição não confirmada — seção 5.3). Com a confirmação da regra, vira filtro de renda mínima no grid de cidades.
- **Busca por município:** decisão de escopo — não há busca textual de município; a navegação é por UF + paginação e pelo ranking. Com mais tempo: endpoint com `ILIKE`/índice e busca no frontend.
- **Testes automatizados e CI:** ainda não existem. Próximo passo: testes dos use cases (para isso inverter a dependência `core → adapter`) e um GitHub Actions com `go vet/test` + `oxlint` + builds.
- **Deploy:** não há link no ar; a ordem é terminar o Docker full-stack e então publicar (o README ganha o link).
- **Atualização agendada dos dados:** hoje o import é manual (`go run ./cmd/import`); com mais tempo, vira job agendado (cron/K8s) — os dados mudam por ano, não por hora.
- **Autenticação:** não existe — decisão consciente (dados públicos do IBGE, API somente leitura). Se a API for exposta na internet, o mínimo é rate limit e CORS restrito.
- **Dado ausente:** 5101837 (Boa Esperança do Norte/MT) não tem registro nas fontes consultadas para os três indicadores; a API devolve `null` e a UI mostra `—`, sem inventar número.
- **Refatoração arquitetural:** inverter `core → adapter` (hoje os use cases conhecem as ports; o ideal é os adapters registrarem implementações) para destravar a suíte de testes.

---

## 8. Como usamos IA

O projeto foi desenvolvido em parceria com uma IA (agente de código **OpenCode**), comigo dirigindo a ferramenta, não o contrário:

- **Direção:** dividi o trabalho por sessões (banco → importador → endpoints → agregações → frontend), com **pesquisa somente-leitura antes de qualquer edição** e validação explícita minha depois de cada etapa relevante.
- **O que aceitei:** a arquitetura em camadas, o contrato dos endpoints e os widgets do dashboard foram propostos pela IA e aprovados por mim (inclusive delegações de "melhor prática", como o read-model 3-em-1 do ranking).
- **O que ajustei:** porta da API mudou de 8081 para 8082 por conflito local; regra de casa de **zero comentários novos no código** (explicação só nas respostas); commits feitos por mim, com histórico incremental; correções de contrato no cliente frontend apontadas por mim na revisão do diff.
- **Verificação:** não confiei em alegação sem testar — as evidências da seção 5 vieram de chamadas ao vivo contra a API do IBGE e de consultas SQL contra o nosso banco, e cada etapa de backend teve script de validação (célula a célula contra oráculo independente) executado antes de avançar.
