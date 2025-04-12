# 🏭 Produtor-Consumidor em Go

Implementação concorrente do clássico problema do Produtor-Consumidor com buffer limitado, demonstrando padrões avançados de sincronização em Go.

## 📚 Visão Geral

Este projeto simula um sistema com:
- Múltiplos **produtores** gerando itens
- Múltiplos **consumidores** processando itens
- **Buffer compartilhado** com capacidade limitada
- Mecanismos de **sincronização thread-safe**

## ✨ Features

✔️ Configuração flexível de produtores/consumidores  
✔️ Buffer circular eficiente  
✔️ Sincronização com `Mutex` e canais  
✔️ Sistema de logs detalhado para depuração  
✔️ Encerramento gracefull controlado por timeout  
✔️ Visualização do estado do buffer em tempo real  

## 🛠️ Como Executar

1. Clone o repositório:
```bash
git clone https://github.com/seu-usuario/produtor-consumidor-go.git
```
2. Executar o programa
```bash
cd produtor-consumidor-go
go run main.go
