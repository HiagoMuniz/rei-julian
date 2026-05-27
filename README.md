# 👑 Rei Julian - Ecossistema de Rastreamento Logístico

> **Disclaimer:** Este projeto foi desenvolvido para a disciplina de **TEC IV** (UFPel). O grande diferencial aqui é a abordagem: estou explorando até onde consigo chegar usando **Inteligência Artificial** como copiloto total, focando na execução rápida e vendo o que acontece quando eu "paro de pensar muito" e deixo a IA estruturar o código e a arquitetura.

## 🚀 Sobre o Projeto
O "Rei Julian" é uma Prova de Conceito (PoC) de um sistema de rastreamento geográfico em tempo real. Ele simula uma frota de entregadores (atualmente o **Julian** e o **Mort**) se movendo pelas ruas de Pelotas-RS, transmitindo suas coordenadas instantaneamente para um dashboard.

A ideia é validar uma infraestrutura capaz de suportar um sistema de delivery real, onde o restaurante monitora seus motoboys e o cliente acompanha seu pedido.

## 🛠️ Tech Stack
- **Backend:** [Go (Golang)](https://go.dev/) - Escolhido pela performance e facilidade com concorrência.
- **Real-time:** SSE (Server-Sent Events) - Para streaming de dados sem o overhead de WebSockets.
- **Frontend:** HTML5, CSS moderno e [Leaflet.js](https://leafletjs.com/) para os mapas.
- **Mapas:** OpenStreetMap.

## 📍 Funcionalidades Atuais
- [x] **Multi-Rastreio Simulado:** Frota de entregadores (Julian, Mort, Maurice, Clover) com movimentação autônoma.
- [x] **Dashboard Profissional:** Interface moderna com sidebar, cards de status e mapa interativo.
- [x] **Persistência em Banco de Dados:** Histórico de rotas salvo em SQLite em tempo real.
- [x] **Visualização de Histórico:** Rastro (trail) visual no mapa para cada entregador.
- [x] **Streaming SSE:** Atualizações instantâneas sem refresh.

## 🏃 Como rodar
1. Tenha o **Go** instalado na sua máquina.
2. Clone o repositório.
3. No terminal, dentro da pasta do projeto, execute:
   ```bash
   go run cmd/rei-julian/main.go
   ```
4. Abra o navegador em: `http://localhost:8080`

## 🗺️ Roadmap Experimental (Próximos Passos)
A ideia é continuar "terceirizando" a lógica para a IA para implementar:
1. **Persistência com SQLite:** Para não perder o rastro toda vez que o servidor reiniciar.
2. **Ingestão via API:** Permitir que um celular real envie sua posição via POST.
3. **Cálculo de ETA:** Usar APIs de roteamento para prever em quanto tempo o "Rei Julian" chega com a sua pizza.

---
**Desenvolvido por:** [Hiago Dona Muniz](https://github.com/hiagomuniz)
**Contexto:** Disciplina de TEC IV - Ciência da Computação (UFPel).
