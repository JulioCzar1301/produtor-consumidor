package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	PR = 5 // número de produtores
	CN = 1 // número de consumidores
	N  = 5 // tamanho do buffer
)

var (
	buffer     [N]int
	indexIn    int
	indexOut   int
	mutex      sync.Mutex
	empty      = make(chan struct{}, N) // semáforo de posições livres
	full       = make(chan struct{}, 0) // semáforo de posições ocupadas
	shutdown   = make(chan struct{})    // canal para encerramento
	wg         sync.WaitGroup
	operations int // contador de operações para debug
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// Inicializando o buffer
	for i := 0; i < N; i++ {
		empty <- struct{}{}
	}

	fmt.Println("[SISTEMA] Buffer inicializado!")
	fmt.Printf("[SISTEMA] Iniciando com %d produtores e %d consumidor\n", PR, CN)

	// Capturar Ctrl+C para encerramento
	go func() {
		time.Sleep(30 * time.Second) // Encerra após 30 segundos para teste
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

func produtor(id int) {
	defer wg.Done()
	for {
		select {
		case <-shutdown:
			fmt.Printf("[PRODUTOR %d] Encerrando\n", id)
			return
		default:
			item := rand.Intn(100) + 1 // produz dado (1-100)

			// Log antes da inserção
			fmt.Printf("[PRODUTOR %d] Quer inserir %d. Posições livres: %d/%d\n",
				id, item, len(empty), N)

			<-empty // aguarda posição livre
			mutex.Lock()

			// Debug: estado antes da inserção
			operations++
			opID := operations
			fmt.Printf("[OP %d PRODUTOR %d] Inserindo %d em buffer[%d]. Estado antes: %v\n",
				opID, id, item, indexIn, buffer)

			buffer[indexIn] = item
			indexIn = (indexIn + 1) % N

			// Debug: estado após inserção
			fmt.Printf("[OP %d PRODUTOR %d] Inserção concluída. Estado após: %v\n",
				opID, id, buffer)

			mutex.Unlock()
			full <- struct{}{} // sinaliza posição ocupada

			// Verificação precisa do buffer cheio
			mutex.Lock()
			if (indexIn == indexOut) && (len(full) == N) {
				fmt.Printf("[BUFFER] CHEIO! Posições ocupadas: %d/%d\n", len(full), N)
			}
			mutex.Unlock()

			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		}
	}
}

func consumidor(id int) {
	defer wg.Done()
	for {
		select {
		case <-shutdown:
			fmt.Printf("[CONSUMIDOR %d] Encerrando\n", id)
			return
		default:
			// Verificação antes de remover
			mutex.Lock()
			if len(full) == 1 {
				fmt.Printf("[CONSUMIDOR %d] ATENÇÃO: Último item no buffer\n", id)
			}
			mutex.Unlock()

			<-full // aguarda posição ocupada
			mutex.Lock()

			// Debug: estado antes da remoção
			operations++
			opID := operations
			fmt.Printf("[OP %d CONSUMIDOR %d] Removendo de buffer[%d]. Estado antes: %v\n",
				opID, id, indexOut, buffer)

			item := buffer[indexOut]
			buffer[indexOut] = 0 // limpa a posição para debug
			indexOut = (indexOut + 1) % N

			// Debug: estado após remoção
			fmt.Printf("[OP %d CONSUMIDOR %d] Removeu %d. Estado após: %v\n",
				opID, id, item, buffer)

			mutex.Unlock()
			empty <- struct{}{} // libera posição

			// Verificação precisa do buffer vazio
			mutex.Lock()
			if indexIn == indexOut && len(full) == 0 {
				fmt.Printf("[BUFFER] VAZIO! Posições ocupadas: %d/%d\n", len(full), N)
			}
			mutex.Unlock()

			time.Sleep(time.Duration(rand.Intn(1500)) * time.Millisecond)
		}
	}
}
