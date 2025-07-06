package main

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/sisoputnfrba/tp-2025-1c-LosCuervosXeneizes/utils"
)

// Variables globales para gestión de CPUs
var (
	cpuClientsMutex          sync.Mutex
	ultimoLogCPUNoDisponible time.Time // Para evitar spam de logs
)

// registrarCPU registra un nuevo cliente CPU
func registrarCPU(nombre string, ip string, puerto int) {
	cpuClientsMutex.Lock()
	defer cpuClientsMutex.Unlock()

	// DIAGNÓSTICO: Verificar estado inicial del mapa
	utils.InfoLog.Debug("Estado del mapa cpuClients ANTES del registro", "total_cpus", len(cpuClients), "keys", obtenerNombresCPUs())

	// Asegurar que el mapa está inicializado
	if cpuClients == nil {
		utils.InfoLog.Warn("Mapa cpuClients no inicializado, creándolo ahora")
		cpuClients = make(map[string]*utils.HTTPClient)
	}

	// Verificar si la CPU ya estaba registrada
	_, existe := cpuClients[nombre]
	if existe {
		utils.InfoLog.Info(fmt.Sprintf("CPU '%s' ya registrada, actualizando conexión", nombre))
	}

	// Registrar la CPU - Usar una clave más distintiva para evitar colisiones
	nombreCPU := nombre
	if nombreCPU == "" {
		nombreCPU = fmt.Sprintf("CPU_%s_%d", ip, puerto)
		utils.InfoLog.Info("Usando nombre generado para CPU", "nombre_generado", nombreCPU)
	}

	// Crear el cliente HTTP para esta CPU
	cpuClients[nombreCPU] = utils.NewHTTPClient(ip, puerto, "Kernel->"+nombreCPU)

	utils.InfoLog.Info(fmt.Sprintf("## CPU '%s' registrada correctamente en %s:%d", nombreCPU, ip, puerto),
		"total_cpus", len(cpuClients))

	// Mostrar todas las CPUs registradas actualmente
	utils.InfoLog.Debug("CPUs registradas DESPUÉS del registro:", "total", len(cpuClients), "nombres", obtenerNombresCPUs())
}

// obtenerNombresCPUs devuelve una lista con los nombres de todas las CPUs registradas
func obtenerNombresCPUs() []string {
	nombres := make([]string, 0, len(cpuClients))
	for nombre := range cpuClients {
		nombres = append(nombres, nombre)
	}
	return nombres
}

// obtenerCPUDisponible retorna un cliente de CPU disponible o nil si no hay ninguno
func obtenerCPUDisponible() (string, *utils.HTTPClient) {
	cpuClientsMutex.Lock()
	defer cpuClientsMutex.Unlock()

	// DIAGNÓSTICO: Información detallada sobre el estado del mapa
	cpuNombres := obtenerNombresCPUs()
	utils.InfoLog.Debug("Verificando CPUs disponibles", "total_registradas", len(cpuClients), "nombres", cpuNombres)

	// Verificar si el mapa está inicializado
	if cpuClients == nil {
		utils.InfoLog.Error("Mapa cpuClients no inicializado, esto no debería ocurrir")
		return "", nil
	}

	if len(cpuClients) == 0 {
		// Limitar los mensajes de log para evitar spam (máximo uno cada 5 segundos)
		ahora := time.Now()
		if ahora.Sub(ultimoLogCPUNoDisponible) > 5*time.Second {
			utils.InfoLog.Warn("No hay CPUs registradas en el sistema")
			ultimoLogCPUNoDisponible = ahora
		}
		return "", nil
	}

	// Por simplicidad, retornamos la primera CPU (en un sistema real, se implementaría un algoritmo de selección)
	for nombre, cliente := range cpuClients {
		utils.InfoLog.Info("Seleccionando CPU para ejecución", "nombre", nombre)
		return nombre, cliente
	}

	// Si por alguna razón no se encontró ninguna (lo cual no debería ocurrir si len > 0)
	utils.InfoLog.Error("Error lógico: len(cpuClients) > 0 pero no se encontró ninguna CPU")
	return "", nil
}

// PlanificarCortoPlazo gestiona transición de procesos entre READY y EXEC
func PlanificarCortoPlazo() {
	for {
		readyMutex.Lock()
		for len(colaReady) == 0 {
			condReady.Wait()
		}

		// Seleccionar próximo proceso según algoritmo
		pcb := seleccionarProcesoSTS()
		
		// Si SRT decide desalojar, pcb será nil. El planificador se re-ejecutará
		if pcb == nil {
			readyMutex.Unlock()
			time.Sleep(100 * time.Millisecond) // Pequeña pausa para que el desalojo se procese
			continue
		}

		readyMutex.Unlock() // Desbloquear la cola de ready una vez tenemos un candidato

		// Bucle para esperar una CPU disponible
		var nombreCPU string
		var cpuClient *utils.HTTPClient
		
		for {
			execMutex.Lock()
			nombreCPU, cpuClient = obtenerCPUDisponibleParaEjecucion()
			if cpuClient != nil {
				// CPU encontrada y reservada lógicamente
				colaExec[nombreCPU] = pcb 
				execMutex.Unlock()
				break
			}
			execMutex.Unlock()
			time.Sleep(200 * time.Millisecond) // Esperar antes de volver a verificar
		}

		// Remover proceso seleccionado de la cola READY (ahora que tenemos CPU)
		readyMutex.Lock()
		removerDeReady(pcb)
		readyMutex.Unlock()

		// Asignar a CPU
		pcb.CambiarEstado(EstadoExec)
		
		utils.InfoLog.Info(fmt.Sprintf("Proceso despachado a CPU %s", nombreCPU), "pid", pcb.PID)

		// Enviar a CPU (comunicación real)
		go enviarProcesoACPU(nombreCPU, cpuClient, pcb)
	}
}

// obtenerCPUDisponibleParaEjecucion busca una CPU que no esté en la cola de ejecución.
// ¡IMPORTANTE! Esta función debe ser llamada con el mutex de colaExec BLOQUEADO.
func obtenerCPUDisponibleParaEjecucion() (string, *utils.HTTPClient) {
	cpuClientsMutex.Lock()
	defer cpuClientsMutex.Unlock()

	for nombre, cliente := range cpuClients {
		if _, ocupada := colaExec[nombre]; !ocupada {
			return nombre, cliente
		}
	}
	return "", nil
}

// seleccionarProcesoSTS selecciona el próximo proceso de cola READY según algoritmo configurado
// ¡IMPORTANTE! Esta función debe ser llamada con el mutex de colaReady BLOQUEADO.
func seleccionarProcesoSTS() *PCB {
	if len(colaReady) == 0 {
		return nil
	}

	algoritmo := kernelConfig.SchedulerAlgorithm
	utils.InfoLog.Debug("STS seleccionando proceso", "algoritmo", algoritmo, "procesos_disponibles", len(colaReady))

	switch algoritmo {
	case "FIFO":
		return seleccionarFIFO()
	case "SJF":
		return seleccionarSJF()
	case "SRT":
		return seleccionarSRT()
	default:
		utils.InfoLog.Warn("Algoritmo STS no reconocido, usando FIFO", "algoritmo", algoritmo)
		return seleccionarFIFO()
	}
}

// seleccionarFIFO implementa selección FIFO
func seleccionarFIFO() *PCB {
	return colaReady[0]
}

// seleccionarSJF implementa Shortest Job First sin desalojo
func seleccionarSJF() *PCB {
	if len(colaReady) == 0 {
		return nil
	}

	// Crear copia para ordenar
	candidatos := make([]*PCB, len(colaReady))
	copy(candidatos, colaReady)

	// Ordenar por estimación de tiempo (menor = mayor prioridad)
	sort.Slice(candidatos, func(i, j int) bool {
		// Si tienen la misma estimación, usar FIFO
		if candidatos[i].EstimacionSiguienteRafaga == candidatos[j].EstimacionSiguienteRafaga {
			return candidatos[i].HoraListo.Before(candidatos[j].HoraListo)
		}
		return candidatos[i].EstimacionSiguienteRafaga < candidatos[j].EstimacionSiguienteRafaga
	})

	seleccionado := candidatos[0]
	utils.InfoLog.Debug("SJF seleccionó proceso",
		"pid", seleccionado.PID,
		"estimacion", seleccionado.EstimacionSiguienteRafaga)

	return seleccionado
}

// seleccionarSRT implementa Shortest Remaining Time (SJF con desalojo)
// ¡IMPORTANTE! Esta función debe ser llamada con el mutex de colaReady BLOQUEADO.
func seleccionarSRT() *PCB {
	if len(colaReady) == 0 {
		return nil
	}

	// Encontrar el proceso con menor tiempo estimado en READY
	mejorCandidatoReady := encontrarMejorCandidatoReady()

	// Comparar con los procesos en ejecución
	execMutex.Lock()
	procesoADesalojar := encontrarProcesoADesalojar(mejorCandidatoReady)
	execMutex.Unlock()

	if procesoADesalojar != nil {
		utils.InfoLog.Info(fmt.Sprintf("## (%d) - Desalojado por algoritmo SJF/SRT por proceso (%d)", procesoADesalojar.PID, mejorCandidatoReady.PID))
		go desalojarProcesoActual(procesoADesalojar)
		return nil
	}
	
	return mejorCandidatoReady
}

// ¡IMPORTANTE! Esta función debe ser llamada con el mutex de colaReady BLOQUEADO.
func encontrarMejorCandidatoReady() *PCB {
	if len(colaReady) == 0 {
		return nil
	}
	// Encontrar el proceso con menor tiempo estimado en READY
	mejorProceso := colaReady[0]
	for _, pcb := range colaReady[1:] {
		if pcb.EstimacionSiguienteRafaga < mejorProceso.EstimacionSiguienteRafaga {
			mejorProceso = pcb
		} else if pcb.EstimacionSiguienteRafaga == mejorProceso.EstimacionSiguienteRafaga {
			if pcb.HoraListo.Before(mejorProceso.HoraListo) {
				mejorProceso = pcb
			}
		}
	}
	return mejorProceso
}

// ¡IMPORTANTE! Esta función debe ser llamada con el mutex de execMutex BLOQUEADO.
func encontrarProcesoADesalojar(candidato *PCB) *PCB {
	var procesoMasLargo *PCB = nil
	if candidato == nil {
		return nil
	}

	for _, pcbEnExec := range colaExec {
		if candidato.EstimacionSiguienteRafaga < pcbEnExec.EstimacionSiguienteRafaga {
			if procesoMasLargo == nil || pcbEnExec.EstimacionSiguienteRafaga > procesoMasLargo.EstimacionSiguienteRafaga {
				procesoMasLargo = pcbEnExec
			}
		}
	}
	return procesoMasLargo
}

// desalojarProcesoActual maneja el desalojo de un proceso por SRT
func desalojarProcesoActual(pcb *PCB) {
	var cpuADesalojar string
	execMutex.Lock()
	// Verificar que sigue siendo el proceso en ejecución y encontrar su CPU
	for cpu, pcbEnExec := range colaExec {
		if pcbEnExec != nil && pcbEnExec.PID == pcb.PID {
			cpuADesalojar = cpu
			break
		}
	}
	execMutex.Unlock()

	if cpuADesalojar == "" {
		utils.InfoLog.Warn("SRT: Intento de desalojar proceso que ya no está en ejecución", "pid", pcb.PID)
		return
	}

	cpuClientsMutex.Lock()
	cpuClient, existe := cpuClients[cpuADesalojar]
	cpuClientsMutex.Unlock()

	if !existe {
		utils.ErrorLog.Error("SRT: No se encontró el cliente para la CPU a desalojar", "cpu", cpuADesalojar)
		return
	}

	// Enviar interrupción a la CPU
	utils.InfoLog.Info("SRT: Enviando interrupción a CPU", "cpu", cpuADesalojar, "pid", pcb.PID)
	_, err := cpuClient.EnviarHTTPOperacion("INTERRUPT", nil) // No necesita payload
	if err != nil {
		utils.ErrorLog.Error("SRT: Fallo al enviar interrupción a la CPU", "cpu", cpuADesalojar, "error", err)
		// Aquí podría haber una lógica de reintento o manejo de fallo de CPU
	}
}

// enviarProcesoACPU envía realmente un proceso a la CPU mediante API
func enviarProcesoACPU(nombreCPU string, cpuClient *utils.HTTPClient, pcb *PCB) {
	utils.InfoLog.Info(fmt.Sprintf("Enviando proceso a CPU %s", nombreCPU), "pid", pcb.PID, "pc", pcb.PC)

	// Preparar datos para CPU
	datos := map[string]interface{}{
		"pid":       pcb.PID,
		"pc":        pcb.PC,
		"operacion": "EJECUTAR_PROCESO",
	}

	// Enviar solicitud a CPU y no esperar respuesta (se manejará en el handler)
	_, err := cpuClient.EnviarHTTPOperacion("EJECUTAR_PROCESO", datos)
	if err != nil {
		utils.ErrorLog.Error(fmt.Sprintf("Error al enviar proceso a CPU %s", nombreCPU),
			"pid", pcb.PID,
			"error", err.Error())

		// Si falla el envío, el proceso debe volver a READY y la CPU liberarse
		execMutex.Lock()
		delete(colaExec, nombreCPU)
		execMutex.Unlock()

		MoverProcesoAReady(pcb)
	}
}

// EnviarProcesoCPU envía un PCB a la CPU para su ejecución
func EnviarProcesoCPU(pcb *PCB, nombreCPU string) bool {
	// Obtener cliente para esta CPU
	cpuClient, existe := cpuClients[nombreCPU]
	if !existe {
		utils.ErrorLog.Error("CPU no encontrada", "cpu_id", nombreCPU)
		return false
	}

	// Preparar datos para enviar a la CPU
	datos := map[string]interface{}{
		"pid": pcb.PID,
		"pc":  pcb.PC,
	}

	// Log del envío
	utils.InfoLog.Info(fmt.Sprintf("## (%d) - Enviando proceso a CPU", pcb.PID),
		"pc", pcb.PC, "cpu", nombreCPU)

	// Enviar solicitud a la CPU
	respuesta, err := cpuClient.EnviarHTTPOperacion("EJECUTAR", datos)
	if err != nil {
		utils.ErrorLog.Error("Error enviando proceso a CPU",
			"pid", pcb.PID,
			"error", err.Error())
		return false
	}

	// Procesar respuesta
	if respuestaMap, ok := respuesta.(map[string]interface{}); ok {
		// Verificar si hay un error específico
		if errorMsg, tieneError := respuestaMap["error"].(string); tieneError {
			utils.ErrorLog.Error("Error reportado por CPU",
				"pid", pcb.PID,
				"mensaje", errorMsg)
			return false
		}

		// Actualizaciones según la respuesta
		utils.InfoLog.Debug("Respuesta de CPU recibida", "respuesta", fmt.Sprintf("%+v", respuestaMap))

		// Actualizar PC
		if nuevoPC, hayPC := respuestaMap["pc"].(float64); hayPC {
			pcb.PC = int(nuevoPC)
			utils.InfoLog.Debug("PC actualizado desde respuesta CPU", "pid", pcb.PID, "nuevo_pc", pcb.PC)
		}

		// Verificar si es una llamada al sistema (syscall)
		if esSyscall, ok := respuestaMap["syscall"].(bool); ok && esSyscall {
			utils.InfoLog.Info(fmt.Sprintf("## (%d) - Syscall detectada", pcb.PID))

			// Actualizar estado del proceso
			pcb.CambiarEstado(EstadoBlocked)

			// Obtener parámetros de la syscall
			if parametros, ok := respuestaMap["parametros"].(map[string]interface{}); ok {
				// Procesar syscall según su tipo
				if tipo, hayTipo := parametros["tipo"].(string); hayTipo {
					utils.InfoLog.Info(fmt.Sprintf("## (%d) - Procesando syscall", pcb.PID), "tipo", tipo)

					// En lugar de usar un manejador dinámico, simplemente finalizamos el proceso
					// con un mensaje que indica que se recibió una syscall
					utils.InfoLog.Info(fmt.Sprintf("## (%d) - Recibida syscall %s", pcb.PID, tipo))
					FinalizarProceso(pcb, fmt.Sprintf("SYSCALL_%s", tipo))
				}
			}
			return true
		}

		// Si no es una syscall, considerar el proceso como completado normalmente
		return true
	}

	// Si llegamos aquí, hubo un error con la respuesta
	utils.ErrorLog.Warn("Formato de respuesta inválido de CPU", "respuesta", fmt.Sprintf("%v", respuesta))
	return false
}
