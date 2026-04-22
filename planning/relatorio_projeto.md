# Relatório de Projeto: Rei Julian - Sistema de Rastreamento em Tempo Real

**Disciplina:** TEC IV
**Aluno:** HIAGO DONA MUNIZ

---

## 1. Introdução
O projeto "Rei Julian" consiste em uma aplicação simplificada para rastreamento geográfico em tempo real. O sistema simula o movimento de um objeto (ou pessoa) partindo de um ponto inicial fixo (Praça Coronel Pedro Osório, Pelotas-RS) e atualiza sua posição via interface web utilizando um mapa interativo.

## 2. Descrição das Operações

1.  **Simulação de Movimento:** O backend executa uma rotina em segundo plano que calcula uma nova coordenada a cada intervalo de 4 a 6 segundos. O cálculo utiliza trigonometria esférica para deslocar o ponto em 10 metros em uma direção aleatória, garantindo uma simulação contínua de deslocamento.
2.  **Comunicação em Tempo Real (SSE):** Para a transmissão de dados, foi adotada a tecnologia **Server-Sent Events (SSE)**. Esta escolha permite que o servidor mantenha um canal de comunicação unidirecional aberto com o cliente, enviando a nova posição em formato JSON assim que ela é gerada, sem a necessidade de requisições repetitivas por parte do navegador (polling).
3.  **Interface de Visualização:** O frontend utiliza a biblioteca **Leaflet** integrada ao **OpenStreetMap** para a renderização cartográfica. Ao receber a nova coordenada via SSE, o navegador atualiza a posição do marcador no mapa e centraliza a visualização automaticamente, proporcionando uma experiência de rastreamento fluida.

## 3. Limites Potenciais e Melhorias

*   **Persistência de Dados:** Atualmente, o sistema armazena a posição apenas em memória volátil. Caso o servidor seja reiniciado, o histórico e o progresso do rastreamento são perdidos, retornando o objeto ao ponto de origem.
*   **Escalabilidade de Conexões:** O uso de SSE mantém uma conexão aberta por cliente. Em um cenário com um número muito elevado de usuários simultâneos, o gerenciamento dessas conexões pode exigir otimizações na infraestrutura do servidor.
*   **Precisão Geográfica:** A simulação utiliza um modelo de movimento aleatório que não considera obstáculos físicos ou malhas viárias. O objeto pode se deslocar sobre áreas intransitáveis por não estar integrado a APIs de roteamento real.
*   **Segurança e Autenticação:** Os canais de dados estão abertos para consulta pública, carecendo de camadas de autenticação ou controle de acesso para proteger as informações de localização.

## 4. Evolução do Projeto
Como etapa posterior, planeja-se transpor esta base de simulação para um cenário real: um sistema de rastreamento para motoboys de um restaurante local. A lógica de mensageria (SSE) e a estrutura de visualização em mapa serão reaproveitadas, substituindo a simulação matemática pela recepção de coordenadas GPS reais enviadas por dispositivos móveis dos entregadores.

