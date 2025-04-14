package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

// Configuração via argumentos de linha de comando
var (
	PR        int  // número de produtores
	CN        int  // número de consumidores
	N         int  // tamanho do buffer
	runTime   int  // tempo de execução em segundos
	debugging bool // modo de depuração
)

var (
	buffer     []int
	indexIn    int
	indexOut   int
	mutex      sync.Mutex
	empty      chan struct{} // semáforo de posições livres
	full       chan struct{} // semáforo de posições ocupadas
	shutdown   = make(chan struct{})    // canal para encerramento
	wg         sync.WaitGroup
	operations int // contador de operações para debug
)

func main() {
	// Inicialização com valores padrão
	PR = 5
	CN = 1
	N = 5
	runTime = 30
	debugging = true

	// Configuração via argumentos de linha de comando
	if len(os.Args) > 1 {
		parseArgs()
	}

	// Verificação de argumentos válidos
	if PR <= 0 || CN <= 0 || N <= 0 {
		fmt.Println("[ERRO] Número de produtores, consumidores e tamanho do buffer devem ser maiores que zero")
		os.Exit(1)
	}

	// Inicialização das estruturas
	rand.Seed(time.Now().UnixNano())
	buffer = make([]int, N)
	empty = make(chan struct{}, N)  // semáforo com capacidade N
	full = make(chan struct{}, N)   // semáforo com capacidade N (corrigido)

	// Inicializando o buffer (todas posições vazias)
	for i := 0; i < N; i++ {
		empty <- struct{}{}
	}

	fmt.Println("[SISTEMA] Buffer inicializado!")
	fmt.Printf("[SISTEMA] Iniciando com %d produtores, %d consumidores e buffer de tamanho %d\n", PR, CN, N)
	fmt.Printf("[SISTEMA] Tempo de execução: %d segundos\n", runTime)

	// Capturar Ctrl+C para encerramento
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		select {
		case <-sigChan:
			fmt.Println("\n[SISTEMA] Sinal de interrupção recebido")
		case <-time.After(time.Duration(runTime) * time.Second):
			fmt.Printf("\n[SISTEMA] Tempo de execução (%ds) atingido\n", runTime)
		}
		close(shutdown)
	}()

	// Criando produtores
	for i := 0; i < PR; i++ {
		wg.Add(1)
		go produtor(i)
	}

	// Criando consumidores
	for i := 0; i < CN; i++ {
		wg.Add(1)
		go consumidor(i)
	}

	wg.Wait()
	fmt.Println("[SISTEMA] Programa encerrado")
}

// Função para analisar argumentos de linha de comando
func parseArgs() {
	// Formato: go run main.go [produtores] [consumidores] [tamanho_buffer] [tempo_execucao] [debug]
	if len(os.Args) > 1 {
		if p, err := strconv.Atoi(os.Args[1]); err == nil {
			PR = p
		}
	}
	if len(os.Args) > 2 {
		if c, err := strconv.Atoi(os.Args[2]); err == nil {
			CN = c
		}
	}
	if len(os.Args) > 3 {
		if n, err := strconv.Atoi(os.Args[3]); err == nil {
			N = n
		}
	}
	if len(os.Args) > 4 {
		if t, err := strconv.Atoi(os.Args[4]); err == nil {
			runTime = t
		}
	}
	if len(os.Args) > 5 {
		if d, err := strconv.ParseBool(os.Args[5]); err == nil {
			debugging = d
		}
	}
}

func log(format string, args ...interface{}) {
	if debugging {
		fmt.Printf(format, args...)
	}
}

func produtor(id int) {
	defer wg.Done()
	produzidos := 0

	for {
		select {
		case <-shutdown:
			fmt.Printf("[PRODUTOR %d] Encerrando após produzir %d itens\n", id, produzidos)
			return
		default:
			item := rand.Intn(100) + 1 // produz dado (1-100)
			
			// Log antes da inserção
			log("[PRODUTOR %d] Quer inserir %d. Posições livres: %d/%d\n", 
				id, item, len(empty), cap(empty))
			
			select {
			case <-empty: // aguarda posição livre
				mutex.Lock()
				// Debug: estado antes da inserção
				operations++
				opID := operations
				log("[OP %d PRODUTOR %d] Inserindo %d em buffer[%d]. Estado antes: %v\n",
					opID, id, item, indexIn, buffer)
				
				buffer[indexIn] = item
				indexIn = (indexIn + 1) % N
				produzidos++
				
				// Debug: estado após inserção
				log("[OP %d PRODUTOR %d] Inserção concluída. Estado após: %v\n",
					opID, id, buffer)
				mutex.Unlock()
				
				full <- struct{}{} // sinaliza posição ocupada
				
				// Verificação do buffer cheio
				if len(full) == cap(full) {
					fmt.Printf("[BUFFER] CHEIO! Posições ocupadas: %d/%d\n", len(full), cap(full))
				}
				
			case <-shutdown:
				fmt.Printf("[PRODUTOR %d] Encerrando após produzir %d itens\n", id, produzidos)
				return
			}
			
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		}
	}
}

func consumidor(id int) {
	defer wg.Done()
	consumidos := 0

	for {
		select {
		case <-shutdown:
			fmt.Printf("[CONSUMIDOR %d] Encerrando após consumir %d itens\n", id, consumidos)
			return
		default:
			// Verificação antes de tentar consumir
			if len(full) == 1 {
				fmt.Printf("[CONSUMIDOR %d] ATENÇÃO: Último item no buffer\n", id)
			}
			
			select {
			case <-full: // aguarda posição ocupada
				mutex.Lock()
				// Debug: estado antes da remoção
				operations++
				opID := operations
				
				log("[OP %d CONSUMIDOR %d] Removendo de buffer[%d]. Estado antes: %v\n",
					opID, id, indexOut, buffer)
				
				item := buffer[indexOut]
				buffer[indexOut] = 0 // limpa a posição para debug
				indexOut = (indexOut + 1) % N
				consumidos++
				
				// Debug: estado após remoção
				log("[OP %d CONSUMIDOR %d] Removeu %d. Estado após: %v\n",
					opID, id, item, buffer)
				
				mutex.Unlock()
				empty <- struct{}{} // libera posição
				
				// Verificação do buffer vazio
				if len(empty) == cap(empty) {
					fmt.Printf("[BUFFER] VAZIO! Posições ocupadas: %d/%d\n", len(full), cap(full))
				}
				
			case <-shutdown:
				fmt.Printf("[CONSUMIDOR %d] Encerrando após consumir %d itens\n", id, consumidos)
				return
			}
			
			time.Sleep(time.Duration(rand.Intn(1500)) * time.Millisecond)
		}
	}
}