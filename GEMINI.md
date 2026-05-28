# Relatório de Projeto: Rei Julian - Ecossistema de Rastreamento Logístico

**Disciplina:** TEC IV
**Aluno:** HIAGO DONA MUNIZ

---

## 1. Introdução e Propósito
O projeto "Rei Julian" nasceu como uma Prova de Conceito (PoC) para um sistema de rastreamento geográfico em tempo real. Embora a versão atual opere em um ambiente de simulação matemática, o objetivo final é fornecer uma solução robusta e acessível para o gerenciamento logístico de estabelecimentos comerciais, com foco inicial no setor de delivery gastronômico em Pelotas-RS.

## 2. Fase Atual: Simulação Multirrastreio
Atualmente, o sistema valida a arquitetura de comunicação necessária para gerenciar múltiplos agentes simultaneamente:
1.  **Multi-Rastreio:** O sistema agora suporta múltiplos indivíduos (ex: "Julian" e "Mort"), cada um com seu próprio ID, simulando uma frota de entregadores.
2.  **Arquitetura Reativa (SSE):** Utiliza *Server-Sent Events* para transmitir atualizações de múltiplos IDs em um único canal, garantindo sincronia no mapa.
3.  **Interface Identificada:** Marcadores no mapa Leaflet agora exibem labels permanentes com o nome do indivíduo, facilitando a distinção visual em tempo real.

## 3. Limites e Desafios de Implementação
Para a transição do protótipo para o produto final, os seguintes desafios técnicos foram mapeados:
*   **Persistência de Dados (Resolvido):** Transição da memória volátil para um banco de dados relacional (SQLite) para armazenamento de históricos de rotas e auditoria de entregas.
*   **Integração de Mapas Reais:** Substituição da movimentação aleatória por APIs de roteamento (como OSRM ou Google Routes) que respeitem a malha viária urbana.
*   **Segurança e Privacidade:** Implementação de camadas de autenticação e criptografia para garantir que os dados de localização sejam acessíveis apenas às partes autorizadas.

## 4. Evolução: O Sistema de Delivery Real
A próxima etapa consiste na aplicação prática desta tecnologia em um cenário de restaurante local:
*   **Captação de Dados Reais:** Substituição do simulador por uma aplicação mobile que envie coordenadas GPS em tempo real dos smartphones dos motoboys.
*   **Painel de Gestão:** Uma interface administrativa para o restaurante monitorar múltiplos entregadores simultaneamente, permitindo uma distribuição de pedidos mais eficiente.

## 5. Visão de Futuro e Impacto Logístico
O "Rei Julian" projeta-se como uma ferramenta de inteligência competitiva para o comércio local:
*   **Transparência para o Cliente:** Envio de links temporários para que o cliente final acompanhe o pedido, aumentando a confiança e reduzindo a ansiedade durante a espera.
*   **Otimização de Tempos (ETA):** Cálculo automático do tempo estimado de chegada com base na velocidade média e tráfego local.
*   **Escalabilidade Setorial:** Embora focado em restaurantes, a estrutura está sendo desenhada para ser adaptável a qualquer serviço de entrega ou manutenção em domicílio, democratizando o acesso a tecnologias de rastreamento de alta performance para pequenos e médios empreendedores.
