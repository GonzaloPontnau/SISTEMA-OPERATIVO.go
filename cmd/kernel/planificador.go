package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/sisoputnfrba/tp-2025-1c-LosCuervosXeneizes/utils"
)

var (
	proximoPID int = 0 // PID arranca en 0 según enunciado
	pidMutex   sync.Mutex

	// Colas de estados
	colaNew         []*PCB              = []*PCB{}
	colaReady       []*PCB              = []*PCB{}
	colaExec        map[string]*PCB     = make(map[string]*PCB) // Mapa de [nombreCPU]*PCB
	colaBlocked     []*PCB              = []*PCB{}
	colaSuspReady   []*PCB              = []*PCB{}
	colaSuspBlocked []*PCB              = []*PCB{}
	colaExit        []*PCB              = []*PCB{}

	// Mutexes optimizados
	newMutex         sync.Mutex
	readyMutex       sync.Mutex
	execMutex        sync.Mutex
	blockedMutex     sync.Mutex
	suspReadyMutex   sync.Mutex
	suspBlockedMutex sync.Mutex
	exitMutex        sync.Mutex
	mapaMutex        sync.RWMutex

	// Conditions
	condNew       *sync.Cond
	condReady     *sync.Cond
	condSuspReady *sync.Cond

	mapaPCBs               map[int]*PCB = make(map[int]*PCB)
	gradoMultiprogramacion int
	semaforoMultiprogram   *utils.Semaforo
	timersSuspension       map[int]*time.Timer
	timersMutex            sync.Mutex
)

// InicializarPlanificador optimizado
func InicializarPlanificador(config *KernelConfig) {
	gradoMultiprogramacion = config.GradoMultiprogramacion
	if gradoMultiprogramacion <= 0 {
		gradoMultiprogramacion = 1
	}
	semaforoMultiprogram = utils.NewSemaforo(gradoMultiprogramacion)

	condNew = sync.NewCond(&newMutex)
	condReady = sync.NewCond(&readyMutex)
	condSuspReady = sync.NewCond(&suspReadyMutex)
	timersSuspension = make(map[int]*time.Timer)

	utils.InfoLog.Info("Planificador inicializado",
		"algoritmo_sts", config.SchedulerAlgorithm,
		"algoritmo_lts", config.ReadyIngressAlgorithm,
		"multiprogramacion", gradoMultiprogramacion)
}

// --- Funciones Generales de Gestión de Procesos ---

// GenerarNuevoPID devuelve un PID único
func GenerarNuevoPID() int {
	pidMutex.Lock()
	defer pidMutex.Unlock()
	pid := proximoPID
	proximoPID++
	return pid
}

// BuscarPCBPorPID busca un PCB en el mapa global
func BuscarPCBPorPID(pid int) *PCB {
	mapaMutex.RLock()
	defer mapaMutex.RUnlock()
	return mapaPCBs[pid]
}

// AgregarProcesoANew optimizado
func AgregarProcesoANew(pcb *PCB) {
	newMutex.Lock()
	colaNew = append(colaNew, pcb)
	newMutex.Unlock()
	condNew.Signal()
}

// MoverProcesoAReady optimizado
func MoverProcesoAReady(pcb *PCB) {
	removerDeBlocked(pcb)
	pcb.CambiarEstado(EstadoReady)

	readyMutex.Lock()
	colaReady = append(colaReady, pcb)
	readyMutex.Unlock()
	condReady.Signal()
}

// MoverProcesoABlocked optimizado
func MoverProcesoABlocked(pcb *PCB, motivo string) {
	execMutex.Lock()
	// Buscar y remover de ejecución si corresponde
	for cpu, pcbEnExec := range colaExec {
		if pcbEnExec != nil && pcbEnExec.PID == pcb.PID {
			delete(colaExec, cpu)
			break
		}
	}
	execMutex.Unlock()

	pcb.MotivoBloqueo = motivo
	pcb.CambiarEstado(EstadoBlocked)
	utils.InfoLog.Info(fmt.Sprintf("## (%d) - Bloqueado por IO: %s", pcb.PID, motivo))

	blockedMutex.Lock()
	colaBlocked = append(colaBlocked, pcb)
	blockedMutex.Unlock()

	go iniciarTimerSuspension(pcb)
}

// iniciarTimerSuspension como goroutine separada
func iniciarTimerSuspension(pcb *PCB) {
	tiempoSuspension := time.Duration(kernelConfig.SuspensionTime) * time.Millisecond
	if tiempoSuspension <= 0 {
		tiempoSuspension = 4500 * time.Millisecond
	}

	timersMutex.Lock()
	if timer, existe := timersSuspension[pcb.PID]; existe {
		timer.Stop()
	}

	timer := time.AfterFunc(tiempoSuspension, func() {
		suspenderProceso(pcb.PID)
	})
	timersSuspension[pcb.PID] = timer
	timersMutex.Unlock()
}

// suspenderProceso mueve un proceso de BLOCKED a SUSP. BLOCKED
func suspenderProceso(pid int) {
	pcb := BuscarPCBPorPID(pid)
	if pcb == nil || pcb.Estado != EstadoBlocked {
		return
	}

	if !removerDeBlocked(pcb) {
		return
	}

	pcb.CambiarEstado(EstadoSuspBlocked)

	suspBlockedMutex.Lock()
	colaSuspBlocked = append(colaSuspBlocked, pcb)
	suspBlockedMutex.Unlock()

	go notificarSwapAMemoria(pcb.PID)
	semaforoMultiprogram.Signal()
	condNew.Signal()
}

// FinalizarProceso optimizado
func FinalizarProceso(pcb *PCB, motivo string) {
	mapaMutex.Lock()
	if _, existe := mapaPCBs[pcb.PID]; !existe || pcb.Estado == EstadoExit {
		mapaMutex.Unlock()
		return
	}
	mapaMutex.Unlock()

	estadoPrevio := pcb.Estado

	// Limpiar timer
	timersMutex.Lock()
	if timer, existe := timersSuspension[pcb.PID]; existe {
		timer.Stop()
		delete(timersSuspension, pcb.PID)
	}
	timersMutex.Unlock()

	// Remover de cola actual

	fueRemovido := false
	_ = fueRemovido
	switch estadoPrevio {
	case EstadoExec:
		execMutex.Lock()
		for cpu, pcbEnExec := range colaExec {
			if pcbEnExec != nil && pcbEnExec.PID == pcb.PID {
				delete(colaExec, cpu)
				fueRemovido = true
				break
			}
		}
		execMutex.Unlock()
	case EstadoReady:
		fueRemovido = removerDeReady(pcb)
	case EstadoBlocked:
		fueRemovido = removerDeBlocked(pcb)
	case EstadoSuspReady:
		fueRemovido = removerDeSuspReady(pcb)
	case EstadoSuspBlocked:
		fueRemovido = removerDeSuspBlocked(pcb)
	case EstadoNew:
		fueRemovido = removerDeNew(pcb)
	case EstadoExit:
		return
	}

	pcb.CambiarEstado(EstadoExit)

	exitMutex.Lock()
	colaExit = append(colaExit, pcb)
	exitMutex.Unlock()

	// Liberar multiprogramación
	if estadoPrevio == EstadoReady || estadoPrevio == EstadoExec || estadoPrevio == EstadoBlocked {
		semaforoMultiprogram.Signal()
	}

	go notificarFinalizacionAMemoria(pcb.PID)

	if estadoPrevio != EstadoExit {
		utils.InfoLog.Info(fmt.Sprintf("## (%d) - Finaliza el proceso", pcb.PID))
		pcb.CalcularMetricas()
	}

	mapaMutex.Lock()
	delete(mapaPCBs, pcb.PID)
	mapaMutex.Unlock()
}

// notificarFinalizacionAMemoria simplificado
func notificarFinalizacionAMemoria(pid int) {
	datos := map[string]interface{}{
		"pid":       pid,
		"operacion": "FINALIZAR_PROCESO",
	}

	_, err := memoriaClient.EnviarHTTPOperacion("FINALIZAR_PROCESO", datos)
	if err != nil {
		utils.ErrorLog.Error("Error notificando finalización a Memoria", "pid", pid, "error", err.Error())
	}
}

// Funciones auxiliares optimizadas con templates
func removerDeReady(pcb *PCB) bool {
	readyMutex.Lock()
	defer readyMutex.Unlock()
	return removerDeCola(&colaReady, pcb)
}

func removerDeBlocked(pcb *PCB) bool {
	blockedMutex.Lock()
	defer blockedMutex.Unlock()
	return removerDeCola(&colaBlocked, pcb)
}

func removerDeNew(pcb *PCB) bool {
	newMutex.Lock()
	defer newMutex.Unlock()
	return removerDeCola(&colaNew, pcb)
}

func removerDeSuspReady(pcb *PCB) bool {
	suspReadyMutex.Lock()
	defer suspReadyMutex.Unlock()
	return removerDeCola(&colaSuspReady, pcb)
}

func removerDeSuspBlocked(pcb *PCB) bool {
	suspBlockedMutex.Lock()
	defer suspBlockedMutex.Unlock()
	return removerDeCola(&colaSuspBlocked, pcb)
}

// Template function para remover de cualquier cola
func removerDeCola(cola *[]*PCB, pcb *PCB) bool {
	for i, p := range *cola {
		if p.PID == pcb.PID {
			*cola = append((*cola)[:i], (*cola)[i+1:]...)
			return true
		}
	}
	return false
}

// Funciones de planificación optimizadas
func intentarAdmitirProceso() {
	condNew.Signal()
}

func despacharProcesoSiCorresponde() {
	condReady.Signal()
}

func notificarSwapAMemoria(pid int) {
	datos := map[string]interface{}{
		"pid":       pid,
		"operacion": "SUSPENDER_PROCESO",
	}
	memoriaClient.EnviarHTTPOperacion("SUSPENDER_PROCESO", datos)
}
